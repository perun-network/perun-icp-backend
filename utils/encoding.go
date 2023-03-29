package utils

// here we defien the encoding structure to use the candid package of the agent-go library

import (
	"github.com/aviate-labs/agent-go/principal"
	"strconv"
    "strings"
    "fmt"
	"regexp"

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

func DecodePrincipal(principalString string) (principal.Principal, error) {
	decPrincipal, err := principal.Decode(principalString)
	if err != nil {
		return principal.Principal{}, fmt.Errorf("error decoding Principal String: %w", err)
	}
	return decPrincipal, nil
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
		if i % 2 == 0 {
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
	blocknumArg := fmt.Sprintf("%d : nat64", blocknum) + " )"
	fullArg += blocknumArg

	return fullArg
}

func ExtractBlock(s string) (uint64, error) {
	re := regexp.MustCompile(`\d+ = (\d+)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no value found")
	}

	value, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value: %w", err)
	}

	return value, nil
}

func ExtractTxAmount(s string) (int, error) {
	re := regexp.MustCompile(`\d+ = (\d+)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no value found")
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert value: %w", err)
	}

	return value, nil
}

func ExtractHoldingsNat(input string) (int, error) {
	re := regexp.MustCompile(`\d+(_\d+)*`)
	numberWithUnderscores := re.FindString(input)
	numberWithoutUnderscores := strings.Replace(numberWithUnderscores, "_", "", -1)

	natValue, err := strconv.Atoi(numberWithoutUnderscores)
	if err != nil {
		return 0, err
	}

	return natValue, nil
}