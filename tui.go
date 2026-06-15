package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Msg types to handle worker results
type StaticDoneMsg struct {
	Result StaticResult
	Err    error
}

type NetworkDoneMsg struct {
	Result NetworkResult
	Err    error
}

// Bubble Tea model
type tuiModel struct {
	password       string
	spinner        spinner.Model
	staticRunning  bool
	staticRes      StaticResult
	staticErr      error
	networkRunning bool
	networkRes     NetworkResult
	networkErr     error
	quitting       bool
}

// runStaticCmd spawns the --mode-static subprocess and feeds the password
func runStaticCmd(password string) tea.Cmd {
	return func() tea.Msg {
		selfPath, err := os.Executable()
		if err != nil {
			return StaticDoneMsg{Err: err}
		}
		cmd := exec.Command(selfPath, "--mode-static")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return StaticDoneMsg{Err: err}
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return StaticDoneMsg{Err: err}
		}

		if err := cmd.Start(); err != nil {
			return StaticDoneMsg{Err: err}
		}

		// Write to stdin in a separate goroutine to prevent deadlocks
		go func() {
			_, _ = stdin.Write([]byte(password))
			_ = stdin.Close()
		}()

		outputBytes, err := io.ReadAll(stdout)
		if err != nil {
			return StaticDoneMsg{Err: err}
		}

		_ = cmd.Wait()

		var res StaticResult
		if err := json.Unmarshal(outputBytes, &res); err != nil {
			return StaticDoneMsg{Err: fmt.Errorf("JSON parse error: %v", err)}
		}

		return StaticDoneMsg{Result: res}
	}
}

// runNetworkCmd spawns the --mode-network subprocess and feeds the password
func runNetworkCmd(password string) tea.Cmd {
	return func() tea.Msg {
		selfPath, err := os.Executable()
		if err != nil {
			return NetworkDoneMsg{Err: err}
		}
		cmd := exec.Command(selfPath, "--mode-network")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return NetworkDoneMsg{Err: err}
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return NetworkDoneMsg{Err: err}
		}

		if err := cmd.Start(); err != nil {
			return NetworkDoneMsg{Err: err}
		}

		// Write to stdin in a separate goroutine to prevent deadlocks
		go func() {
			_, _ = stdin.Write([]byte(password))
			_ = stdin.Close()
		}()

		outputBytes, err := io.ReadAll(stdout)
		if err != nil {
			return NetworkDoneMsg{Err: err}
		}

		_ = cmd.Wait()

		var res NetworkResult
		if err := json.Unmarshal(outputBytes, &res); err != nil {
			return NetworkDoneMsg{Err: fmt.Errorf("JSON parse error: %v", err)}
		}

		return NetworkDoneMsg{Result: res}
	}
}

// Init function initializes spinners and workers
func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		runStaticCmd(m.password),
		runNetworkCmd(m.password),
	)
}

// Update handles state changes
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case StaticDoneMsg:
		m.staticRunning = false
		m.staticRes = msg.Result
		m.staticErr = msg.Err
		return m, nil
	case NetworkDoneMsg:
		m.networkRunning = false
		m.networkRes = msg.Result
		m.networkErr = msg.Err
		return m, nil
	}
	return m, nil
}

// Styles definition
var (
	purpleBorderColor = lipgloss.Color("#6366f1") // Violet/indigo
	cyanColor         = lipgloss.Color("#06b6d4") // Cyan
	greenColor        = lipgloss.Color("#10b981") // Emerald green
	orangeColor       = lipgloss.Color("#f59e0b") // Amber orange
	redColor          = lipgloss.Color("#ef4444") // Red
	mutedColor        = lipgloss.Color("#6b7280") // Gray
	whiteColor        = lipgloss.Color("#f9fafb") // Off-white

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteColor).
			Background(purpleBorderColor).
			Padding(0, 2).
			MarginLeft(1).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purpleBorderColor).
			Padding(1, 2).
			Width(44).
			Height(14)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(cyanColor).
			MarginBottom(1)

	metricLabelStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	metricValueStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Bold(true)

	checkMark = lipgloss.NewStyle().Foreground(greenColor).Render("✔")
	crossMark = lipgloss.NewStyle().Foreground(redColor).Render("✖")
	warnMark  = lipgloss.NewStyle().Foreground(orangeColor).Render("⚠")
)

