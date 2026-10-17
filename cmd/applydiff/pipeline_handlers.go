package main

import (
	"net/http"
)

type pipelineJob struct {
	JobID    string `json:"job_id"`
	Company  string `json:"company"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	JobFeed  string `json:"job_feed_status"`
	HasDiff  bool   `json:"has_proposal"`
}

type pipelineView struct {
	Matched     []pipelineJob `json:"matched"`
	Shortlisted []pipelineJob `json:"shortlisted"`
	Tailoring   []pipelineJob `json:"tailoring"`
	Ready       []pipelineJob `json:"ready"`
	WithDiff    []pipelineJob `json:"with_proposal"`
}

func (s *server) handlePipeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	items, err := s.app.Store.ListApplications(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := pipelineView{
		Matched:     []pipelineJob{},
		Shortlisted: []pipelineJob{},
		Tailoring:   []pipelineJob{},
		Ready:       []pipelineJob{},
		WithDiff:    []pipelineJob{},
	}
	for _, item := range items {
		j, err := s.app.Store.GetJob(r.Context(), item.JobID)
		jobFeed := ""
		if err == nil {
			jobFeed = j.Status
		}
		_, propErr := s.app.Store.LoadProposal(r.Context(), item.JobID)
		hasDiff := propErr == nil
		pj := pipelineJob{
			JobID: item.JobID, Company: item.Company, Title: item.Title, URL: item.URL,
			Status: item.Status, JobFeed: jobFeed, HasDiff: hasDiff,
		}
		switch item.Status {
		case "matched":
			view.Matched = append(view.Matched, pj)
		case "shortlisted":
			view.Shortlisted = append(view.Shortlisted, pj)
		case "tailoring":
			view.Tailoring = append(view.Tailoring, pj)
		case "ready":
			view.Ready = append(view.Ready, pj)
		}
		if hasDiff {
			view.WithDiff = append(view.WithDiff, pj)
		}
	}
	writeJSON(w, view)
}
