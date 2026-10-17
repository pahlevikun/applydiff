package present

import (
	"fmt"
	"strings"

	"github.com/pahlevikun/applydiff/internal/tailor"
)

// FormatUnified renders hunks as unified diff text.
func FormatUnified(hunks []tailor.Hunk) string {
	var b strings.Builder
	for i, h := range hunks {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("--- hunk %d (%s)\n", i+1, h.FieldID))
		b.WriteString(fmt.Sprintf("+++ %s\n", h.Section))
		b.WriteString(fmt.Sprintf("@@ %s @@\n", h.Status))
		b.WriteString("-" + h.Before + "\n")
		b.WriteString("+" + h.After + "\n")
		if h.Rationale != "" {
			b.WriteString("# " + h.Rationale + "\n")
		}
	}
	return b.String()
}

// ANSI colors use terminal escapes; tokens are semantic names resolved by caller.
func ColorizeUnified(text string, addColor, delColor string, enable bool) string {
	if !enable {
		return text
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			lines[i] = addColor + line + "\033[0m"
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			lines[i] = delColor + line + "\033[0m"
		}
	}
	return strings.Join(lines, "\n")
}
