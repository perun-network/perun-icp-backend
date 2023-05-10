// SPDX-License-Identifier: Apache-2.0
package test

import (
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

	s := chtest.NewMinterSetup(t)
	c := s.Conns

	ret := &Setup{Setup: s}

	for i := 0; i < len(s.Accs); i++ {
		dep := channel.NewDepositor(c[i])
		ret.Deps = append(ret.Deps, dep)
		ret.Funders = append(ret.Funders, channel.NewFunder(&s.Accs[i], c[i]))
		ret.Adjs = append(ret.Adjs, channel.NewAdjudicator(&s.Accs[i], c[i]))
	}

	return ret
}
