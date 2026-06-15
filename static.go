package main

import (
	"encoding/json"
	"io"
	"math"
	"os"
	"strings"
)

// StaticResult represents the JSON output of the static analysis worker.
type StaticResult struct {
	Length         int     `json:"length"`
	PoolSize       int     `json:"pool_size"`
	Entropy        float64 `json:"entropy"`
	HasLowercase   bool    `json:"has_lowercase"`
	HasUppercase   bool    `json:"has_uppercase"`
	HasDigit       bool    `json:"has_digit"`
	HasSpecial     bool    `json:"has_special"`
	MaxConsecutive int     `json:"max_consecutive"`
	HasSequential  bool    `json:"has_sequential"`
	Error          string  `json:"error,omitempty"`
}

// runStaticAnalysis reads a password from standard input, calculates static complexity metrics,
// and writes the JSON serialized StaticResult to stdout.
func runStaticAnalysis() {
	var result StaticResult

	// Read all input from stdin
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		result.Error = "failed to read from stdin: " + err.Error()
		writeJSONResponse(result)
		return
	}

	// Clean trailing newlines
	password := strings.TrimRight(string(inputBytes), "\r\n")
	result.Length = len([]rune(password))

	if result.Length == 0 {
		writeJSONResponse(result)
		return
	}

	// Analyze character diversity
	var (
		lowercaseCount = 0
		uppercaseCount = 0
		digitCount     = 0
		specialCount   = 0
	)

	runes := []rune(password)
	for _, r := range runes {
		if r >= 'a' && r <= 'z' {
			lowercaseCount++
		} else if r >= 'A' && r <= 'Z' {
			uppercaseCount++
		} else if r >= '0' && r <= '9' {
			digitCount++
		} else {
			specialCount++
		}
	}

	poolSize := 0
	if lowercaseCount > 0 {
		result.HasLowercase = true
		poolSize += 26
	}
	if uppercaseCount > 0 {
		result.HasUppercase = true
		poolSize += 26
	}
	if digitCount > 0 {
		result.HasDigit = true
		poolSize += 10
	}
	if specialCount > 0 {
		result.HasSpecial = true
		poolSize += 33 // Standard printable ASCII special character count
	}

	result.PoolSize = poolSize

	// Compute Shannon Entropy: E = L * log2(R)
	if poolSize > 0 && result.Length > 0 {
		result.Entropy = float64(result.Length) * math.Log2(float64(poolSize))
	}

	// Compute consecutive repeated characters
	maxConsecutive := 0
	if len(runes) > 0 {
		currentConsecutive := 1
		maxConsecutive = 1
		for i := 1; i < len(runes); i++ {
			if runes[i] == runes[i-1] {
				currentConsecutive++
				if currentConsecutive > maxConsecutive {
					maxConsecutive = currentConsecutive
				}
			} else {
				currentConsecutive = 1
			}
		}
	}
	result.MaxConsecutive = maxConsecutive

	// Check for sequential runs of length >= 3 (e.g., "123", "abc", "qwe")
	result.HasSequential = hasSequentialPatterns(password)

	writeJSONResponse(result)
}

// hasSequentialPatterns checks for simple alphabetic or numeric ascending/descending runs of length >= 3
func hasSequentialPatterns(s string) bool {
	runes := []rune(strings.ToLower(s))
	if len(runes) < 3 {
		return false
	}

	for i := 0; i < len(runes)-2; i++ {
		// Ascending sequence (e.g. 'a','b','c' or '1','2','3')
		if runes[i+1] == runes[i]+1 && runes[i+2] == runes[i]+2 {
			return true
		}
		// Descending sequence (e.g. '3','2','1' or 'c','b','a')
		if runes[i+1] == runes[i]-1 && runes[i+2] == runes[i]-2 {
			return true
		}
	}
	return false
}

// writeJSONResponse formats the result struct to JSON and writes it to stdout.
func writeJSONResponse(v interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(v)
}
