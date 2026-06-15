package main

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// NetworkResult represents the JSON output of the network threat hunter worker.
type NetworkResult struct {
	IsBreached  bool   `json:"is_breached"`
	BreachCount int    `json:"breach_count"`
	Error       string `json:"error,omitempty"`
}

// runNetworkAnalysis reads a password from standard input, performs a k-anonymity check
// against HaveIBeenPwned API, and writes the JSON serialized NetworkResult to stdout.
func runNetworkAnalysis() {
	var result NetworkResult

	// Read all input from stdin
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		result.Error = "failed to read from stdin: " + err.Error()
		writeJSONResponse(result)
		return
	}

	// Clean trailing newlines
	password := strings.TrimRight(string(inputBytes), "\r\n")
	if len(password) == 0 {
		writeJSONResponse(result)
		return
	}

	// 1. Compute SHA-1 hash locally
	hasher := sha1.New()
	hasher.Write([]byte(password))
	hashBytes := hasher.Sum(nil)
	hashStr := fmt.Sprintf("%X", hashBytes) // Uppercase hexadecimal

	if len(hashStr) < 5 {
		result.Error = "invalid sha1 hash length"
		writeJSONResponse(result)
		return
	}

	// 2. Extract prefix (5 bytes) and suffix
	prefix := hashStr[:5]
	suffix := hashStr[5:]

	// 3. Query HaveIBeenPwned API via k-Anonymity channel
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.pwnedpasswords.com/range/"+prefix, nil)
	if err != nil {
		result.Error = "failed to create HTTP request: " + err.Error()
		writeJSONResponse(result)
		return
	}

	// HIBP API recommends setting a User-Agent
	req.Header.Set("User-Agent", "pass-analy-cli-agent")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = "network request failed (offline?): " + err.Error()
		writeJSONResponse(result)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Sprintf("API returned status code %d", resp.StatusCode)
		writeJSONResponse(result)
		return
	}

	// 4. Parse matches and execute local suffix comparison
	scanner := bufio.NewScanner(resp.Body)
	matched := false
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			apiSuffix := parts[0]
			if strings.EqualFold(apiSuffix, suffix) {
				count, err := strconv.Atoi(parts[1])
				if err == nil {
					result.IsBreached = true
					result.BreachCount = count
					matched = true
					break
				}
			}
		}
	}

	if err := scanner.Err(); err != nil && !matched {
		result.Error = "failed to parse API response: " + err.Error()
	}

	writeJSONResponse(result)
}
