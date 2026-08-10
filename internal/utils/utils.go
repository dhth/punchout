package utils

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"
)

func ExpandTilde(path, homeDir string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

func RightPadTrim(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return s[:length-3] + "..."
		}
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

func Trim(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return s[:length-3] + "..."
		}
		return s[:length]
	}
	return s
}

func HumanizeDuration(durationInSecs int) string {
	duration := time.Duration(durationInSecs) * time.Second

	if duration.Seconds() < 60 {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}

	if duration.Minutes() < 60 {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}

	modMins := int(math.Mod(duration.Minutes(), 60))
	hours := int(duration.Hours())

	if hours > 24 {
		days := hours / 24
		modHours := hours % 24

		if modHours == 0 {
			return fmt.Sprintf("%dd", days)
		}

		return fmt.Sprintf("%dd %dh", days, modHours)
	}

	if modMins == 0 {
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dh %dm", hours, modMins)
}
