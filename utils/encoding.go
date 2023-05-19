// SPDX-License-Identifier: Apache-2.0

package utils

// here we defien the encoding structure to use the candid package of the agent-go library

import (
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"regexp"
	"strconv"
	"strings"
)

// We encode the data according to the Motoko format as a vec{nat8} represenation

func ByteToVecString(memo []byte) string {
	var str string
	for i, b := range memo {
		str += strconv.Itoa(int(b))
		if i < len(memo)-1 {
			str += "; "
		}
	}
	return "vec{" + str + "}"
}

func DecodePrincipal(principalString string) (*principal.Principal, error) {
	decPrincipal, err := principal.Decode(principalString)
	if err != nil {
		return &principal.Principal{}, fmt.Errorf("error decoding Principal String: %w", err)
	}
	return &decPrincipal, nil
}

func FormatWithUnderscores(n uint64) string {
	s := fmt.Sprintf("%d", n)
	parts := make([]string, 0, (len(s)+2)/3)

	for len(s) > 0 {
		chunkSize := len(s) % 3
		if chunkSize == 0 {
			chunkSize = 3
		}
		parts = append(parts, s[:chunkSize])
		s = s[chunkSize:]
	}

	return strings.Join(parts, "_")
}

func FormatFundingArgs(addr, chanId []byte) string {
	return fmt.Sprintf("( record { channel = blob\"%s\"; participant = %s } )", FormatHexByte(chanId), FormatVec(addr))
}
func FormatFidArgs(addr, chanId []byte) string {
	return fmt.Sprintf("(record {channel = %s; participant = %s})", FormatVec(chanId), FormatVec(addr))
}

func FormatParamsArgs(nonce []byte, parts [][]byte, duration uint64) string {
	return fmt.Sprintf("(record {nonce = %s; participants = vec{%s ; %s} ; duration = %d : nat64})", FormatVec(nonce), FormatVec(parts[0]), FormatVec(parts[1]), duration)
}

func FormatStateArgs(chanId []byte, version uint64, alloc []uint64, finalized bool) string {
	return fmt.Sprintf("(record {channel = %s; version = %d : nat64; allocation = vec{%d ; %d} ; isFinal = %t : bool})", FormatVec(chanId), version, alloc[0], alloc[1], finalized)
}

// gwjf3-rxk3d-lfwux-5evls-qw4gc-fyh4e-ohkeg-zg32g-vqfcw-yaaqs-tqe
func FormatWithdrawalArgs(addr, chanId, sig []byte) string { //, prince string
	return fmt.Sprintf("(record { channel = blob \"%s\"; participant = %s ; receiver = principal \"exqrz-uemtb-qnd6t-mvbn7-mxjre-bodlr-jnqql-tnaxm-ur6uc-mmgb4-jqe\" ; signature = %s})", FormatHexByte(chanId), FormatVec(addr), FormatVec(sig))
}

func FormatConcludeCLIArgs(nonce []byte, addrs [][]byte, chDur uint64, chanId []byte, version uint64, alloc []int, finalized bool, sig [][]byte) string { //, prince string
	return fmt.Sprintf(
		"(record { nonce = blob \"%s\"; participants = vec{ %s ; %s } ; challenge_duration = %d: nat64 ; channel = blob \"%s\" ; version = %d : nat64; allocation = vec{ %d : nat ; %d : nat } ; finalized = %t : bool ; sigs = vec{ %s ; %s}})",
		FormatHexByte(nonce), FormatVec(addrs[0]), FormatVec(addrs[1]), chDur, FormatHexByte(chanId), version, alloc[0], alloc[1], finalized, FormatVec(sig[0]), FormatVec(sig[1]))
}

