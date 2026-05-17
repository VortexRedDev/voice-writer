package output

import (
	"time"

	"github.com/atotto/clipboard"
	"github.com/go-vgo/robotgo"
)

// Outputter handles text output using clipboard manipulation.
type Outputter struct{}

// NewOutputter creates a new Outputter.
func NewOutputter() *Outputter {
	return &Outputter{}
}

// TypeText simulates keyboard input for the given text.
// It uses clipboard pasting for reliable Chinese character support.
func (o *Outputter) TypeText(text string) error {
	if text == "" {
		return nil
	}

	// 1. Backup current clipboard
	original, _ := clipboard.ReadAll()
	
	// 2. Write text to clipboard
	err := clipboard.WriteAll(text)
	if err != nil {
		return err
	}

	// 3. Give it a moment to stabilize
	time.Sleep(50 * time.Millisecond)

	// 4. Robust Ctrl+V sequence
	robotgo.KeyToggle("control", "down")
	time.Sleep(50 * time.Millisecond)
	robotgo.KeyTap("v")
	time.Sleep(50 * time.Millisecond)
	robotgo.KeyToggle("control", "up")

	// 5. Short delay before restoring to avoid premature restoration
	time.Sleep(200 * time.Millisecond)

	// 6. Restore original clipboard
	clipboard.WriteAll(original)

	return nil
}
