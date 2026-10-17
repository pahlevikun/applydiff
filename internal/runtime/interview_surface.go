package runtime

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pahlevikun/applydiff/internal/interview"
)

// InterviewWorkspace loads kit, stories, and latest mock for a job key.
func (a *App) InterviewWorkspace(ctx context.Context, key string, mockType string) (*interview.Workspace, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, ErrNoMatch{Reason: err.Error()}
	}
	if mockType == "" {
		mockType = "system-design"
	}
	fw := interview.FrameworkByID(mockType)
	if fw == nil {
		return nil, fmt.Errorf("interview: unknown mock type %q", mockType)
	}
	var kit *interview.Kit
	kits, err := a.Store.ListInterviewKits(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	if len(kits) > 0 {
		kit = &kits[0]
	}
	stories, err := a.Store.ListStories(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	ws := &interview.Workspace{
		ApplicationID:   j.ID,
		Kit:             kit,
		Stories:         stories,
		Categories:      interview.StoryCategories,
		EmptyCategories: interview.EmptyStoryCategories(stories),
		MockType:        fw.ID,
		Framework:       fw,
	}
	if kit != nil {
		ws.Loop = interview.ParseStringListJSON(kit.LoopJSON)
		ws.Topics = interview.ParseStringListJSON(kit.TopicsJSON)
		ws.Questions = interview.ParseStringListJSON(kit.QuestionsJSON)
	}
	sessions, err := a.Store.ListMockSessions(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	for _, sess := range sessions {
		if sess.Type == fw.ID {
			ws.LatestSession = &sess
			steps, err := a.Store.ListMockSteps(ctx, sess.ID)
			if err != nil {
				return nil, err
			}
			ws.LatestSteps = steps
			ws.LatestRubric, _ = a.Store.ListRubricScores(ctx, sess.ID)
			break
		}
	}
	return ws, nil
}

// SaveInterviewMock persists a completed mock session with steps and rubric.
func (a *App) SaveInterviewMock(ctx context.Context, key, mockType string, steps []interview.MockStep, scores map[string]int, comments map[string]string) (*interview.MockSession, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, ErrNoMatch{Reason: err.Error()}
	}
	fw := interview.FrameworkByID(mockType)
	if fw == nil {
		return nil, fmt.Errorf("interview: unknown mock type %q", mockType)
	}
	if _, err := a.Store.GetApplication(ctx, j.ID); err == sql.ErrNoRows {
		if _, err = a.Store.CreateApplication(ctx, j.ID); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	prompt := fmt.Sprintf("%s mock for %s — %s", fw.Name, j.Company, j.Title)
	sess := &interview.MockSession{
		ApplicationID: j.ID,
		Type:          fw.ID,
		Prompt:        prompt,
		Status:        "completed",
	}
	now := time.Now().UTC().Format(time.RFC3339)
	sess.FinishedAt = now
	for i := range steps {
		if steps[i].Output == "" && steps[i].TimeboxS > 0 {
			steps[i].Weak = true
		}
	}
	if err := a.Store.CreateMockSession(ctx, sess, steps); err != nil {
		return nil, err
	}
	if err := a.Store.FinishMockSession(ctx, sess.ID, "completed", now); err != nil {
		return nil, err
	}
	var rubric []interview.RubricScore
	for dim, score := range scores {
		if score < 1 || score > 5 {
			continue
		}
		sc := interview.RubricScore{
			SessionID: sess.ID,
			Dimension: dim,
			Score:     score,
			Comment:   comments[dim],
		}
		if err := a.Store.SaveRubricScore(ctx, &sc); err != nil {
			return nil, err
		}
		rubric = append(rubric, sc)
	}
	_ = rubric
	return sess, nil
}

// UpdateInterviewStory updates an existing story with validation.
func (a *App) UpdateInterviewStory(ctx context.Context, storyID string, story interview.Story) (*interview.Story, error) {
	existing, err := a.Store.GetStory(ctx, storyID)
	if err != nil {
		return nil, err
	}
	story.ID = storyID
	story.ApplicationID = existing.ApplicationID
	if story.EvidenceJSON == "" {
		story.EvidenceJSON = existing.EvidenceJSON
	}
	facts := a.FactsOrFile(ctx)
	if err := interview.ValidateStory(facts, story); err != nil {
		return nil, err
	}
	if err := a.Store.UpdateStory(ctx, &story); err != nil {
		return nil, err
	}
	return &story, nil
}

// InterviewExportResult is CLI/API output for export.
type InterviewExportResult struct {
	Markdown string `json:"markdown"`
	PDF      string `json:"pdf,omitempty"`
}

// ExportInterviewPrep writes markdown (and optional PDF) to output/.
func (a *App) ExportInterviewPrep(ctx context.Context, key string, withPDF bool) (*InterviewExportResult, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, ErrNoMatch{Reason: err.Error()}
	}
	kits, _ := a.Store.ListInterviewKits(ctx, j.ID)
	if len(kits) == 0 {
		return nil, fmt.Errorf("interview export: no kit for job %s — run interview kit first", key)
	}
	kit := kits[0]
	loop := interview.ParseStringListJSON(kit.LoopJSON)
	topics := interview.ParseStringListJSON(kit.TopicsJSON)
	questions := interview.ParseStringListJSON(kit.QuestionsJSON)
	stories, err := a.Store.ListStories(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	sessions, err := a.Store.ListMockSessions(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	var sessExports []interview.PrepSessionExport
	var weak []interview.WeakStep
	for _, sess := range sessions {
		steps, err := a.Store.ListMockSteps(ctx, sess.ID)
		if err != nil {
			return nil, err
		}
		scores, err := a.Store.ListRubricScores(ctx, sess.ID)
		if err != nil {
			return nil, err
		}
		sessExports = append(sessExports, interview.PrepSessionExport{
			Session: sess,
			Steps:   steps,
			Scores:  scores,
		})
		for _, st := range steps {
			if st.Weak {
				weak = append(weak, interview.WeakStep{
					SessionType: sess.Type,
					StepName:    st.Name,
					Note:        st.Note,
				})
			}
		}
	}
	doc := interview.PrepDocument{
		Job:       *j,
		Loop:      loop,
		Topics:    topics,
		Questions: questions,
		Stories:   stories,
		Sessions:  sessExports,
		WeakSteps: weak,
	}
	md := interview.RenderPrepMarkdown(doc)
	outRoot := a.OutputRoot()
	at := time.Now().UTC()
	mdPath := interview.PrepOutputPath(outRoot, *j, at)
	if err := interview.WritePrepMarkdown(mdPath, md); err != nil {
		return nil, err
	}
	res := &InterviewExportResult{Markdown: mdPath}
	if withPDF {
		pdfPath := interview.PrepPDFPath(mdPath)
		if err := interview.WritePrepPDF(md, pdfPath); err != nil {
			return nil, err
		}
		res.PDF = pdfPath
	}
	return res, nil
}
