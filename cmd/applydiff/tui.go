package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Open the terminal UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := newTUIApp()
			p := tea.NewProgram(m, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				return err
			}
			return nil
		},
	}
}

type tuiModel struct {
	quitting bool
}

func newTUIApp() tuiModel {
	return tuiModel{}
}

func (m tuiModel) Init() tea.Cmd { return nil }

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m tuiModel) View() string {
	if m.quitting {
		return ""
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")).Render("ApplyDiff")
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(
		"Local job tailoring with diff approval.\n\n" +
			"  serve — API + embedded web UI\n" +
			"  build <job-key> — PDF to output/\n\n" +
			"Press q to quit.",
	)
	return title + "\n\n" + body + "\n"
}

func runCLI() error {
	fmt.Fprintln(os.Stderr, "ApplyDiff — run `applydiff serve` for the web UI or `applydiff tui` for this screen.")
	return nil
}