// View renders the terminal output
func (m tuiModel) View() string {
	if m.quitting {
		return "\n  Goodbye! Keep your passwords secure. 🔒\n\n"
	}

	// 1. Header Title
	header := titleStyle.Render("⚡ UNIVERSAL PASSWORD STRENGTH ANALYZER ⚡")

	// Mask the password for the UI
	maskedPass := strings.Repeat("*", len([]rune(m.password)))
	if len(m.password) > 20 {
		maskedPass = strings.Repeat("*", 20) + "..."
	}
	passDisplay := fmt.Sprintf("  Target: %s (%d chars)\n\n", metricValueStyle.Render(maskedPass), len([]rune(m.password)))

	// 2. Build Left Box (Static Geometry Audit)
	var leftBody string
	if m.staticRunning {
		leftBody = fmt.Sprintf("\n  %s Analyzing password geometry...", m.spinner.View())
	} else if m.staticErr != nil {
		leftBody = fmt.Sprintf("\n  %s Error: %s", crossMark, m.staticErr.Error())
	} else {
		res := m.staticRes
		// Entropy category & color
		var strengthText string
		var strengthBar string
		var strengthStyle lipgloss.Style

		switch {
		case res.Entropy < 40.0:
			strengthText = "VERY WEAK"
			strengthStyle = lipgloss.NewStyle().Foreground(redColor).Bold(true)
			strengthBar = fmt.Sprintf("[##%s]", strings.Repeat(".", 8))
		case res.Entropy >= 40.0 && res.Entropy < 60.0:
			strengthText = "WEAK"
			strengthStyle = lipgloss.NewStyle().Foreground(orangeColor).Bold(true)
			strengthBar = fmt.Sprintf("[####%s]", strings.Repeat(".", 6))
		case res.Entropy >= 60.0 && res.Entropy < 80.0:
			strengthText = "STRONG"
			strengthStyle = lipgloss.NewStyle().Foreground(greenColor).Bold(true)
			strengthBar = fmt.Sprintf("[########%s]", strings.Repeat(".", 2))
		default:
			strengthText = "VERY STRONG"
			strengthStyle = lipgloss.NewStyle().Foreground(cyanColor).Bold(true)
			strengthBar = "[##########]"
		}

		charTypes := []string{}
		if res.HasLowercase {
			charTypes = append(charTypes, fmt.Sprintf("%s a-z", checkMark))
		} else {
			charTypes = append(charTypes, fmt.Sprintf("%s a-z", crossMark))
		}
		if res.HasUppercase {
			charTypes = append(charTypes, fmt.Sprintf("%s A-Z", checkMark))
		} else {
			charTypes = append(charTypes, fmt.Sprintf("%s A-Z", crossMark))
		}
		if res.HasDigit {
			charTypes = append(charTypes, fmt.Sprintf("%s 0-9", checkMark))
		} else {
			charTypes = append(charTypes, fmt.Sprintf("%s 0-9", crossMark))
		}
		if res.HasSpecial {
			charTypes = append(charTypes, fmt.Sprintf("%s Sym", checkMark))
		} else {
			charTypes = append(charTypes, fmt.Sprintf("%s Sym", crossMark))
		}

		consecText := fmt.Sprintf("%d (Excellent)", res.MaxConsecutive)
		if res.MaxConsecutive > 2 {
			consecText = fmt.Sprintf("%d (Repeated)", res.MaxConsecutive)
		}

		seqText := "Clean"
		if res.HasSequential {
			seqText = "Warning (Runs found)"
		}

		leftBody = fmt.Sprintf(
			"Strength: %s\n"+
				"Bar:      %s\n"+
				"Entropy:  %s bits\n"+
				"Pool:     %s characters\n\n"+
				"Character Diversity:\n"+
				"  %s\n\n"+
				"Consecutive: %s\n"+
				"Sequences:   %s",
			strengthStyle.Render(strengthText),
			strengthStyle.Render(strengthBar),
			metricValueStyle.Render(fmt.Sprintf("%.2f", res.Entropy)),
			metricValueStyle.Render(fmt.Sprintf("%d", res.PoolSize)),
			strings.Join(charTypes, "  "),
			metricValueStyle.Render(consecText),
			metricValueStyle.Render(seqText),
		)
	}
	leftBox := boxStyle.Copy().Render(
		headerStyle.Render("📊 STATIC GEOMETRY AUDIT") + "\n" + leftBody,
	)

	// 3. Build Right Box (Network Threat Audit)
	var rightBody string
	if m.networkRunning {
		rightBody = fmt.Sprintf("\n  %s Searching threat databases...", m.spinner.View())
	} else if m.networkErr != nil {
		rightBody = fmt.Sprintf("\n  %s Threat Hunt Offline:\n  %s\n\n  Check your connection.", warnMark, m.networkErr.Error())
	} else {
		res := m.networkRes
		if res.Error != "" {
			rightBody = fmt.Sprintf("\n  %s Error: %s", crossMark, res.Error)
		} else if res.IsBreached {
			breachText := fmt.Sprintf("%d times", res.BreachCount)
			warningText := "🚨 CRITICAL BREACH IDENTIFIED 🚨"
			rightBody = fmt.Sprintf(
				"\n"+
					"Status:  %s\n\n"+
					"Leaks:   %s\n\n"+
					"Warning:\n"+
					"This credential was found in public leaks.\n"+
					"DO NOT use this password anywhere!",
				lipgloss.NewStyle().Foreground(redColor).Bold(true).Render("BREACHED"),
				lipgloss.NewStyle().Foreground(redColor).Bold(true).Render(breachText),
			)
			// Pad top warning
			rightBody = lipgloss.NewStyle().Foreground(redColor).Bold(true).Render(warningText) + "\n" + rightBody
		} else {
			rightBody = fmt.Sprintf(
				"\n"+
					"Status:  %s\n\n"+
					"Leaks:   %s\n\n"+
					"Info:\n"+
					"This credential does not appear in any\n"+
					"indexed public database leaks.\n"+
					"It is currently safe from network threat vectors.",
				lipgloss.NewStyle().Foreground(greenColor).Bold(true).Render("SECURE"),
				lipgloss.NewStyle().Foreground(greenColor).Bold(true).Render("0 leaks"),
			)
		}
	}
	rightBox := boxStyle.Copy().Render(
		headerStyle.Render("🌐 BREACH DATABASE CHECK") + "\n" + rightBody,
	)

	// 4. Combine boxes side-by-side
	grid := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	// 5. Help / Footer
	footer := metricLabelStyle.Render("\n  [Press ESC / Q / Ctrl+C to Exit]")

	return header + "\n" + passDisplay + grid + footer + "\n"
}

// runTUI configures and runs the Bubble Tea CLI interface
func runTUI(password string) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(cyanColor)

	m := tuiModel{
		password:       password,
		spinner:        s,
		staticRunning:  true,
		networkRunning: true,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
