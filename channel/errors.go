// Copyright 2023 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
