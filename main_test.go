// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"testing"
	"time"

	_ "github.com/aviate-labs/agent-go/ledger"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/stretchr/testify/require"
	"perun.network/perun-icp-backend/channel"
	"perun.network/perun-icp-backend/setup"
	"perun.network/perun-icp-backend/utils"
	"perun.network/perun-icp-backend/wallet"
)

const (
	perunID                   = "r7inp-6aaaa-aaaaa-aaabq-cai"
	ledgerID                  = "rrkah-fqaaa-aaaaa-aaaaq-cai"
	userAFundingAmount        = 400000
	userBFundingAmount uint64 = 6000
	noFee                     = 0
)

func TestPerunDeposit(t *testing.T) {

	fmt.Println("Starting test: TestPerunDeposit()")

	testConfig := setup.DfxTestParams

	recipPerunID, err := utils.DecodePrincipal(perunID)
	if err != nil {
		t.Errorf("DecodePrincipal() error: %v", err)
	}

	recipLedgerID, err := utils.DecodePrincipal(ledgerID)
	if err != nil {
		t.Errorf("DecodePrincipal() error: %v", err)
	}

	recipientPerun := channel.Recipient{ID: recipPerunID}
	recipPerunAcc := recipientPerun.ID.AccountIdentifier(principal.DefaultSubAccount)

	// start DFX and deploy the Ledger and the Perun canisters
	dfx, err := setup.StartDeployDfx()
	if err != nil {
		log.Fatalf("StartDfxWithConfig() error: %v", err)
	}

	log.Println("DFX started, Ledger and Perun canisters have been deployed.")

	// initalize Perun users
	userClientA, err := channel.NewUserClient(testConfig, recipLedgerID)
	if err != nil {
		log.Fatalf("Error creating user A client: %v", err)
	}

	userClientB, err := channel.NewUserClient(testConfig, recipPerunID)
	if err != nil {
		log.Fatalf("Error creating user B client: %v", err)
	}

	users := []*channel.UserClient{
		userClientA,
		userClientB,
	}

	// Set up rng to generate a channelID, the unique Perun channel identifier
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	params := channel.Params{
		Nonce:             channel.NonceHash(rng),                                                             // random hash to calculate a unique channel ID
		Parts:             []wallet.Address{users[0].L2Account.ICPAddress(), users[1].L2Account.ICPAddress()}, // two participants
		ChallengeDuration: 3,                                                                                  // dispute challenge duration in seconds
	}

	pid, err := params.ParamsIDCandid()
	if err != nil {
		fmt.Println("error for paramsID: ", err)
	}

	channelID := channel.ChannelID{ID: pid}

	// Funding identifies uniquely Perun channels and their participants
	fundingList := []*channel.Funding{
		{ChannelId: channelID, L2Address: params.Parts[0]},
		{ChannelId: channelID, L2Address: params.Parts[1]},
	}

	// Tx arguments to transfer funds to the channel

	//subAccDefault := ledger.SubAccount(principal.DefaultSubAccount)
	txArgsList := []channel.TxArgs{
		{
			Amount: userAFundingAmount,
			Fee:    noFee,
			To:     recipPerunAcc.String(),
		},
		{
			Amount: userBFundingAmount,
			Fee:    noFee,
			To:     recipPerunAcc.String(),
		},
	}
	// We wil use this definition as soon as the error in the TransferDfx function is understood
	// txArgsList := []ledger.TransferArgs{
	// 	{
	// 		Memo:           0,
	// 		Amount:         ledger.Tokens{E8S: userAFundingAmount},
	// 		Fee:            ledger.Tokens{E8S: noFee},
	// 		FromSubAccount: &subAccDefault,
	// 		To:             recipPerunAcc,
	// 		CreatedAtTime: &ledger.TimeStamp{
	// 			TimestampNanos: uint64(time.Now().UnixNano()),
	// 		},
	// 	},
	// 	{
	// 		Memo:           0,
	// 		Amount:         ledger.Tokens{E8S: userBFundingAmount},
	// 		Fee:            ledger.Tokens{E8S: noFee},
	// 		FromSubAccount: &subAccDefault,
	// 		To:             recipPerunAcc,
	// 		CreatedAtTime: &ledger.TimeStamp{
	// 			TimestampNanos: uint64(time.Now().UnixNano()),
	// 		},
	// 	},
	// }

	for i, user := range users {

		params := HandleUserFundingParams{
			T:              t,
			User:           user,
			Funding:        fundingList[i],
			TxArgs:         txArgsList[i],
			LedgerID:       ledgerID,
			ExecPath:       testConfig.ExecPath,
			RecipientPerun: recipientPerun,
			I:              i,
		}
		err := handleUserFunding(params)
		if err != nil {
			t.Fatalf("handleUserFunding failed: %v", err)
		}
	}

	// exit the DFX replica
	err = setup.StopDFX(dfx)
	if err != nil {
		t.Fatalf("StopDFX() error: %v", err)
	}

	log.Println("DFX has been shutdown.")

}

