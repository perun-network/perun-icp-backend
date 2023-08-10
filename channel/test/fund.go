// SPDX-License-Identifier: Apache-2.0
package test

import (
	"context"
	"github.com/perun-network/perun-icp-backend/channel"
	pchannel "perun.network/go-perun/channel"
	pkgerrors "polycry.pt/poly-go/errors"
)

func FundAll(ctx context.Context, funders []*channel.Funder, reqs []*pchannel.FundingReq) error {
	g := pkgerrors.NewGatherer()
	for i := range funders {
		i := i
		g.Go(func() error {
			return funders[i].Fund(ctx, *reqs[i])
		})
	}

	if g.WaitDoneOrFailedCtx(ctx) {
		return ctx.Err()
	}
	return g.Err()
}
