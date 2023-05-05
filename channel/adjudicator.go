// SPDX-License-Identifier: Apache-2.0
package channel

//package connector

import (
	"github.com/pkg/errors"
	"perun.network/go-perun/log"
	chanconn "perun.network/perun-icp-backend/channel/connector"

	pwallet "perun.network/go-perun/wallet"
)

// Adjudicator implements the Perun Adjudicator interface.
type Adjudicator struct {
	log.Embedding

	cnr     *chanconn.Connector
	onChain pwallet.Account
}

var (
	// ErrConcludedDifferentVersion a channel was concluded with a different version.
	ErrConcludedDifferentVersion = errors.New("channel was concluded with a different version")
	// ErrAdjudicatorReqIncompatible the adjudicator request was not compatible.
	ErrAdjudicatorReqIncompatible = errors.New("adjudicator request was not compatible")
	// ErrAdjudicatorReqIncompatible the adjudicator request was not compatible.
	ErrReqVersionTooLow = errors.New("request version too low")
)

// NewAdjudicator returns a new Adjudicator.
func NewAdjudicator(onChain pwallet.Account, conn *chanconn.Connector) *Adjudicator {
	return &Adjudicator{log.MakeEmbedding(log.Default()), conn, onChain}
}
