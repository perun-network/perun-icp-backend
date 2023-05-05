// SPDX-License-Identifier: Apache-2.0
package setup_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"perun.network/perun-icp-backend/setup"
	"perun.network/perun-icp-backend/utils"
	"testing"
)

func TestDFXEnvironment(t *testing.T) {
	// Define a test config
	testConfig := setup.DfxConfig{
		Host:        "http://127.0.0.1",
		Port:        8000,
		ExecPath:    "../test/testdata/",
		AccountPath: filepath.Join(utils.SetHomeDir(), ".config", "dfx", "identity", "minter", "identity.pem"),
	}

	// Create a new Setup instance
	testSetup := setup.NewDfxSetup(testConfig)

	// Start and deploy DFX
	err := testSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")
	assert.NotNil(t, testSetup.DfxCmd, "DFX cmd should not be nil")

	// Stop DFX
	err = testSetup.StopDFX()
	assert.NoError(t, err, "Failed to stop DFX environment")
}
