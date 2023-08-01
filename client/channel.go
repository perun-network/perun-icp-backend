package client

import (
	"context"
	"math/big"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
)

// PaymentChannel is a wrapper for a Perun channel for the payment use case.
type PaymentChannel struct {
	ch       *client.Channel
	currency channel.Asset
}

func (c *PaymentChannel) GetChannel() *client.Channel {
	return c.ch
}
func (c *PaymentChannel) GetChannelParams() *channel.Params {
	return c.ch.Params()
}

func (c *PaymentChannel) GetChannelState() *channel.State {
	return c.ch.State()
}

// newPaymentChannel creates a new payment channel.
func newPaymentChannel(ch *client.Channel, currency channel.Asset) *PaymentChannel {
	return &PaymentChannel{
		ch:       ch,
		currency: currency,
	}
}

// SendPayment sends a payment to the channel peer.
func (c PaymentChannel) SendPayment(amount int64) {
	// Transfer the given amount from us to peer.
	// Use UpdateBy to update the channel state.
	err := c.ch.Update(context.TODO(), func(state *channel.State) { // We use context.TODO to keep the code simple.
		icp := big.NewInt(amount)
		actor := c.ch.Idx()
		peer := 1 - actor
		state.Allocation.TransferBalance(actor, peer, c.currency, icp)
	})
	if err != nil {
		panic(err) // We panic on error to keep the code simple.
	}
}

// Settle settles the payment channel and withdraws the funds.
func (c PaymentChannel) Settle() {
	// If the channel is not finalized: Finalize the channel to enable fast settlement.

	if !c.ch.State().IsFinal {
		err := c.ch.Update(context.TODO(), func(state *channel.State) {
			state.IsFinal = true
		})
		if err != nil {
			panic(err)
		}
	}

	// Settle concludes the channel and withdraws the funds.
	err := c.ch.Settle(context.TODO(), false)
	if err != nil {
		panic(err)
	}

	// Close frees up channel resources.
	c.ch.Close()
}
