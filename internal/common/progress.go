package common

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var Silent = false
var Verbose = false

// ProgressIndicator provides visual feedback for long-running operations
type ProgressIndicator struct {
	message    string
	spinner    []string
	current    int
	startTime  time.Time
	isRunning  bool
	lastUpdate time.Time
}

// NewProgressIndicator creates a new progress indicator
func NewProgressIndicator(message string) *ProgressIndicator {
	return &ProgressIndicator{
		message:    message,
		spinner:    []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		current:    0,
		startTime:  time.Now(),
		isRunning:  false,
		lastUpdate: time.Now(),
	}
}
var Progress *ProgressIndicator

// Start begins the progress indicator
func (p *ProgressIndicator) Start() {
	if p.isRunning {
		return
	}
	p.isRunning = true
	p.startTime = time.Now()
	p.current = 0
	p.lastUpdate = time.Now()

	// Print initial message
	fmt.Fprintf(os.Stderr, "%s %s", p.spinner[p.current], p.message)
}

// Update updates the progress indicator with a new message
func (p *ProgressIndicator) Update(message string) {
	if !p.isRunning {
		return
	}

	// Only update if enough time has passed (to avoid flickering)
	if time.Since(p.lastUpdate) < 100*time.Millisecond {
		return
	}

	// Clear the line and update
	fmt.Fprintf(os.Stderr, "\r%s %s", p.spinner[p.current], message)
	p.lastUpdate = time.Now()
}

// Tick advances the spinner animation
func (p *ProgressIndicator) Tick() {
	if !p.isRunning {
		return
	}

	p.current = (p.current + 1) % len(p.spinner)

	// Only update if enough time has passed
	if time.Since(p.lastUpdate) < 100*time.Millisecond {
		return
	}

	fmt.Fprintf(os.Stderr, "\r%s %s", p.spinner[p.current], p.message)
	p.lastUpdate = time.Now()
}

// Stop stops the progress indicator with a completion message
func (p *ProgressIndicator) Stop(message string) {
	if !p.isRunning {
		return
	}

	duration := time.Since(p.startTime)

	// Clear the line and show completion message
	if message != "" {
		fmt.Fprintf(os.Stderr, "\r✅ %s (%.2fs)\n", message, duration.Seconds())
	} else {
		fmt.Fprintf(os.Stderr, "\r✅ %s (%.2fs)\n", p.message, duration.Seconds())
	}

	p.isRunning = false
}

// StopWithError stops the progress indicator with an error message
func (p *ProgressIndicator) StopWithError(message string) {
	if !p.isRunning {
		return
	}

	duration := time.Since(p.startTime)

	// Clear the line and show error message
	if message != "" {
		fmt.Fprintf(os.Stderr, "\r❌ %s (%.2fs)\n", message, duration.Seconds())
	} else {
		fmt.Fprintf(os.Stderr, "\r❌ %s (%.2fs)\n", p.message, duration.Seconds())
	}

	p.isRunning = false
}

// ProgressBar provides a visual progress bar for operations with known progress
type ProgressBar struct {
	message   string
	total     int
	current   int
	width     int
	isRunning bool
	startTime time.Time
}

// NewProgressBar creates a new progress bar
func NewProgressBar(message string, total int) *ProgressBar {
	return &ProgressBar{
		message:   message,
		total:     total,
		current:   0,
		width:     50,
		isRunning: false,
		startTime: time.Now(),
	}
}

// Start begins the progress bar
func (p *ProgressBar) Start() {
	if p.isRunning {
		return
	}
	p.isRunning = true
	p.startTime = time.Now()
	p.current = 0

	fmt.Fprintf(os.Stderr, "%s\n", p.message)
	p.update()
}

// Update updates the progress bar with current progress
func (p *ProgressBar) Update(current int) {
	if !p.isRunning {
		return
	}

	p.current = current
	if p.current > p.total {
		p.current = p.total
	}

	p.update()
}

// Increment advances the progress bar by one step
func (p *ProgressBar) Increment() {
	if !p.isRunning {
		return
	}

	p.current++
	if p.current > p.total {
		p.current = p.total
	}

	p.update()
}

// update renders the progress bar
func (p *ProgressBar) update() {
	percentage := float64(p.current) / float64(p.total)
	filled := int(float64(p.width) * percentage)

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	fmt.Fprintf(os.Stderr, "\r[%s] %d/%d (%.1f%%)", bar, p.current, p.total, percentage*100)
}

// Stop stops the progress bar with a completion message
func (p *ProgressBar) Stop(message string) {
	if !p.isRunning {
		return
	}

	duration := time.Since(p.startTime)

	// Complete the progress bar
	p.current = p.total
	p.update()

	// Show completion message
	if message != "" {
		fmt.Fprintf(os.Stderr, "\n✅ %s (%.2fs)\n", message, duration.Seconds())
	} else {
		fmt.Fprintf(os.Stderr, "\n✅ Completed (%.2fs)\n", duration.Seconds())
	}

	p.isRunning = false
}

// StopWithError stops the progress bar with an error message
func (p *ProgressBar) StopWithError(message string) {
	if !p.isRunning {
		return
	}

	duration := time.Since(p.startTime)

	// Show error message
	if message != "" {
		fmt.Fprintf(os.Stderr, "\n❌ %s (%.2fs)\n", message, duration.Seconds())
	} else {
		fmt.Fprintf(os.Stderr, "\n❌ Failed (%.2fs)\n", duration.Seconds())
	}

	p.isRunning = false
}

// ShouldShowProgress determines if progress indicators should be shown
func ShouldShowProgress() bool {
	// Don't show progress in silent mode
	if Silent {
		return false
	}

	// Don't show progress in test mode
	if os.Getenv("MAESTRO_K_TEST_MODE") == "true" {
		return false
	}

	// Don't show progress if output is redirected
	fileInfo, err := os.Stderr.Stat()
	if err != nil {
		return false
	}

	// Check if stderr is a terminal
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
