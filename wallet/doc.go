// Copyright 2023 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package wallet contains the off-chain identity and signature handling of
// go-perun's internet computer backend. It uses ed25519 keys as identities and
// the EdDSA signature algorithm. Anonymously import the package from your
// application to inject the backend into go-perun.
package wallet // import "perun.network/perun-icp-backend/wallet"
