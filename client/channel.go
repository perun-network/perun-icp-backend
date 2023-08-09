package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"strconv"
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
	err := c.ch.Update(context.TODO(), func(state *channel.State) {
		icp := big.NewInt(amount)
		actor := c.ch.Idx()
		peer := 1 - actor
		state.Allocation.TransferBalance(actor, peer, c.currency, icp)
	})
	if err != nil {
		panic(err) // We panic on error to keep the code simple.
	}
}

func (p *PaymentClient) SendPaymentToPeer(amount float64) {
	if !p.HasOpenChannel() {
		return
	}
	amountInt64 := int64(amount)
	p.Channel.SendPayment(amountInt64)
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

func FormatState(c *PaymentChannel, state *channel.State) string {
	id := c.ch.ID()
	parties := c.ch.Params().Parts

	bigIntA := state.Allocation.Balance(0, c.currency)
	bigFloatA := new(big.Float).SetInt(bigIntA)
	balA, _ := bigFloatA.Float64()
	balAStr := strconv.FormatFloat(balA, 'f', 4, 64)

	fstPartyPaymentAddr := parties[0].String()
	sndPartyPaymentAddr := parties[1].String()

	bigIntB := state.Allocation.Balance(1, c.currency)
	bigFloatB := new(big.Float).SetInt(bigIntB)
	balB, _ := bigFloatB.Float64()

	balBStr := strconv.FormatFloat(balB, 'f', 4, 64)
	if len(parties) != 2 {
		log.Fatalf("invalid parties length: " + strconv.Itoa(len(parties)))
	}
	ret := fmt.Sprintf(
		"Channel ID: [green]%s[white]\nBalances:\n    %s: [green]%s[white] IC Token\n    %s: [green]%s[white] IC Token\nFinal: [green]%t[white]\nVersion: [green]%d[white]",
		hex.EncodeToString(id[:]),
		fstPartyPaymentAddr,
		balAStr,
		sndPartyPaymentAddr,
		balBStr,
		state.IsFinal,
		state.Version,
	)
	return ret
}

func (p *PaymentClient) Settle() {
	if !p.HasOpenChannel() {
		return
	}
	p.Channel.Settle()
}
