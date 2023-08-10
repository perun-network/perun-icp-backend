package client_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"testing"
	"time"
)

func TestPrincipalTransfers(t *testing.T) {
	s := SimpleTxSetup(t)

	for i := 0; i < len(s.L1Users); i++ {
		bal, err := s.L1Users[i].GetBalance()
		require.NoError(t, err, "Failed to get balance")

		log.Println("Balance before sending: ", *bal)
	}

	// balances to transfer

	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	txBalances := make([]uint64, len(s.L1Users))
	for i := range txBalances {
		txBalances[i] = uint64(rand.Intn(4001) + 1000) // Assign a random uint64 value between 1000 and 5000
	}

	ledgerPrincipal := "bkyz2-fmaaa-aaaaa-qaaaq-cai"
	perunPrincipal := "be2us-64aaa-aaaaa-qaabq-cai"

	perunID, err := principal.Decode(perunPrincipal)
	require.NoError(t, err, "Failed to decode principal")
	ledgerID, err := principal.Decode(ledgerPrincipal)
	require.NoError(t, err, "Failed to decode principal")

	perunaccountID := perunID.AccountIdentifier(principal.DefaultSubAccount)
	txArgsList := make([]icpledger.TransferArgs, len(s.L1Users))

	for i := 0; i < len(s.L1Users); i++ {
		//fromSubaccount := s.L1Users[i].Prince.AccountIdentifier(principal.DefaultSubAccount).Bytes()
		toAccount := perunaccountID.Bytes()
		txArgsList[i] = icpledger.TransferArgs{
			Memo: uint64(i),
			Amount: struct {
				E8s uint64 "ic:\"e8s\""
			}{E8s: txBalances[i]},
			Fee: struct {
				E8s uint64 "ic:\"e8s\""
			}{E8s: chanconn.DfxTransferFee},
			//FromSubaccount: &fromSubaccount,
			To: toAccount,
		}

		_, err := s.L1Users[i].TransferDfx(txArgsList[i], ledgerID)
		require.NoError(t, err, "Failed to transfer")
		_, err = s.L1Users[i].GetBalance()
		require.NoError(t, err, "Failed to get balance")

	}
	_, err = s.PerunNode.GetBalance()
	require.NoError(t, err, "Failed to get balance")
}

func NewL1User(prince *principal.Principal, c *chanconn.Connector) *L1User {
	return &L1User{prince, c}
}

type L1Setup struct {
	T           *testing.T
	Accs        []*principal.Principal
	MinterAcc   *principal.Principal
	PerunPrince *principal.Principal
	Conns       []*chanconn.Connector
	ConnPerun   *chanconn.Connector
}

type OnChainSetup struct {
	*L1Setup
	L1Users   []*L1User
	PerunNode *L1User
}

type OnChainBareSetup struct {
	*L1Setup
	L1Users   []*L1User
	PerunNode *L1User
}

type L1User struct {
	Prince *principal.Principal
	Conn   *chanconn.Connector
}

func (u *L1User) GetBalance() (*uint64, error) {

	accountID := u.Prince.AccountIdentifier(principal.DefaultSubAccount)

	ledgerAgent := u.Conn.LedgerAgent
	onChainBal, err := ledgerAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: accountID.Bytes()})
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	return &onChainBal.E8s, nil
}

func (u *L1User) TransferDfx(txArgs icpledger.TransferArgs, canID principal.Principal) (uint64, error) {
	ldg := u.Conn.LedgerAgent

	transferResult, err := ldg.Transfer(txArgs)
	if err != nil {
		return 0, fmt.Errorf("dfx transfer command in TransferDfx failed: %v", err)
	}

	if transferResult.Err != nil {
		switch {
		case transferResult.Err.BadFee != nil:
			fmt.Printf("Transfer failed due to bad fee: expected fee: %v\n", transferResult.Err.BadFee.ExpectedFee)
		case transferResult.Err.InsufficientFunds != nil:
			fmt.Printf("Transfer failed due to insufficient funds: current balance: %v\n", transferResult.Err.InsufficientFunds.Balance)
		case transferResult.Err.TxTooOld != nil:
			fmt.Printf("Transfer failed because the transaction is too old. Allowed window (in nanos): %v\n", transferResult.Err.TxTooOld.AllowedWindowNanos)
		case transferResult.Err.TxCreatedInFuture != nil:
			fmt.Println("Transfer failed because the transaction was created in the future.")
		case transferResult.Err.TxDuplicate != nil:
			fmt.Printf("Transfer failed because it's a duplicate of transaction at block index: %v\n", transferResult.Err.TxDuplicate.DuplicateOf)
		default:
			fmt.Println("Transfer failed due to unknown reasons.")
		}
		return 0, fmt.Errorf("transfer failed with error: %v", transferResult.Err)
	}

	blnm := transferResult.Ok
	if blnm == nil {
		return 0, fmt.Errorf("blockNum is nil")
	}

	return *blnm, nil
}

func SimpleTxSetup(t *testing.T) *OnChainBareSetup {

	s := TransferSetup(t)
	c := s.Conns
	cp := s.ConnPerun
	pP := s.PerunPrince

	ret := &OnChainBareSetup{L1Setup: s}

	for i := 0; i < len(s.Accs); i++ {
		dep := NewL1User(s.Accs[i], c[i])
		ret.L1Users = append(ret.L1Users, dep)
	}
	pnode := NewL1User(pP, cp)

	ret.PerunNode = pnode
	return ret
}

func TransferSetup(t *testing.T) *L1Setup {

	Host := "http://127.0.0.1"
	Port := 4943

	aliceAccPath := "./../userdata/identities/usera_identity.pem"
	bobAccPath := "./../userdata/identities/userb_identity.pem"
	minterAccPath := "./../userdata/identities/minter_identity.pem"

	aliceAcc, err := chanconn.NewIdentity(aliceAccPath)
	if err != nil {
		panic(err)
	}
	bobAcc, err := chanconn.NewIdentity(bobAccPath)
	if err != nil {
		panic(err)
	}

	minterAcc, err := chanconn.NewIdentity(minterAccPath)
	if err != nil {
		panic(err)
	}

	alicePrince := (*aliceAcc).Sender()
	bobPrince := (*bobAcc).Sender()
	minterPrince := (*minterAcc).Sender()

	perunID := "be2us-64aaa-aaaaa-qaabq-cai"
	ledgerID := "bkyz2-fmaaa-aaaaa-qaaaq-cai"

	perunPrince, err := principal.Decode(perunID)
	if err != nil {
		panic(err)
	}

	accs := []*principal.Principal{&alicePrince, &bobPrince}
	conn1 := chanconn.NewDfxConnector(perunID, ledgerID, aliceAccPath, Host, Port)
	conn2 := chanconn.NewDfxConnector(perunID, ledgerID, bobAccPath, Host, Port)
	connPerun := chanconn.NewDfxConnector(perunID, ledgerID, minterAccPath, Host, Port)

	conns := []*chanconn.Connector{conn1, conn2}
	return &L1Setup{t, accs, &minterPrince, &perunPrince, conns, connPerun}
}
