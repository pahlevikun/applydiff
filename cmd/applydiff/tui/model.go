package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pahlevikun/applydiff/internal/interview"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/pahlevikun/applydiff/internal/store"
)

type screen int

const (
	screenFeed      screen = 0
	screenJob       screen = 1
	screenDiff      screen = 2
	screenSrc       screen = 3
	screenApps      screen = 4
	screenInterview screen = 5
)

const screenCount = 6

type model struct {
	app      *runtime.App
	ctx      context.Context
	width    int
	height   int
	screen   screen
	help     bool
	filter   string
	jobs     []job.JobView
	presets  []store.FilterPreset
	cursor   int
	selected string
	detail   *runtime.JobDetail
	proposal *store.Proposal
	hunkIdx  int
	hunkSel  map[int]bool
	viewport viewport.Model
	status   string
	err      string
	ivMockType   string
	ivLoop       []string
	ivTopics     []string
	ivFramework  *interview.Framework
	ivStepIdx    int
	ivRemainSec  []int
	ivOutput     []string
	ivNote       []string
	ivWeak       []bool
	ivTimerOn    bool
	ivRubric     map[string]int
	ivRubricFocus int
	ivLoadErr    string
	styles   struct {
		bg, surface, surface2, line, fg, muted, dim, accent, add, del, amber lipgloss.Style
	}
}

func initialModel(app *runtime.App) model {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Background(lipColor("bg")).Foreground(lipColor("fg"))
	return model{
		app:     app,
		ctx:     context.Background(),
		screen:       screenFeed,
		hunkSel:        map[int]bool{},
		viewport:       vp,
		styles:         baseStyles(),
		status:         "Feed",
		ivMockType:     "system-design",
		ivRubric:       map[string]int{},
	}
}

func (m model) Init() tea.Cmd {
	return m.loadJobsCmd()
}

func (m model) loadJobsCmd() tea.Cmd {
	return func() tea.Msg {
		jobs, err := m.app.ListJobs(m.ctx)
		if err != nil {
			return errMsg{err: err}
		}
		presets, _ := m.app.Store.ListFilterPresets(m.ctx)
		return jobsMsg{jobs: jobs, presets: presets}
	}
}

type jobsMsg struct {
	jobs    []job.JobView
	presets []store.FilterPreset
}

type errMsg struct{ err error }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
		return m, nil
	case jobsMsg:
		m.jobs = filterJobs(msg.jobs, m.filter)
		m.presets = msg.presets
		if m.cursor >= len(m.jobs) {
			m.cursor = max(0, len(m.jobs)-1)
		}
		return m, nil
	case errMsg:
		m.err = msg.err.Error()
		return m, nil
	case ivLoadMsg:
		if msg.err != nil {
			m.ivLoadErr = msg.err.Error()
			m.ivFramework = nil
			return m, nil
		}
		m.ivLoadErr = ""
		m.ivLoop = msg.loop
		m.ivTopics = msg.topics
		m.ivFramework = msg.fw
		m.ensureIVSteps(msg.fw)
		return m, nil
	case tickMsg:
		if m.screen == screenInterview {
			return m.tickInterview()
		}
		return m, nil
	case tea.KeyMsg:
		if m.help {
			if msg.String() == "?" || msg.String() == "esc" {
				m.help = false
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.help = true
			return m, nil
		case "tab":
			m.screen = screen((int(m.screen) + 1) % screenCount)
			m.status = screenName(m.screen)
			if m.screen == screenInterview {
				return m, m.loadInterviewCmd()
			}
			return m, nil
		case "shift+tab":
			m.screen = screen((int(m.screen) + screenCount - 1) % screenCount)
			m.status = screenName(m.screen)
			if m.screen == screenInterview {
				return m, m.loadInterviewCmd()
			}
			return m, nil
		}
		switch m.screen {
		case screenFeed:
			return m.updateFeed(msg)
		case screenDiff:
			return m.updateDiff(msg)
		case screenInterview:
			return m.updateInterview(msg)
		default:
			return m, nil
		}
	}
	return m, nil
}

func (m model) updateFeed(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.jobs)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g":
		m.cursor = 0
	case "G":
		m.cursor = max(0, len(m.jobs)-1)
	case "enter", "o":
		if len(m.jobs) > 0 {
			m.selected = m.jobs[m.cursor].Job.ID
			detail, err := m.app.ShowJob(m.ctx, m.selected)
			if err != nil {
				m.err = err.Error()
				return m, nil
			}
			m.detail = detail
			m.screen = screenJob
			m.status = "Job"
		}
	case "/":
		m.filter = ""
		m.status = "filter: "
	case "backspace":
		if m.filter != "" {
			m.filter = m.filter[:len(m.filter)-1]
			m.jobs = filterJobs(m.reloadList(), m.filter)
		}
	default:
		if len(msg.Runes) > 0 && m.status == "filter: " {
			m.filter += string(msg.Runes)
			m.jobs = filterJobs(m.reloadList(), m.filter)
		}
	}
	return m, nil
}

