// SPDX-License-Identifier: Apache-2.0
package test

import (
	"context"
	"fmt"
	pchannel "perun.network/go-perun/channel"
	"perun.network/perun-icp-backend/channel"
	pkgerrors "polycry.pt/poly-go/errors"
	"sync"
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

func FundMtx(ctx context.Context, funders []*channel.Funder, reqs []*pchannel.FundingReq) error {
	g := pkgerrors.NewGatherer()

	fundersWithMutex := make([]*channel.FunderWithMutex, len(funders))

	glbMtx := &sync.Mutex{}

	for i := range funders {

		fundersWithMutex[i] = &channel.FunderWithMutex{
			Funder: funders[i],
			Mutex:  glbMtx,
		}
	}

	for i := range fundersWithMutex {
		i := i
		g.Go(func() error {
			return fundersWithMutex[i].Fund(ctx, *reqs[i])
		})
	}

	err := g.Wait() // Wait for all goroutines to finish.

	if err != nil {
		return fmt.Errorf("failed to fund: %w", err)
	}

	return nil
}

// FundAll executes all requests with the given funders in parallel.
func FundConc(ctx context.Context, funders []*channel.Funder, reqs []*pchannel.FundingReq) error {
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

func FundAllAG(ctx context.Context, funders []*channel.Funder, reqs []*pchannel.FundingReq) error {
	for i := range funders {
		err := funders[i].FundAG(ctx, *reqs[i])
		if err != nil {
			return err
		}

	}
	return nil
}
