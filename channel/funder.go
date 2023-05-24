// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	pwallet "perun.network/go-perun/wallet"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	utils "perun.network/perun-icp-backend/utils"
	"perun.network/perun-icp-backend/wallet"
	"sync"

	"time"
)

type DepositReq struct {
	Balance pchannel.Bal
	Fee     pchannel.Bal
	Account wallet.Account
	Funding chanconn.Funding
}

type Funder struct {
	acc  *wallet.Account
	log  log.Embedding
	conn *chanconn.Connector
}

func (f *Funder) GetAcc() *wallet.Account {
	return f.acc
}

func (d *Depositor) Deposit(ctx context.Context, req *DepositReq) error {
	depositArgs, err := d.cnr.BuildDeposit(req.Account, req.Balance, req.Fee, req.Funding)
	if err != nil {
		return fmt.Errorf("failed to build deposit: %w", err)
	}

	blockNum, err := d.cnr.ExecuteDFXTransfer(depositArgs, *d.cnr.LedgerID, d.cnr.ExecPath, d.cnr.TransferDfx)
	if err != nil {
		return fmt.Errorf("failed to execute DFX transfer during channel opening: %w", err)
	}

	_, err = d.cnr.NotifyTransferToPerun(blockNum, *d.cnr.PerunID, d.cnr.ExecPath)

	if err != nil {
		return fmt.Errorf("failed to notify transfer to perun: %w", err)
	}

	addr := req.Account.ICPAddress()
	memo, err := req.Funding.Memo()
	if err != nil {
		return fmt.Errorf("failed to get memo from funding: %w", err)
	}

	_, err = d.cnr.DepositToPerunChannel(addr, req.Funding.Channel, memo, *d.cnr.PerunID, d.cnr.ExecPath)
	if err != nil {
		return fmt.Errorf("failed to deposit to perun channel: %w", err)
	}

	return nil
}

func (f *Funder) FundAG(ctx context.Context, req pchannel.FundingReq) error {

	// timestamp the funding procedure
	tstamp := time.Now().UnixNano()

	wReq, err := NewDepositReqFromPerun(&req, f.acc)

	if err != nil {
		return err
	}
	if err := NewDepositor(f.conn).Deposit(ctx, wReq); err != nil {
		return err
	}

	chanID := wReq.Funding.Channel

	qEventsvArgs := utils.FormatChanTimeArgs([]byte(chanID[:]), uint64(tstamp))

	eventsString, err := f.conn.QueryEventsCLI(qEventsvArgs, *f.conn.PerunID, f.conn.ExecPath)
	if err != nil {
		return fmt.Errorf("Error for parsing channel events: %v", err)
	}

	eventList, err := chanconn.StringIntoEvents(eventsString)
	if err != nil {
		return fmt.Errorf("Error for parsing channel events: %v", err)
	}

	evli := make(chan chanconn.Event, 1)

	go func() {
		for _, event := range eventList {
			evli <- event
		}
	}()

	return nil
}

type FunderWithMutex struct {
	Funder *Funder
	Mutex  *sync.Mutex
}

func (fm *FunderWithMutex) Fund(ctx context.Context, req pchannel.FundingReq) error {
	fm.Mutex.Lock()

	defer fm.Mutex.Unlock()

	return fm.Funder.Fund(ctx, req)
}
func (f *Funder) Fund(ctx context.Context, req pchannel.FundingReq) error {
	f.conn.Mutex.Lock()

	defer f.conn.Mutex.Unlock()

	return fundLocked(ctx, req, f.acc, f.conn)
}

func fundLocked(ctx context.Context, req pchannel.FundingReq, acc pwallet.Account, conn *chanconn.Connector) error {
	tstamp := time.Now().UnixNano()

	wReq, err := NewDepositReqFromPerun(&req, acc)
	if err != nil {
		return err
	}

	if err := NewDepositor(conn).Deposit(ctx, wReq); err != nil {
		return err
	}

	chanID := wReq.Funding.Channel

	qEventsvArgs := utils.FormatChanTimeArgs([]byte(chanID[:]), uint64(tstamp))

	eventsString, err := conn.QueryEventsCLI(qEventsvArgs, *conn.PerunID, conn.ExecPath)
	if err != nil {
		return fmt.Errorf("Error for parsing channel events: %v", err)
	}

	eventList, err := chanconn.StringIntoEvents(eventsString)
	if err != nil {
		return fmt.Errorf("Error for parsing channel events: %v", err)
	}

	evli := make(chan chanconn.Event, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for _, event := range eventList {
			evli <- event
		}
	}()

	// Wait for the NotifyTransferToPerun operation to complete.
	wg.Wait()
	return nil
}

func (f *Funder) waitforFundings(ctx context.Context, evLi chan chanconn.Event, req pchannel.FundingReq) error {
	fundingEventCount := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-evLi: //src.Events():
			if event.EventType == "Funded" {
				fundingEventCount++
				if fundingEventCount == 1 {
					return nil
				}
			}
		}
	}
}

func NewFunder(acc wallet.Account, c *chanconn.Connector) *Funder {
	return &Funder{
		acc:  &acc,
		conn: c,
		log:  log.MakeEmbedding(log.Default()),
	}
}

func NewDepositReqFromPerun(req *pchannel.FundingReq, acc pwallet.Account) (*DepositReq, error) {
	if !req.Agreement.Equal(req.State.Balances) && (len(req.Agreement) == 1) {
		return nil, chanconn.ErrFundingReqIncompatible
	}
	bal := req.Agreement[0][req.Idx]
	fee := big.NewInt(0)
	fReq, err := MakeFundingReq(req)
	if err != nil {
		return nil, errors.WithMessage(chanconn.ErrFundingReqIncompatible, err.Error())
	}
	convAcc := *acc.(*wallet.Account)
	return NewDepositReq(bal, fee, convAcc, fReq), err
}

func MakeFundingReq(req *pchannel.FundingReq) (chanconn.Funding, error) {
	ident, err := chanconn.MakeOffIdent(req.Params.Parts[req.Idx])

	return chanconn.Funding{
		Channel: req.State.ID,
		Part:    ident,
	}, err
}

func NewDepositReq(bal, fee pchannel.Bal, acc wallet.Account, funding chanconn.Funding) *DepositReq {
	return &DepositReq{bal, fee, acc, funding}
}

// NewDepositor returns a new Depositor.
func NewDepositor(cnr *chanconn.Connector) *Depositor {
	return &Depositor{log.MakeEmbedding(log.Default()), cnr}
}

func (d *Depositor) VerifySig(nonce chanconn.Nonce, parts []pwallet.Address, chDur uint64, chanId chanconn.ChannelID, vers chanconn.Version, alloc *pchannel.Allocation, finalized bool, sigs []pwallet.Sig) (string, error) {

	execPath := d.cnr.ExecPath
	canID := d.cnr.PerunID
	verifyResult, err := d.cnr.VerifySig(nonce, parts, chDur, chanId, vers, alloc, finalized, sigs, *canID, execPath)
	if err != nil {
		return "", fmt.Errorf("failed to build deposit: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get memo from funding: %w", err)
	}

	return verifyResult, nil
}

type Depositor struct {
	log.Embedding

	cnr *chanconn.Connector
}
