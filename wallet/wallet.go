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

package wallet

import (
	"bytes"
	cr "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	ed "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"io"
	"math/rand"
	"os"
	"sync"

	"perun.network/go-perun/wallet"
)

// FsWallet is a garbage-collected file system key store, removing all keys when
// they are no longer used. Generated keys will not be persisted to permanent
// storage unless IncrementUsage() is called on them. Once a key is no longer
// used (as indicated by DecrementUsage()), it is deleted from storage.
type FsWallet struct {
	mutex sync.Mutex
	file  string

	seed      [24]byte            // the wallet's random seed.
	latestAcc uint64              // the next account's nonce.
	openAccs  map[string]*openAcc // all currently stored accounts.
}

type openAcc struct {
	nonce    uint64
	useCount uint32
	acc      Account
}

var bo = binary.LittleEndian

// NewRAMWallet creates an unpersisted FsWallet.
func NewRAMWallet(gen io.Reader) (*FsWallet, error) {
	w := FsWallet{
		openAccs: make(map[string]*openAcc),
	}

	_, err := io.ReadFull(gen, w.seed[:])
	if err != nil {
		//return nil, fmt.Errorf("error reading random seed: %v", err)
		panic(err)
	}

	return &w, nil
}

func NewWallet() *FsWallet {
	wlt, err := NewRAMWallet(cr.Reader)
	if err != nil {
		panic(err)
	}
	return wlt
}

// CreateOrLoadFsWallet loads the wallet from the requested path, otherwise, it
// creates a new one and saves it to the requested path.
func CreateOrLoadFsWallet(path string, gen io.Reader) (*FsWallet, error) {
	w := FsWallet{
		file:     path,
		openAccs: make(map[string]*openAcc),
	}

	if file, err := os.ReadFile(path); err == nil {
		r := bytes.NewReader(file)
		if err := w.load(r); err != nil {
			return nil, err
		}
	} else {
		if _, err := io.ReadFull(gen, w.seed[:]); err != nil {
			return nil, err
		}
		if err := w.save(); err != nil {
			return nil, err
		}
	}
	return &w, nil
}

func (w *FsWallet) load(r io.Reader) error {
	if _, err := io.ReadFull(r, w.seed[:]); err != nil {
		return err
	}
	if err := binary.Read(r, bo, &w.latestAcc); err != nil {
		return err
	}
	var openAccs uint32
	if err := binary.Read(r, bo, &openAccs); err != nil {
		return err
	}
	w.openAccs = make(map[string]*openAcc, openAccs)
	for i := uint32(0); i < openAccs; i++ {
		pk := make(Address, ed.PublicKeySize)
		if _, err := io.ReadFull(r, pk[:]); err != nil {
			return err
		}

		acc := &openAcc{}
		if err := binary.Read(r, bo, &acc.nonce); err != nil {
			return err
		}
		if err := binary.Read(r, bo, &acc.useCount); err != nil {
			return err
		}

		w.openAccs[string(pk)] = acc
	}
	return nil
}

func (w *FsWallet) save() error {
	if w.file == "" {
		return nil
	}

	file := new(bytes.Buffer)
	file.Write(w.seed[:])

	err := binary.Write(file, bo, w.latestAcc)
	if err != nil {
		panic(err)
	}

	err = binary.Write(file, bo, uint32(len(w.openAccs)))
	if err != nil {
		panic(err)
	}

	for pk, acc := range w.openAccs {
		file.Write([]byte(pk))

		err = binary.Write(file, bo, acc.nonce)
		if err != nil {
			panic(err)
		}

		err = binary.Write(file, bo, acc.useCount)
		if err != nil {
			panic(err)
		}
	}

	return os.WriteFile(w.file, file.Bytes(), 0644)
}

func (w *FsWallet) genAcc(id uint64) Account {
	seed := new(bytes.Buffer)
	seed.Write(w.seed[:])

	err := binary.Write(seed, bo, id)
	if err != nil {
		panic(fmt.Sprintf("error writing id to seed buffer: %v", err))
	}

	_, sk, err := ed.GenerateKey(seed)
	if err != nil {
		panic("logic error: generating key should not have failed")
	}

	return Account(sk)
}

// NewAccount creates a fresh unlocked account. This account is not persisted
// until IncrementUsage() is called on it.
func (w *FsWallet) NewAccount() Account {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	acc := w.genAcc(w.latestAcc)
	w.openAccs[string(*acc.Address().(*Address))] = &openAcc{
		nonce:    w.latestAcc,
		useCount: 0,
		acc:      acc,
	}

	w.latestAcc++
	return acc
}

// NewRandomAccount creates a new random account using the wallet package.
func (w *FsWallet) NewRandomAccount(_ *rand.Rand) wallet.Account {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	acc := w.genAcc(w.latestAcc)
	w.openAccs[string(*acc.Address().(*Address))] = &openAcc{
		nonce:    w.latestAcc,
		useCount: 0,
		acc:      acc,
	}

	w.latestAcc++
	return acc
}

// Unlock retrieves the account belonging to the requested address.
func (w *FsWallet) Unlock(a wallet.Address) (wallet.Account, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	addr := *a.(*Address)
	acc, ok := w.openAccs[string(addr[:])]
	if !ok {
		return nil, errors.New("no such account")
	}

	if acc.acc == nil {
		acc.acc = w.genAcc(acc.nonce)
	}
	return acc.acc, nil
}

// LockAll disables all currently unlocked accounts.
func (w *FsWallet) LockAll() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, acc := range w.openAccs {
		acc.acc.clear()
		acc.acc = nil
	}
}

// IncrementUsage tracks how many times an account is in use. Use
// DecrementUsage() when an account is no longer used. Once the counter reaches
// 0, the account is deleted.
func (w *FsWallet) IncrementUsage(a wallet.Address) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	acc, ok := w.openAccs[string(*a.(*Address))]
	if !ok {
		panic("IncrementUsage: account not found")
	}
	acc.useCount++

	if err := w.save(); err != nil {
		panic("Error in IncrementUsage during save")
	}
}

// DecrementUsage completements IncrementUsage().
func (w *FsWallet) DecrementUsage(a wallet.Address) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	key := string(*a.(*Address))
	acc, ok := w.openAccs[key]
	if !ok {
		panic("IncrementUsage: account not found!")
	}
	if acc.useCount == 0 {
		panic("DecrementUsage: unused account!")
	}
	acc.useCount--
	if acc.useCount == 0 {
		acc.acc.clear()
		delete(w.openAccs, key)
	}

	if err := w.save(); err != nil {
		panic("Error in DecrementUsage during save")
	}
}