func (m *model) reloadList() []job.JobView {
	jobs, err := m.app.ListJobs(m.ctx)
	if err != nil {
		return m.jobs
	}
	return jobs
}

func (m model) updateDiff(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.proposal == nil && m.selected != "" {
		p, err := m.app.LoadProposal(m.ctx, m.selected)
		if err == nil {
			m.proposal = p
		}
	}
	if m.proposal == nil || len(m.proposal.Hunks) == 0 {
		return m, nil
	}
	switch msg.String() {
	case "j", "down":
		if m.hunkIdx < len(m.proposal.Hunks)-1 {
			m.hunkIdx++
		}
	case "k", "up":
		if m.hunkIdx > 0 {
			m.hunkIdx--
		}
	case " ":
		m.hunkSel[m.hunkIdx] = !m.hunkSel[m.hunkIdx]
	case "a":
		_, err := m.app.ApproveHunks(m.ctx, m.selected, []int{m.hunkIdx + 1}, false, false)
		if err != nil {
			m.err = err.Error()
		} else {
			p, _ := m.app.LoadProposal(m.ctx, m.selected)
			m.proposal = p
		}
	case "x":
		_, err := m.app.RejectHunks(m.ctx, m.selected, []int{m.hunkIdx + 1}, false)
		if err != nil {
			m.err = err.Error()
		} else {
			p, _ := m.app.LoadProposal(m.ctx, m.selected)
			m.proposal = p
		}
	case "b":
		_, err := m.app.BuildPDF(m.ctx, m.selected, false)
		if err != nil {
			m.err = err.Error()
		} else {
			m.status = "PDF built"
		}
	}
	return m, nil
}

func filterJobs(jobs []job.JobView, q string) []job.JobView {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return jobs
	}
	var out []job.JobView
	for _, j := range jobs {
		hay := strings.ToLower(j.Job.Title + " " + j.Job.Company + " " + j.Job.Region)
		if strings.Contains(hay, q) {
			out = append(out, j)
		}
	}
	return out
}

