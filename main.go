package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func main() {
	// 1. Single-Binary Routing Logic: Check for child process worker flags
	if len(os.Args) > 1 {
		if os.Args[1] == "--mode-static" {
			runStaticAnalysis()
			return
		}
		if os.Args[1] == "--mode-network" {
			runNetworkAnalysis()
			return
		}
	}

	// 2. Determine target password string
	var password string
	if len(os.Args) > 1 {
		password = os.Args[1]
	}

	// If no password is provided on CLI, prompt the user securely
	if password == "" {
		fmt.Print("🔑 Enter password to analyze: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Printf("\n❌ Error reading password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println() // Add newline after password entry
		password = string(bytePassword)
	}

	// Trim whitespace/newlines
	password = strings.TrimRight(password, "\r\n")

	if len(password) == 0 {
		fmt.Println("❌ Error: Password cannot be empty.")
		os.Exit(1)
	}

	// 3. Initialize terminal UI tracking and layout engine loops
	runTUI(password)
}
