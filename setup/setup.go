// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"fmt"
	"os/exec"

	"time"
)

// "./.config/.dfx/identities/minter/identity.pem", // minter test identity generated with keysmith:
// keysmith generate
// keysmith private-key
// -> this generates a seed.txt, then uses the seed.txt to generate the identity.pem
// https://github.com/dfinity/keysmith and imported with dfx 0.13.1

type DfxConfig struct {
	Host        string
	Port        int
	ExecPath    string
	AccountPath string // use local path to a minter .pem file
}

type UserConfig struct {
	Host        string
	Port        int
	Balance     uint64
	AccountPath string
}

type Dfx struct {
	ExecPath string
}

func (d *Dfx) StartDeployDfx(config DfxConfig) (*exec.Cmd, error) {
	err := d.checkDFXInstallation()
	if err != nil {
		return nil, fmt.Errorf("DFX CLI Environment not installed. Check installation typing 'dfx --version' in your terminal %v", err)
	}

	path, err := exec.LookPath("dfx")
	if err != nil {
		return nil, err
	}

	dfx, err := d.startDFX(path, config.ExecPath)
	if err != nil {
		return nil, err
	}

	err = DeployCanisters(path, config.ExecPath)
	if err != nil {
		return nil, err
	}

	return dfx, nil
}

func (d *Dfx) checkDFXInstallation() error {
	_, err := exec.LookPath("dfx")
	return err
}

func (d *Dfx) startDFX(path, execPath string) (*exec.Cmd, error) {
	dfx := exec.Command(path, "start", "--background", "--clean", "--host", "127.0.0.1:4943")
	dfx.Dir = execPath

	err := dfx.Start()
	if err != nil {
		return nil, err
	}

	fmt.Println("Starting DFX...")
	time.Sleep(3 * time.Second)
	return dfx, nil
}

func (d *Dfx) StopDFX(dfx *exec.Cmd) error {
	path, err := exec.LookPath("dfx")
	if err != nil {
		return err
	}

	cmd := exec.Command(path, "stop")
	cmd.Dir = d.ExecPath
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("Stopping DFX...")
	if err := dfx.Process.Kill(); err != nil {
		return err
	}

	fmt.Println("Stopped DFX.")
	return nil
}

func DeployCanisters(path, execPath string) error {
	ledgerArg := createLedgerArg()
	err := deployLedger(path, execPath, ledgerArg)
	if err != nil {
		return err
	}

	err = deployPerun(path, execPath)
	if err != nil {
		return err
	}

	fmt.Println("Deployed Canisters.")

	return nil
}

func createLedgerArg() string {
	//dfx ledger account-id
	const (
		ICP_PERUN_MINT_ACC  = "433bd8e9dd65bdfb34259667578e749136f3e0ea1566e10af1e0dd324cbd9144"
		ICP_PERUN_USERA_ACC = "97520b79b03e38d3f6b38ce5026a813ccc9d1a3e830edb6df5970e6ca6ad84be"
		ICP_PERUN_USERB_ACC = "40fd2dc85bc7d264b31f1fa24081d7733d303b49b7df84e3d372338f460aa678"
	)

	return fmt.Sprintf(
		"(record {minting_account = \"%s\"; initial_values = vec { record { \"%s\"; record { e8s=80_000_000 } }; record { \"%s\"; record { e8s=120_000_000 } }}; send_whitelist = vec {}})",
		ICP_PERUN_MINT_ACC, ICP_PERUN_USERA_ACC, ICP_PERUN_USERB_ACC,
	)
}

func deployLedger(path, execPath, ledgerArg string) error {
	fmt.Println("Deploying the Ledger with the following parameters: ", ledgerArg)
	deployLedger := exec.Command(path, "deploy", "ledger", "--argument", ledgerArg)
	deployLedger.Dir = execPath

	outputLedger, err := deployLedger.CombinedOutput()
	if err != nil {
		fmt.Printf("Error deploying ledger:\n%s\n", string(outputLedger))
		return fmt.Errorf("error deploying ledger: %w", err)
	}

	return nil
}

func deployPerun(path, execPath string) error {
	deployPerun := exec.Command(path, "deploy", "icp_perun")
	deployPerun.Dir = execPath

	_, err := deployPerun.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error deploying icp_perun: %w", err)
	}

	return nil
}

type DfxSetup struct {
	DfxInstance *Dfx
	DfxConfig   DfxConfig
	DfxCmd      *exec.Cmd
}

func NewDfxSetup(config DfxConfig) *DfxSetup {
	return &DfxSetup{
		DfxInstance: &Dfx{},
		DfxConfig:   config,
	}
}

func (d *DfxSetup) StartDeployDfx() error {
	dfxCmd, err := d.DfxInstance.StartDeployDfx(d.DfxConfig)
	if err != nil {
		return err
	}
	d.DfxCmd = dfxCmd
	return nil
}

func (d *DfxSetup) StopDFX() error {
	return d.DfxInstance.StopDFX(d.DfxCmd)
}
