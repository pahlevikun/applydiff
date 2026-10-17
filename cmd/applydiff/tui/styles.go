package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/pahlevikun/applydiff/internal/theme"
)

func lipColor(name string) lipgloss.Color {
	return lipgloss.Color(theme.MustColor(name, theme.Color("fg")))
}

func baseStyles() struct {
	bg, surface, surface2, line, fg, muted, dim, accent, add, del lipgloss.Style
} {
	return struct {
		bg, surface, surface2, line, fg, muted, dim, accent, add, del lipgloss.Style
	}{
		bg:       lipgloss.NewStyle().Background(lipColor("bg")),
		surface:  lipgloss.NewStyle().Background(lipColor("surface")).Foreground(lipColor("fg")),
		surface2: lipgloss.NewStyle().Background(lipColor("surface2")).Foreground(lipColor("fg")),
		line:     lipgloss.NewStyle().Foreground(lipColor("line")),
		fg:       lipgloss.NewStyle().Foreground(lipColor("fg")),
		muted:    lipgloss.NewStyle().Foreground(lipColor("muted")),
		dim:      lipgloss.NewStyle().Foreground(lipColor("dim")),
		accent:   lipgloss.NewStyle().Foreground(lipColor("accent")).Bold(true),
		add:      lipgloss.NewStyle().Foreground(lipColor("add")),
		del:      lipgloss.NewStyle().Foreground(lipColor("del")),
	}
}
