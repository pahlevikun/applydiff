package runtime

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pahlevikun/applydiff/internal/interview"
)

// GenerateInterviewKit builds and stores a kit for a job key.
func (a *App) GenerateInterviewKit(ctx context.Context, key string) (*interview.Kit, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, ErrNoMatch{Reason: err.Error()}
	}
	if _, err := a.Store.GetApplication(ctx, j.ID); err == sql.ErrNoRows {
		if _, err = a.Store.CreateApplication(ctx, j.ID); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	loop, topics, prompts, questions := interview.BuildKitJSON(j)
	kit := &interview.Kit{
		ApplicationID: j.ID,
		LoopJSON:      loop,
		TopicsJSON:    topics,
		PromptsJSON:   prompts,
		QuestionsJSON: questions,
	}
	if err := a.Store.SaveInterviewKit(ctx, kit); err != nil {
		return nil, err
	}
	return kit, nil
}

// StartInterviewMock creates a mock session with framework steps.
func (a *App) StartInterviewMock(ctx context.Context, key, mockType string) (*interview.MockSession, []interview.MockStep, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, nil, ErrNoMatch{Reason: err.Error()}
	}
	fw := interview.FrameworkByID(mockType)
	if fw == nil {
		return nil, nil, fmt.Errorf("interview: unknown mock type %q", mockType)
	}
	prompt := fmt.Sprintf("%s mock for %s — %s", fw.Name, j.Company, j.Title)
	sess := &interview.MockSession{
		ApplicationID: j.ID,
		Type:          fw.ID,
		Prompt:        prompt,
		Status:        "active",
	}
	steps := make([]interview.MockStep, len(fw.Steps))
	for i, st := range fw.Steps {
		steps[i] = interview.MockStep{
			Idx:      st.Idx,
			Name:     st.Name,
			TimeboxS: st.Minutes * 60,
		}
	}
	if err := a.Store.CreateMockSession(ctx, sess, steps); err != nil {
		return nil, nil, err
	}
	saved, err := a.Store.ListMockSteps(ctx, sess.ID)
	return sess, saved, err
}

// InterviewDrill returns topic note and drill for a slug.
func (a *App) InterviewDrill(ctx context.Context, topic string, minutes int) (*interview.TopicNote, *interview.Drill, error) {
	note, err := a.Store.GetTopicNote(ctx, topic)
	if err != nil {
		return nil, nil, fmt.Errorf("interview: unknown topic %q", topic)
	}
	drill, err := a.Store.GetDrillForTopic(ctx, topic)
	if err == sql.ErrNoRows {
		drill = &interview.Drill{TopicSlug: topic, Prompt: "Practice explaining " + note.Title + " with a concrete example from your experience.", Minutes: 15}
	} else if err != nil {
		return nil, nil, err
	}
	if minutes > 0 {
		drill.Minutes = minutes
	}
	return note, drill, nil
}

// AddInterviewStory validates and stores a story.
func (a *App) AddInterviewStory(ctx context.Context, key string, story interview.Story) (*interview.Story, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, ErrNoMatch{Reason: err.Error()}
	}
	story.ApplicationID = j.ID
	if story.EvidenceJSON == "" {
		story.EvidenceJSON = "[]"
	}
	facts := a.FactsOrFile(ctx)
	if err := interview.ValidateStory(facts, story); err != nil {
		return nil, err
	}
	if err := a.Store.SaveStory(ctx, &story); err != nil {
		return nil, err
	}
	return &story, nil
}

// ListInterviewStories returns stories for a job.
func (a *App) ListInterviewStories(ctx context.Context, key string) ([]interview.Story, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, ErrNoMatch{Reason: err.Error()}
	}
	return a.Store.ListStories(ctx, j.ID)
}

// ScoreInterviewSession records rubric scores for a mock session.
func (a *App) ScoreInterviewSession(ctx context.Context, sessionID string, scores map[string]int) ([]interview.RubricScore, error) {
	if _, err := a.Store.GetMockSession(ctx, sessionID); err != nil {
		return nil, err
	}
	var out []interview.RubricScore
	for dim, score := range scores {
		if score < 1 || score > 5 {
			return nil, fmt.Errorf("interview: score for %q must be 1-5", dim)
		}
		sc := interview.RubricScore{
			SessionID: sessionID,
			Dimension: dim,
			Score:     score,
		}
		if err := a.Store.SaveRubricScore(ctx, &sc); err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, nil
}

// ListInterviewSummary returns kits, sessions, and stories for a job.
func (a *App) ListInterviewSummary(ctx context.Context, key string) (*interview.ListSummary, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, ErrNoMatch{Reason: err.Error()}
	}
	kits, err := a.Store.ListInterviewKits(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	sessions, err := a.Store.ListMockSessions(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	stories, err := a.Store.ListStories(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	return &interview.ListSummary{
		ApplicationID: j.ID,
		Kits:          kits,
		Sessions:      sessions,
		Stories:       stories,
	}, nil
}

// ParseDimensionScores parses --dimension communication=4,depth=3.
func ParseDimensionScores(raw string) (map[string]int, error) {
	out := map[string]int{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out, nil
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid dimension pair %q", part)
		}
		name := strings.TrimSpace(kv[0])
		var score int
		if _, err := fmt.Sscanf(strings.TrimSpace(kv[1]), "%d", &score); err != nil {
			return nil, fmt.Errorf("invalid score in %q", part)
		}
		out[name] = score
	}
	return out, nil
}

// MockSessionJSON bundles session output for CLI.
type MockSessionJSON struct {
	Session interview.MockSession `json:"session"`
	Steps   []interview.MockStep  `json:"steps"`
}

// DrillJSON is CLI output for drill command.
type DrillJSON struct {
	Topic interview.TopicNote `json:"topic"`
	Drill interview.Drill     `json:"drill"`
}

// EncodeEvidenceJSON wraps fact keys as JSON array.
func EncodeEvidenceJSON(keys []string) string {
	if len(keys) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(keys)
	return string(b)
}
