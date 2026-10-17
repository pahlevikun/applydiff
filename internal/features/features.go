package features

// InterviewPrepEnabled gates interview prep across CLI, API, TUI, MCP, and web UI.
// Keep web/utils/features.ts in sync when changing this.
const InterviewPrepEnabled = false

// View is exposed on GET /api/config for the web shell.
type View struct {
	InterviewPrep bool `json:"interview_prep"`
}

func APIView() View {
	return View{InterviewPrep: InterviewPrepEnabled}
}
