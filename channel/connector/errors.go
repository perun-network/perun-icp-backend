// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"github.com/pkg/errors"
)

var (

	// ErrNonceOutOfRange a nonce was out of range of valid values.
	ErrNonceOutOfRange = errors.New("nonce values was out of range")
	// ErrAllocIncompatible an allocation was incompatible.
	ErrAllocIncompatible = errors.New("incompatible allocation")
	// ErrStateIncompatible a state was incompatible.
	ErrStateIncompatible = errors.New("incompatible state")
	// ErrIdentLenMismatch the length of an identity was wrong.
	ErrIdentLenMismatch = errors.New("length of an identity was wrong")
	// Channel was assumed concluded, but is not
	ErrNotConcluded = errors.New("channel not concluded")
	ErrFundTransfer = errors.New("funding transfer failed")
)
