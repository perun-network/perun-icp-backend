package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
)

// HandleProposal is the callback for incoming channel proposals.
func (c *PaymentClient) HandleProposal(p client.ChannelProposal, r *client.ProposalResponder) {
	lcp, err := func() (*client.LedgerChannelProposalMsg, error) {
		// Ensure that we got a ledger channel proposal.
		lcp, ok := p.(*client.LedgerChannelProposalMsg)
		if !ok {
			return nil, fmt.Errorf("Invalid proposal type: %T\n", p)
		}

		// Check that we have the correct number of participants.
		if lcp.NumPeers() != 2 {
			return nil, fmt.Errorf("Invalid number of participants: %d", lcp.NumPeers())
		}

		// Check that the channel has the expected assets and funding balances.
		const assetIdx, clientIdx, peerIdx = 0, 0, 1
		if err := channel.AssertAssetsEqual(lcp.InitBals.Assets, []channel.Asset{c.currency}); err != nil {
			return nil, fmt.Errorf("Invalid assets: %v\n", err)
		} else if lcp.FundingAgreement[assetIdx][clientIdx].Cmp(lcp.FundingAgreement[assetIdx][peerIdx]) != 0 {
			return nil, fmt.Errorf("Invalid funding balance")
		}
		return lcp, nil
	}()
	if err != nil {
		errReject := r.Reject(context.TODO(), err.Error())
		if errReject != nil {
			// Log the error or take other action as needed
			fmt.Printf("Error rejecting proposal: %v\n", errReject)
		}
	}

	// Create a channel accept message and send it.
	accept := lcp.Accept(
		c.account.Address(),      // The account we use in the channel.
		client.WithRandomNonce(), // Our share of the channel nonce.
	)
	ch, err := r.Accept(context.TODO(), accept)
	if err != nil {
		fmt.Printf("Error accepting channel proposal: %v\n", err)
		return
	}

	// Start the on-chain event watcher. It automatically handles disputes.
	c.startWatching(ch)

	// Store channel.
	c.channels <- newPaymentChannel(ch, c.currency)
	c.AcceptedChannel()

}

// HandleUpdate is the callback for incoming channel updates.
func (c *PaymentClient) HandleUpdate(cur *channel.State, next client.ChannelUpdate, r *client.UpdateResponder) {
	// We accept every update that increases our balance.
	err := func() error {
		err := channel.AssertAssetsEqual(cur.Assets, next.State.Assets)
		if err != nil {
			return fmt.Errorf("Invalid assets: %v", err)
		}

		receiverIdx := 1 - next.ActorIdx // This works because we are in a two-party channel.
		curBal := cur.Allocation.Balance(receiverIdx, c.currency)
		nextBal := next.State.Allocation.Balance(receiverIdx, c.currency)
		if nextBal.Cmp(curBal) < 0 {
			return fmt.Errorf("Invalid balance: %v", nextBal)
		}
		return nil
	}()
	if err != nil {
		r.Reject(context.TODO(), err.Error()) //nolint:errcheck // It's OK if rejection fails.
	}

	// Send the acceptance message.
	err = r.Accept(context.TODO())
	if err != nil {
		panic(err)
	}
}

// HandleAdjudicatorEvent is the callback for smart contract events.
func (c *PaymentClient) HandleAdjudicatorEvent(e channel.AdjudicatorEvent) {
	log.Printf("Adjudicator event: type = %T, client = %v", e, c.account.Address())
}

func (c *PaymentClient) GetChannel() (*PaymentChannel, error) {
	select {
	case channel := <-c.channels:
		c.channels <- channel // Put the channel back into the channels channel.
		return channel, nil
	case <-time.After(time.Second): // Set a timeout duration (e.g., 1 second).
		return nil, fmt.Errorf("no channel available")
	}
}
