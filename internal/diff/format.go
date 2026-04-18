package diff

import (
	"fmt"
	"io"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorGray   = "\033[90m"
)

// FormatOptions controls output rendering.
type FormatOptions struct {
	Color   bool
	ShowUnchanged bool
}

// Render writes a human-readable diff to w.
func Render(w io.Writer, result *Result, opts FormatOptions) {
	for _, c := range result.Changes {
		if c.Type == Unchanged && !opts.ShowUnchanged {
			continue
		}
		line := formatChange(c, opts.Color)
		fmt.Fprintln(w, line)
	}
}

func formatChange(c Change, color bool) string {
	switch c.Type {
	case Added:
		msg := fmt.Sprintf("+ %s = %s", c.Key, c.NewValue)
		return colorize(msg, colorGreen, color)
	case Removed:
		msg := fmt.Sprintf("- %s = %s", c.Key, c.OldValue)
		return colorize(msg, colorRed, color)
	case Modified:
		lines := []string{
			colorize(fmt.Sprintf("~ %s", c.Key), colorYellow, color),
			colorize(fmt.Sprintf("  - %s", c.OldValue), colorRed, color),
			colorize(fmt.Sprintf("  + %s", c.NewValue), colorGreen, color),
		}
		return strings.Join(lines, "\n")
	case Unchanged:
		msg := fmt.Sprintf("  %s = %s", c.Key, c.OldValue)
		return colorize(msg, colorGray, color)
	}
	return ""
}

func colorize(s, code string, enabled bool) string {
	if !enabled {
		return s
	}
	return code + s + colorReset
}