type HandleUserFundingParams struct {
	T              *testing.T
	User           *channel.UserClient
	Funding        *channel.Funding
	TxArgs         channel.TxArgs //ledger.TransferArgs
	LedgerID       string
	ExecPath       string
	RecipientPerun channel.Recipient
	I              int
}

func handleUserFunding(params HandleUserFundingParams) error {
	memoFunding, err := params.Funding.Memo()

	if err != nil {
		params.T.Errorf("Error for retrieving the memo: %v", err)
		return err
	}
	tstamp := 0 //time.Now().UnixNano()

	// Set memo in transaction arguments. This user-specific identifier identifies the funds tied to a user to the Perun channel.
	params.TxArgs.Memo = memoFunding

	// Execute DFX transfer to the ICP Ledger using the Memo to make the transfer identifiable
	//blockValue, err := channel.ExecuteDFXTransfer(params.User, params.TxArgs, ledgerID)
	blockValue, err := channel.ExecuteDFXTransfer(params.TxArgs, params.LedgerID, params.ExecPath)

	if err != nil {
		log.Printf("Error querying blocks: %v", err)
		return err
	}

	userLabel := "A"
	if params.I == 1 {
		userLabel = "B"
	}
	log.Printf("Transfer of user %s to the Perun ICP Ledger has been validated in block %d.", userLabel, blockValue)

	// Notify Perun about the transfer by looking for the specific memoFunding in the block in which the above transfer has been issued
	fundedValue, err := channel.NotifyTransferToPerun(params.User, blockValue, params.RecipientPerun)
	if err != nil {
		params.T.Errorf("Error during notification of the funded value: %v", err)
		return err
	}
	require.Equal(params.T, params.TxArgs.Amount, uint64(fundedValue), "The value we transfer should be the same value we get the transaction notification for the user")
	log.Printf("The Perun canister received %d tokens from user %s.", fundedValue, userLabel)

	// Deposit to Perun channel by draining the funds from the ICP ledger identified by the unique, channel- and user-specific memoFunding
	depositResult, err := channel.DepositToPerunChannel(params.User, params.Funding, memoFunding, perunID, params.ExecPath)
	if err != nil {
		params.T.Errorf("Error for handling the deposit: %v", err)
		return err
	}

	fmt.Println("Deposit result: outputfundmemo, ", depositResult.OutputFundMemo, depositResult.FundingOutput, depositResult.ChannelAlloc)

	r := regexp.MustCompile(`channel\s+=\s+blob\s+"(.*?)"`)

	channelValue := ""
	matches := r.FindStringSubmatch(depositResult.OutputFundMemo)
	if len(matches) > 1 {
		channelValue = matches[1]
		channelValue = "(blob \"" + channelValue + "\")"
		fmt.Println(channelValue)
	} else {
		err := errors.New("channel value not found")
		fmt.Println(err.Error())
	}

	fmt.Println("channelValue: ", channelValue)

	// qsArgs := channel.DepositArgs{
	// 	ChannelId:   params.Funding.ChannelId.ID,
	// 	Participant: params.User.L2Account.ICPAddress(),
	// 	Memo:        memoFunding,
	// }
	//fmt.Println("Query state args: ", qsArgs)

	// err = channel.QueryStateCLI(qsArgs, perunID, params.ExecPath)
	// if err != nil {
	// 	params.T.Errorf("Error for querying the state: %v", err)
	// 	return err
	// }

	err = channel.QueryCandidCLI("()", perunID, params.ExecPath)
	if err != nil {
		params.T.Errorf("Error for querying candid: %v", err)
		return err

	}

	err = channel.QueryStateCLI(channelValue, perunID, params.ExecPath)
	if err != nil {
		params.T.Errorf("Error for querying channel events: %v", err)
		return err
	}

	fmt.Println("Timestamp: ", tstamp)
	qEventsArgs := channel.FormatChanTimeArgs(params.Funding.ChannelId.ID, uint64(tstamp))

	fmt.Println("Query events args: ", qEventsArgs)
	eventsOut, err := channel.QueryEventsCLI(qEventsArgs, perunID, params.ExecPath)
	if err != nil {
		params.T.Errorf("Error for querying channel events: %v", err)
		return err
	}

	_ = channel.StringIntoEvents(eventsOut)
	//channel.EmitLastEvents(eventsOut)

	// Ensure that we have deposited the exact amount to the Perun channel which we planned to do in the txArgs field.
	require.Equal(params.T, params.TxArgs.Amount, uint64(depositResult.ChannelAlloc), "The number of tokens available in the funded channel should be equal to the amount the user has initially transferred to the Perun Ledger address.")
	log.Printf("The Perun canister has deposited %d tokens from user %s into the channel %x: ", depositResult.ChannelAlloc, userLabel, params.Funding.ChannelId)

	return nil
}
