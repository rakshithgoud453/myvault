// Package clipboard provides cross-platform clipboard operations
// with automatic clearing after a timeout.
//
// macOS: uses pbcopy/pbpaste
// Linux: will use xclip/xsel (future)
// Windows: will use clip.exe (future)
package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// DefaultClearTimeout is how long the password stays on the clipboard.
const DefaultClearTimeout = 30 * time.Second

// Copy places text on the system clipboard.
func Copy(text string) error {
	cmd, err := clipboardCmd()
	if err != nil {
		return err
	}

	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdin = strings.NewReader(text)

	if err := c.Run(); err != nil {
		return fmt.Errorf("copying to clipboard: %w", err)
	}

	return nil
}

// Clear empties the system clipboard.
func Clear() error {
	return Copy("")
}

// CopyAndClear copies text to the clipboard and spawns a detached background
// process that clears it after the given timeout. The CLI exits immediately —
// the clear happens independently via an OS-level process.
func CopyAndClear(text string, timeout time.Duration) error {
	if err := Copy(text); err != nil {
		return err
	}

	// Spawn a detached process: sleep N seconds, then write empty string to clipboard.
	seconds := int(timeout.Seconds())
	clearCmd, err := clearAfterCmd(seconds)
	if err != nil {
		return err
	}

	cmd := exec.Command(clearCmd[0], clearCmd[1:]...)
	// Start the process detached — don't wait for it.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting clipboard clear timer: %w", err)
	}
	// Release the child process so it survives after we exit.
	_ = cmd.Process.Release()

	return nil
}

// clipboardCmd returns the OS-specific command for writing to the clipboard.
func clipboardCmd() ([]string, error) {
	switch runtime.GOOS {
	case "darwin":
		return []string{"pbcopy"}, nil
	case "linux":
		// Try xclip first, fall back to xsel.
		if path, err := exec.LookPath("xclip"); err == nil {
			return []string{path, "-selection", "clipboard"}, nil
		}
		if path, err := exec.LookPath("xsel"); err == nil {
			return []string{path, "--clipboard", "--input"}, nil
		}
		return nil, fmt.Errorf("clipboard requires xclip or xsel on Linux")
	default:
		return nil, fmt.Errorf("clipboard not supported on %s yet", runtime.GOOS)
	}
}

// clearAfterCmd returns an OS command that sleeps then clears the clipboard.
func clearAfterCmd(seconds int) ([]string, error) {
	switch runtime.GOOS {
	case "darwin":
		// sh -c 'sleep 30 && echo -n "" | pbcopy'
		return []string{"sh", "-c", fmt.Sprintf("sleep %d && echo -n '' | pbcopy", seconds)}, nil
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			return []string{"sh", "-c", fmt.Sprintf("sleep %d && echo -n '' | xclip -selection clipboard", seconds)}, nil
		}
		if _, err := exec.LookPath("xsel"); err == nil {
			return []string{"sh", "-c", fmt.Sprintf("sleep %d && echo -n '' | xsel --clipboard --input", seconds)}, nil
		}
		return nil, fmt.Errorf("clipboard requires xclip or xsel on Linux")
	default:
		return nil, fmt.Errorf("clipboard clear not supported on %s yet", runtime.GOOS)
	}
}
