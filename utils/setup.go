// SPDX-License-Identifier: Apache-2.0
package utils

import (
	"os"
)

func SetHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDir
}
