// internal/common/console.go
package common

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	// Color definitions
	headerColor  = color.New(color.FgMagenta)
	okColor      = color.New(color.FgGreen)
	warningColor = color.New(color.FgYellow)
	errorColor   = color.New(color.FgRed)
	boldColor    = color.New(color.Bold)
)

// Console provides methods for formatted console output
type Console struct {
	Verbose bool
	Silent  bool
}

// NewConsole creates a new Console instance
func NewConsole(verbose, silent bool) *Console {
	return &Console{
		Verbose: verbose,
		Silent:  silent,
	}
}

// Print prints a message to the console
func (c *Console) Print(msg string) {
	fmt.Print(msg)
}

// Println prints a message to the console with a newline
func (c *Console) Println(msg string) {
	fmt.Println(msg)
}

// Ok prints a success message in green
func (c *Console) Ok(msg string) {
	if !c.Silent {
		okColor.Println(msg)
	}
}

// Warn prints a warning message in yellow
func (c *Console) Warn(msg string) {
	warningColor.Println("Warning: " + msg)
}

// Error prints an error message in red
func (c *Console) Error(msg string) {
	errorColor.Println("Error: " + msg)
}

// VerbosePrint prints a message only if verbose mode is enabled
func (c *Console) VerbosePrint(msg string) {
	if c.Verbose {
		headerColor.Println(msg)
	}
}

// ReadInput reads input from the console with a prompt
func (c *Console) ReadInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return input
}

// Progress displays a progress bar
func (c *Console) Progress(count, total int, status string) {
	if c.Silent {
		return
	}
	
	barLen := 60
	filledLen := int(float64(barLen) * float64(count) / float64(total))
	
	bar := ""
	for i := 0; i < filledLen; i++ {
		bar += "="
	}
	for i := filledLen; i < barLen; i++ {
		bar += "-"
	}
	
	percent := float64(count) / float64(total) * 100
	fmt.Printf("\r[%s] %.1f%% ...%s", bar, percent, status)
	
	if count == total {
		fmt.Println()
	}
}

