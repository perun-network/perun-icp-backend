// SPDX-License-Identifier: Apache-2.0
package channel

import (
	"perun.network/go-perun/channel"
)

func init() {
	channel.SetBackend(new(backend))
}
