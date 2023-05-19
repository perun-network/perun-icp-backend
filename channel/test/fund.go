// SPDX-License-Identifier: Apache-2.0
package test

import (
	"context"
	pchannel "perun.network/go-perun/channel"
	"perun.network/perun-icp-backend/channel"
)

// FundAll executes all requests with the given funders
func FundAll(ctx context.Context, funders []*channel.Funder, reqs []*pchannel.FundingReq) error {
	for i := range funders {
		err := funders[i].Fund(ctx, *reqs[i])
		if err != nil {
			return err
		}

	}
	return nil
}

func FundAllAG(ctx context.Context, funders []*channel.Funder, reqs []*pchannel.FundingReq) error {
	for i := range funders {
		err := funders[i].FundAG(ctx, *reqs[i])
		if err != nil {
			return err
		}

	}
	return nil
}
