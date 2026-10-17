package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pahlevikun/applydiff/internal/interview"
)

type ivLoadMsg struct {
	loop   []string
	topics []string
	fw     *interview.Framework
	err    error
}

type tickMsg time.Time

func (m model) loadInterviewCmd() tea.Cmd {
	return func() tea.Msg {
		key := m.interviewJobKey()
		if key == "" {
			return ivLoadMsg{err: fmt.Errorf("select a job on Feed (enter) or move cursor to a job")}
		}
		ws, err := m.app.InterviewWorkspace(m.ctx, key, m.ivMockType)
		if err != nil {
			return ivLoadMsg{err: err}
		}
		if ws.Framework == nil {
			return ivLoadMsg{err: fmt.Errorf("unknown mock type")}
		}
		return ivLoadMsg{loop: ws.Loop, topics: ws.Topics, fw: ws.Framework}
	}
}

func (m *model) interviewJobKey() string {
	if m.selected != "" {
		return m.selected
	}
	if len(m.jobs) > 0 && m.cursor < len(m.jobs) {
		return m.jobs[m.cursor].Job.ID
	}
	return ""
}

func (m *model) ensureIVSteps(fw *interview.Framework) {
	if fw == nil {
		return
	}
	n := len(fw.Steps)
	if len(m.ivRemainSec) == n {
		return
	}
	m.ivRemainSec = make([]int, n)
	m.ivOutput = make([]string, n)
	m.ivNote = make([]string, n)
	m.ivWeak = make([]bool, n)
	for i, st := range fw.Steps {
		m.ivRemainSec[i] = st.Minutes * 60
	}
	if m.ivRubric == nil {
		m.ivRubric = map[string]int{}
	}
	if m.ivStepIdx >= n {
		m.ivStepIdx = 0
	}
}

func (m model) updateInterview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ivFramework == nil {
		if msg.String() == "r" {
			return m, m.loadInterviewCmd()
		}
		return m, nil
	}
	switch msg.String() {
	case " ":
		m.ivTimerOn = !m.ivTimerOn
		if m.ivTimerOn {
			return m, tickEvery()
		}
		return m, nil
	case "n":
		if m.ivStepIdx < len(m.ivFramework.Steps)-1 {
			m.ivStepIdx++
			m.ivTimerOn = false
		}
	case "p":
		if m.ivStepIdx > 0 {
			m.ivStepIdx--
			m.ivTimerOn = false
		}
	case "d":
		if len(m.ivFramework.RubricDims) > 0 {
			m.ivRubricFocus = (m.ivRubricFocus + 1) % len(m.ivFramework.RubricDims)
		}
	case "enter":
		return m.saveInterviewSession()
	case "1", "2", "3", "4", "5":
		if len(m.ivFramework.RubricDims) == 0 {
			return m, nil
		}
		score := int(msg.String()[0] - '0')
		dim := m.ivFramework.RubricDims[m.ivRubricFocus]
		m.ivRubric[dim] = score
		m.status = fmt.Sprintf("rubric %s = %d", dim, score)
	}
	return m, nil
}

func (m model) saveInterviewSession() (tea.Model, tea.Cmd) {
	key := m.interviewJobKey()
	if key == "" || m.ivFramework == nil {
		m.err = "no job for interview save"
		return m, nil
	}
	steps := make([]interview.MockStep, len(m.ivFramework.Steps))
	for i, st := range m.ivFramework.Steps {
		weak := m.ivWeak[i]
		if strings.TrimSpace(m.ivOutput[i]) == "" && m.ivRemainSec[i] <= 0 {
			weak = true
		}
		steps[i] = interview.MockStep{
			Idx:      st.Idx,
			Name:     st.Name,
			TimeboxS: st.Minutes * 60,
			Output:   m.ivOutput[i],
			Note:     m.ivNote[i],
			Weak:     weak,
		}
	}
	_, err := m.app.SaveInterviewMock(m.ctx, key, m.ivFramework.ID, steps, m.ivRubric, nil)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.status = "Interview session saved"
	m.ivTimerOn = false
	return m, nil
}

