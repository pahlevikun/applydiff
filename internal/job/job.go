package job

type Job struct {
	ID          string `json:"id"`
	Company     string `json:"company"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Source      string `json:"source"`
	PostedAt    string `json:"posted_at"`
	Remote      string `json:"remote"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	SalaryPost  string `json:"salary_post,omitempty"`
	SalaryATS   string `json:"salary_ats,omitempty"`
}
