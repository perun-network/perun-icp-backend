package channel

import (
	"errors"
)

const ResponseErrorConcludingChannel = "error concluding the channel"

var (
	ErrNotFundedInTime         = errors.New("funding not in time")
	ErrFundingReqIncompatible  = errors.New("incompatible funding request")
	ErrFailWithdrawal          = errors.New("withdrawal failed")
	ErrFailDispute             = errors.New("error disputing")
	ErrFailConclude            = errors.New("error concluding")
	ErrFinalizedNotConcludable = errors.New("channel finalized but not concludable")
)

func WrapError(err error, msg string) error {
	return errors.New(msg + ": " + err.Error())
}
