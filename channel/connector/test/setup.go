// SPDX-License-Identifier: Apache-2.0
package test

import (
	"fmt"

	"perun.network/perun-icp-backend/channel"
	chtest "perun.network/perun-icp-backend/channel/test"

	"testing"
)

type Setup struct {
	*chtest.Setup

	Deps    []*channel.Depositor
	Funders []*channel.Funder
	Adjs    []*channel.Adjudicator
}

func NewSetup(t *testing.T) *Setup {

	s := chtest.NewSetup(t)
	c := s.Conns

	ret := &Setup{Setup: s}

	for i := 0; i < len(s.Accs); i++ {
		fmt.Println("NewConn: ", c[i])
		dep := channel.NewDepositor(c[i])
		ret.Deps = append(ret.Deps, dep)
		ret.Funders = append(ret.Funders, channel.NewFunder(&s.Accs[i], c[i]))
		//ret.Adjs = append(ret.Adjs, pallet.NewAdjudicator(s.Accs[i].Acc, p, s.API, PastBlocks))
	}

	return ret
}
