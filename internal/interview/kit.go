package interview

import (
	"encoding/json"
	"strings"

	"github.com/pahlevikun/applydiff/internal/job"
)

// BuildKitJSON derives loop, topics, prompts, and questions from a job (no LLM).
func BuildKitJSON(j *job.Job) (loop, topics, prompts, questions string) {
	title := strings.ToLower(j.Title)
	desc := strings.ToLower(j.Description)
	loopStages := []string{"recruiter screen", "technical screen"}
	if strings.Contains(title, "senior") || strings.Contains(title, "staff") {
		loopStages = append(loopStages, "system design", "behavioral")
	} else {
		loopStages = append(loopStages, "coding", "behavioral")
	}
	if strings.Contains(desc, "product") || strings.Contains(title, "product") {
		loopStages = append(loopStages, "product case")
	}
	if strings.Contains(desc, "business") || strings.Contains(desc, "strategy") {
		loopStages = append(loopStages, "business case")
	}

	topicSet := map[string]struct{}{
		"scalability": {}, "caching": {}, "databases": {}, "rate-limiting": {},
	}
	for _, kw := range []struct{ needle, topic string }{
		{"api", "rate-limiting"},
		{"payment", "message-queues"},
		{"webhook", "event-driven"},
		{"distributed", "cap-theorem"},
		{"latency", "latency-vs-throughput"},
		{"cdn", "dns-cdn"},
		{"shard", "sharding"},
		{"cache", "caching"},
		{"queue", "message-queues"},
		{"load balanc", "load-balancers"},
	} {
		if strings.Contains(desc, kw.needle) || strings.Contains(title, kw.needle) {
			topicSet[kw.topic] = struct{}{}
		}
	}
	topicsList := make([]string, 0, len(topicSet))
	for t := range topicSet {
		topicsList = append(topicsList, t)
	}
	if len(topicsList) == 0 {
		topicsList = []string{"scalability", "caching", "databases"}
	}

	promptList := []map[string]string{
		{
			"type":   "system-design",
			"prompt": "Design a backend for " + j.Company + "'s core product surface described in the JD.",
		},
	}
	if strings.Contains(desc, "billing") || strings.Contains(desc, "subscription") {
		promptList[0]["prompt"] = "Design a subscription billing system with webhooks and idempotent charges for " + j.Company + "."
	}

	questionsList := []string{
		"How does this team measure success in the first 90 days?",
		"What does on-call look like for this role?",
		"Where is the biggest technical risk on the roadmap?",
	}

	loopB, _ := json.Marshal(loopStages)
	topicsB, _ := json.Marshal(topicsList)
	promptsB, _ := json.Marshal(promptList)
	questionsB, _ := json.Marshal(questionsList)
	return string(loopB), string(topicsB), string(promptsB), string(questionsB)
}
