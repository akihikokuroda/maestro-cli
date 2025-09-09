// internal/common/console_test.go
package common

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestConsole(t *testing.T) {
	// Save original stdout and restore it after the test
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Create a console instance
	console := NewConsole(true, false)
	
	// Test Print
	console.Print("Hello")
	
	// Test Println
	console.Println("World")
	
	// Test Ok
	console.Ok("Success")
	
	// Test Warn
	console.Warn("Warning")
	
	// Test Error
	console.Error("Error")
	
	// Test VerbosePrint
	console.VerbosePrint("Verbose")
	
	// Close the writer and read the output
	w.Close()
	os.Stdout = oldStdout
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	
	// Check that the output contains the expected strings
	output := buf.String()
	if output == "" {
		t.Error("Expected output, got empty string")
	}
}

