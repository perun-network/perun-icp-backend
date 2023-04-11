// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"perun.network/perun-icp-backend/channel"
	"perun.network/perun-icp-backend/setup"
	"perun.network/perun-icp-backend/utils"
	"perun.network/perun-icp-backend/wallet"
	"regexp"
	"testing"
	"time"
)

const (
	perunID            = "r7inp-6aaaa-aaaaa-aaabq-cai"
	ledgerID           = "rrkah-fqaaa-aaaaa-aaaaq-cai"
	userAFundingAmount = 40000
	userBFundingAmount = 60000
	noFee              = 0
)

func TestPerunDeposit(t *testing.T) {

	testConfig := setup.DfxTestParams

	recipPerunID, err := utils.DecodePrincipal(perunID)
	if err != nil {
		t.Errorf("DecodePrincipal() error: %v", err)
	}

	recipientPerun := channel.Recipient{ID: recipPerunID}
	var input [32]byte
	recipPerunAcc := recipientPerun.ID.AccountIdentifier(input)

	// start DFX and deploy the Ledger and the Perun canisters
	dfx, err := setup.StartDeployDfx()
	if err != nil {
		log.Fatalf("StartDfxWithConfig() error: %v", err)
	}

	log.Println("DFX started, Ledger and Perun canisters have been deployed.")

	// initalize Perun users
	userClientA, err := channel.NewUserClient(testConfig)
	if err != nil {
		log.Fatalf("Error creating user A client: %v", err)
	}

	userClientB, err := channel.NewUserClient(testConfig)
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

	for i, user := range users {
		err := handleUserFunding(t, user, fundingList[i], txArgsList[i], ledgerID, testConfig.ExecPath, recipientPerun, i)
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

func handleUserFunding(t *testing.T, user *channel.UserClient, funding *channel.Funding, txArgs channel.TxArgs, ledgerID, execPath string, recipientPerun channel.Recipient, i int) error {
	// Retrieve memo from funding.
	// Memo hashes the funding struct, providing a unique identification of transfers targeted to fund Perun channels.
	memoFunding, err := funding.Memo()
	if err != nil {
		t.Errorf("Error for retrieving the memo: %v", err)
		return err
	}

	// Set memo in transaction arguments. This user-specific identifier identifies the funds tied to a user to the Perun channel.
	txArgs.Memo = memoFunding

	// Execute DFX transfer to the ICP Ledger using the Memo to make the transfer identifiable
	blockValue, err := channel.ExecuteDFXTransfer(txArgs, ledgerID, execPath)
	if err != nil {
		log.Printf("Error querying blocks: %v", err)
		return err
	}

	userLabel := "A"
	if i == 1 {
		userLabel = "B"
	}
	log.Printf("Transfer of user %s to the Perun ICP Ledger has been validated in block %d.", userLabel, blockValue)

	// Notify Perun about the transfer by looking for the specific memoFunding in the block in which the above transfer has been issued
	fundedValue, err := channel.NotifyTransferToPerun(user, blockValue, recipientPerun)
	if err != nil {
		t.Errorf("Error during notification of the funded value: %v", err)
		return err
	}
	require.Equal(t, txArgs.Amount, uint64(fundedValue), "The value we transfer should be the same value we get the transaction notification for the user")
	log.Printf("The Perun canister received %d tokens from user %s.", fundedValue, userLabel)

	// Deposit to Perun channel by draining the funds from the ICP ledger identified by the unique, channel- and user-specific memoFunding
	depositResult, err := channel.DepositToPerunChannel(user, funding, memoFunding, perunID, execPath)
	if err != nil {
		t.Errorf("Error for handling the deposit: %v", err)
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

	qsArgs := channel.DepositArgs{
		ChannelId:   funding.ChannelId.ID,
		Participant: user.L2Account.ICPAddress(),
		Memo:        memoFunding,
	}

	fmt.Println("Query state args: ", qsArgs)

	// err = channel.QueryStateCLI(qsArgs, perunID, execPath)
	// if err != nil {
	// 	t.Errorf("Error for querying the state: %v", err)
	// 	return err
	// }

	// err = channel.QueryStateCLI(channelValue, perunID, execPath)
	// if err != nil {
	// 	t.Errorf("Error for querying channel events: %v", err)
	// 	return err
	// }

	err = channel.QueryCandidCLI("()", perunID, execPath)
	if err != nil {
		t.Errorf("Error for querying candid: %v", err)
		return err

	}

	// Ensure that we have deposited the exact amount to the Perun channel which we planned to do in the txArgs field.
	require.Equal(t, txArgs.Amount, uint64(depositResult.ChannelAlloc), "The number of tokens available in the funded channel should be equal to the amount the user has initially transferred to the Perun Ledger address.")
	log.Printf("The Perun canister has deposited %d tokens from user %s into the channel %x: ", depositResult.ChannelAlloc, userLabel, funding.ChannelId)

	return nil
}
