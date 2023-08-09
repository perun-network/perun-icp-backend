package channel

import (
	"errors"
)

const ResponseErrorConcludingChannel = "error concluding the channel"
const WithdrawalSuccessResponse = "successful withdrawal"
const DisputeSuccess = "successful initialization of a dispute"

var (
	ErrNotFundedInTime            = errors.New("funding not in time")
	ErrFundingReqIncompatible     = errors.New("incompatible funding request")
	ErrFailWithdrawal             = errors.New("withdrawal failed")
	ErrFailDispute                = errors.New("error disputing")
	ErrFailConclude               = errors.New("error concluding")
	ErrFinalizedNotConcludable    = errors.New("channel finalized but not concludable")
	ErrConcludedDifferentVersion  = errors.New("channel was concluded with a different version")
	ErrAdjudicatorReqIncompatible = errors.New("adjudicator request was not compatible")
	ErrReqVersionTooLow           = errors.New("request version too low")
)

func WrapError(err error, msg string) error {
	return errors.New(msg + ": " + err.Error())
}