func screenName(s screen) string {
	switch s {
	case screenFeed:
		return "Feed"
	case screenJob:
		return "Job"
	case screenDiff:
		return "Diff"
	case screenSrc:
		return "Sources"
	case screenApps:
		return "Applications"
	case screenInterview:
		return "Interview"
	default:
		return "ApplyDiff"
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// View renders the current frame (exported for snapshot tests).
func (m model) View() string {
	if m.help {
		return m.styles.surface.Render(m.helpOverlay())
	}
	var b strings.Builder
	header := m.styles.surface.Width(m.width).Render(
		m.styles.accent.Render("ApplyDiff") + "  " + m.styles.muted.Render(screenName(m.screen)),
	)
	b.WriteString(header + "\n")
	switch m.screen {
	case screenFeed:
		b.WriteString(m.viewFeed())
	case screenJob:
		b.WriteString(m.viewJob())
	case screenDiff:
		b.WriteString(m.viewDiff())
	case screenSrc:
		b.WriteString(m.viewSources())
	case screenApps:
		b.WriteString(m.viewApplications())
	case screenInterview:
		b.WriteString(m.viewInterview())
	}
	b.WriteString("\n")
	b.WriteString(m.styles.surface2.Width(m.width).Render(m.statusBar()))
	if m.err != "" {
		b.WriteString("\n" + m.styles.del.Render(m.err))
	}
	return m.styles.bg.Width(m.width).Height(m.height).Render(b.String())
}

func (m model) statusBar() string {
	hint := "j/k move · tab screen · / filter · ? help · q quit"
	return fmt.Sprintf(" %s · %s", m.status, hint)
}

func (m model) helpOverlay() string {
	return lipgloss.NewStyle().Padding(1, 2).Render(`ApplyDiff TUI — Phosphor
? close help
Feed: j/k g/G move · enter job · / filter
Diff: j/k hunk · space toggle · a accept · x reject · b build PDF
Interview: space timer · n/p step · d rubric focus · 1-5 score · enter save · r reload
tab / shift+tab: switch screen`)
}

func (m model) viewFeed() string {
	leftW := 24
	if m.width > 0 && m.width/4 > leftW {
		leftW = m.width / 4
	}
	var left strings.Builder
	left.WriteString(m.styles.dim.Render("Presets") + "\n")
	for _, p := range m.presets {
		left.WriteString("  " + p.Name + "\n")
	}
	if len(m.presets) == 0 {
		left.WriteString(m.styles.dim.Render("  (config)") + "\n")
	}
	var right strings.Builder
	right.WriteString(m.styles.dim.Render(fmt.Sprintf("%-36s %-14s %6s %5s", "Title", "Company", "Region", "Score")) + "\n")
	for i, j := range m.jobs {
		riskBadge := ""
		if j.Job.Risk > 0 {
			riskBadge = fmt.Sprintf(" R%d", j.Job.Risk)
			hard := false
			for _, f := range j.Job.Flags {
				if f.Severity == "hard" {
					hard = true
					break
				}
			}
			if hard {
				riskBadge = m.styles.del.Render(riskBadge)
			} else {
				riskBadge = m.styles.amber.Render(riskBadge)
			}
		}
		line := fmt.Sprintf("%-36s %-14s %6s %5d%s",
			trunc(j.Job.Title, 36), trunc(j.Job.Company, 14), trunc(j.Job.Region, 6), j.Filter.Score, riskBadge)
		if i == m.cursor {
			right.WriteString(m.styles.accent.Render("› "+line) + "\n")
		} else {
			right.WriteString("  " + line + "\n")
		}
	}
	if len(m.jobs) == 0 {
		right.WriteString(m.styles.muted.Render("  No jobs — run scan") + "\n")
	}
	panel := lipgloss.JoinHorizontal(lipgloss.Top,
		m.styles.surface2.Width(leftW).Padding(0, 1).Render(left.String()),
		m.styles.surface.Padding(0, 1).Render(right.String()),
	)
	return panel
}

func (m model) viewJob() string {
	if m.detail == nil {
		return m.styles.muted.Render("Select a job from Feed (enter)")
	}
	var b strings.Builder
	b.WriteString(m.styles.fg.Render(m.detail.Job.Title) + " @ " + m.detail.Job.Company + "\n")
	if m.detail.Job.Risk > 0 || len(m.detail.Job.Flags) > 0 {
		b.WriteString(m.styles.amber.Render(fmt.Sprintf("Risk %d", m.detail.Job.Risk)) + "\n")
		for _, f := range m.detail.Job.Flags {
			style := m.styles.amber
			if f.Severity == "hard" {
				style = m.styles.del
			}
			b.WriteString(style.Render(fmt.Sprintf("  [%s] %s", f.Severity, f.Message)) + "\n")
		}
	}
	if m.detail.Match != nil {
		b.WriteString(m.styles.muted.Render(fmt.Sprintf("Match %d/25", m.detail.Match.Total)) + "\n")
		for _, r := range m.detail.Match.Requirements {
			b.WriteString(fmt.Sprintf("  %s %s\n", r.Status, r.Phrase))
		}
	}
	b.WriteString("\n" + trunc(m.detail.Job.Description, 2000))
	return m.styles.surface.Padding(1).Render(b.String())
}

func (m model) viewDiff() string {
	if m.proposal == nil {
		return m.styles.muted.Render("Open a job with a proposal")
	}
	if len(m.proposal.Hunks) == 0 {
		return m.styles.muted.Render("No hunks")
	}
	h := m.proposal.Hunks[m.hunkIdx]
	var b strings.Builder
	b.WriteString(m.styles.accent.Render(fmt.Sprintf("Hunk %d/%d · %s", m.hunkIdx+1, len(m.proposal.Hunks), h.FieldID)) + "\n")
	b.WriteString(m.styles.del.Render("- "+h.Before) + "\n")
	b.WriteString(m.styles.add.Render("+ "+h.After) + "\n")
	b.WriteString(m.styles.dim.Render(h.Rationale))
	return m.styles.surface.Padding(1).Render(b.String())
}

func (m model) viewSources() string {
	view := m.app.Config.BuildView(m.ctx, func(ctx context.Context, key string) (string, error) {
		return m.app.Store.GetSetting(ctx, key)
	})
	var b strings.Builder
	for _, s := range view.Sources {
		state := "off"
		if s.Enabled {
			state = "ok"
		}
		b.WriteString(fmt.Sprintf("%s  %s  %s\n", s.ID, state, s.Hint))
	}
	return m.styles.surface.Padding(1).Render(b.String())
}

func (m model) viewApplications() string {
	items, err := m.app.Store.ListApplications(m.ctx)
	if err != nil {
		return m.styles.del.Render(err.Error())
	}
	var b strings.Builder
	for _, a := range items {
		b.WriteString(fmt.Sprintf("%s  %s @ %s  [%s]  %s\n", a.JobID, a.Title, a.Company, a.Status, a.Channel))
	}
	if len(items) == 0 {
		b.WriteString(m.styles.muted.Render("No applications yet"))
	}
	return m.styles.surface.Padding(1).Render(b.String())
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// Run starts the interactive TUI.
func Run(app *runtime.App) error {
	if app == nil {
		return fmt.Errorf("tui: app required")
	}
	m := initialModel(app)
	if !isInteractive() {
		if os.Getenv("APPLYDIFF_TUI_SNAPSHOT") != "" {
			m.screen = screenInterview
			m.status = "Interview"
			m.width = 100
			m.height = 24
			if fw := interview.FrameworkByID("system-design"); fw != nil {
				m.ivFramework = fw
				m.ivLoop = []string{"recruiter screen", "technical screen", "coding"}
				m.ivTopics = []string{"caching", "scalability"}
				m.ensureIVSteps(fw)
			}
		}
		fmt.Println(m.View())
		return nil
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func isInteractive() bool {
	return os.Getenv("APPLYDIFF_TUI_SNAPSHOT") == ""
}
