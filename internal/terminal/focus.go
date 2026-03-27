package terminal

import (
	"fmt"
	"os/exec"
	"strings"
)

// FocusTabByPID finds the Terminal.app tab running the given PID and brings it to front.
func FocusTabByPID(pid int) error {
	// Get the TTY for this PID
	out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "tty=").Output()
	if err != nil {
		return fmt.Errorf("looking up TTY for PID %d: %w", pid, err)
	}

	tty := strings.TrimSpace(string(out))
	if tty == "" || tty == "??" {
		return fmt.Errorf("no TTY found for PID %d", pid)
	}

	// Ensure /dev/ prefix for matching against Terminal.app's tty property
	if !strings.HasPrefix(tty, "/dev/") {
		tty = "/dev/" + tty
	}

	// AppleScript to find and activate the matching Terminal tab
	script := fmt.Sprintf(`
tell application "Terminal"
	activate
	set targetTTY to %q
	repeat with w in windows
		set tabIndex to 0
		repeat with t in tabs of w
			set tabIndex to tabIndex + 1
			if tty of t is targetTTY then
				set selected tab of w to t
				set index of w to 1
				return
			end if
		end repeat
	end repeat
end tell
`, tty)

	cmd := exec.Command("osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("activating terminal tab: %w (%s)", err, string(out))
	}
	return nil
}