func tickEvery() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) tickInterview() (tea.Model, tea.Cmd) {
	if !m.ivTimerOn || m.ivFramework == nil {
		return m, nil
	}
	i := m.ivStepIdx
	if i < len(m.ivRemainSec) && m.ivRemainSec[i] > 0 {
		m.ivRemainSec[i]--
		if m.ivRemainSec[i] == 0 && strings.TrimSpace(m.ivOutput[i]) == "" {
			m.ivWeak[i] = true
		}
	}
	return m, tickEvery()
}

func formatMMSS(sec int) string {
	if sec < 0 {
		sec = 0
	}
	return fmt.Sprintf("%02d:%02d", sec/60, sec%60)
}

func (m model) viewInterview() string {
	if m.ivFramework == nil {
		msg := "Interview — load a job (cursor on Feed or open a job), then press r to reload"
		if m.ivLoadErr != "" {
			msg = m.ivLoadErr
		}
		return m.styles.surface.Padding(1).Render(m.styles.muted.Render(msg))
	}
	leftW := 32
	if m.width > 0 && m.width/3 > leftW {
		leftW = m.width / 3
	}
	var left strings.Builder
	left.WriteString(m.styles.dim.Render("Loop") + "\n")
	for i, s := range m.ivLoop {
		left.WriteString(fmt.Sprintf("  %d. %s\n", i+1, trunc(s, 28)))
	}
	if len(m.ivLoop) == 0 {
		left.WriteString(m.styles.muted.Render("  (no kit)\n"))
	}
	left.WriteString("\n" + m.styles.dim.Render("Topics") + "\n")
	for _, t := range m.ivTopics {
		left.WriteString("  · " + trunc(t, 26) + "\n")
	}
	left.WriteString("\n" + m.styles.dim.Render("Steps") + "\n")
	for i, st := range m.ivFramework.Steps {
		line := fmt.Sprintf("%s %s", formatMMSS(m.ivRemainSec[i]), trunc(st.Name, 18))
		if m.ivWeak[i] {
			line = m.styles.muted.Render("⚠ ") + m.styles.muted.Render(line)
		} else if i == m.ivStepIdx {
			line = m.styles.accent.Render("› " + line)
		} else {
			line = "  " + line
		}
		left.WriteString(line + "\n")
	}

	cur := m.ivFramework.Steps[m.ivStepIdx]
	var right strings.Builder
	timerState := "paused"
	if m.ivTimerOn {
		timerState = "running"
	}
	right.WriteString(m.styles.accent.Render(cur.Name) + "  ")
	right.WriteString(m.styles.muted.Render(formatMMSS(m.ivRemainSec[m.ivStepIdx])+" · "+timerState) + "\n\n")
	right.WriteString(m.styles.dim.Render("Required output") + "\n")
	right.WriteString(cur.RequiredOutput + "\n\n")
	right.WriteString(m.styles.dim.Render("Probes") + "\n")
	for _, p := range cur.Probes {
		right.WriteString("  · " + p + "\n")
	}
	if m.ivWeak[m.ivStepIdx] {
		right.WriteString("\n" + m.styles.muted.Render("Weak step: time expired with no output recorded") + "\n")
	}
	right.WriteString("\n" + m.styles.dim.Render("Rubric (d focus, 1-5 score)") + "\n")
	for i, dim := range m.ivFramework.RubricDims {
		sc := m.ivRubric[dim]
		line := fmt.Sprintf("  %s: %d/5", dim, sc)
		if i == m.ivRubricFocus {
			line = m.styles.accent.Render("› " + line)
		}
		right.WriteString(line + "\n")
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.styles.surface2.Width(leftW).Padding(0, 1).Render(left.String()),
		m.styles.surface.Padding(0, 1).Render(right.String()),
	)
}