func FormatConcludeAGArgs(nonce []byte, addrs [][]byte, chDur uint64, chanId []byte, version uint64, alloc []int, finalized bool, sig [][]byte) string { //, prince string
	return fmt.Sprintf(
		"(record { nonce = blob \"%s\"; participants = vec{ blob \"%s\"; blob \"%s\"} ; challenge_duration = %d: nat64 ; channel = blob \"%s\" ; version = %d : nat64; allocation = vec{ %d : nat ; %d : nat } ; finalized = %t : bool ; sigs = vec{ blob \"%s\" ; blob \"%s\"}})",
		FormatHexByte(nonce), FormatHexByte(addrs[0]), FormatHexByte(addrs[1]), chDur, FormatHexByte(chanId), version, alloc[0], alloc[1], finalized, FormatHexByte(sig[0]), FormatHexByte(sig[1]))
}

func FormatFundingMemoArgs(addr, chanId []byte, memo uint64) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(
		"(record {channel = blob\"%s\"; participant = %s; memo = %d : nat64 })",
		FormatHexByte(chanId),
		FormatVec(addr),
		memo,
	))

	return builder.String()
}

func FormatHexByte(input []byte) string {
	var result strings.Builder

	for _, b := range input {
		result.WriteString(fmt.Sprintf("%02x", b))
	}

	return result.String()
}

func FormatHex(hexStr string) string {
	var result strings.Builder

	for i := 0; i < len(hexStr); i += 2 {
		if i > 0 || i == 0 {
			result.WriteString("\\")
		}
		result.WriteString(hexStr[i : i+2])
	}

	return result.String()
}

func InsertBackslash(hash string) string {
	modified := ""
	for i := 0; i < len(hash); i++ {
		if i%2 == 0 {
			modified += "\\"
		}
		modified += string(hash[i])
	}
	return modified
}

func FormatVec(data []uint8) string {
	var elements []string
	for _, element := range data {
		elements = append(elements, fmt.Sprintf("%d", element))
	}
	return "vec {" + strings.Join(elements, "; ") + "}"
}

func FormatNotifyArgs(blocknum uint64) string {
	fullArg := "("
	blocknumArg := fmt.Sprintf("%d : nat64", blocknum) + ")"

	fullArg += blocknumArg

	return fullArg
}

func ExtractBlock(s string) (uint64, error) {
	re := regexp.MustCompile(`\d+ = (\d+)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no value found extracting the block")
	}

	value, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value: %w", err)
	}

	return value, nil
}

func FormatVerifySigArgs(addrs [][]byte, sig [][]byte, chanId []byte, version uint64, alloc []int, finalized bool) string {
	return fmt.Sprintf(
		"(record { participants = vec{ %s; %s} ; signatures = vec{ %s ; %s}} ; channel = blob \"%s\" ; version = %d : nat64; allocation = vec{ %d ; %d } ; finalized = %t : bool ; )",
		FormatVec(addrs[0]), FormatVec(addrs[1]), FormatVec(sig[0]), FormatVec(sig[1]), FormatHexByte(chanId), version, alloc[0], alloc[1], finalized)
}

func FormatTransferArgs(memo, amount, fee uint64, sendTo string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(
		"(record {memo = %d : nat64; amount = record { e8s=%s : nat64}; fee = record { e8s=%s : nat64}; from_subaccount = null; to = blob \"%s\"; created_at_time = null; })",
		memo,
		FormatWithUnderscores(amount),
		FormatWithUnderscores(fee),
		FormatHex(sendTo),
	))

	return builder.String()
}

func ExtractTxAmount(s string) (int, error) {
	re := regexp.MustCompile(`\d+ = (\d+)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no value found extracting the amount")
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert value: %w", err)
	}

	return value, nil
}

func ExtractHoldingsNat(input string) (uint64, error) {
	re := regexp.MustCompile(`\d+(_\d+)*`)
	numberWithUnderscores := re.FindString(input)
	numberWithoutUnderscores := strings.Replace(numberWithUnderscores, "_", "", -1)

	natValue, err := strconv.Atoi(numberWithoutUnderscores)
	if err != nil {
		return 0, err
	}

	return uint64(natValue), nil
}

func FormatChanTimeArgs(chanId []byte, tstamp uint64) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(
		"(record {chanid = blob \"%s\";  time = %d : nat64 })",
		FormatHexByte(chanId),
		tstamp,
	))

	return builder.String()
}
