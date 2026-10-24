package job

type Job struct {
	ID          string `json:"id"`
	DedupeKey   string `json:"dedupe_key,omitempty"`
	Company     string `json:"company"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Source      string   `json:"source"`
	Origins     []string `json:"origins,omitempty"`
	PostedAt    string `json:"posted_at"`
	Remote      string `json:"remote"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	SalaryPost  string `json:"salary_post,omitempty"`
	SalaryATS   string `json:"salary_ats,omitempty"`

	Arrangement          string   `json:"arrangement,omitempty"`
	DaysOnsite           int      `json:"days_onsite,omitempty"`
	WeeklyDays           float64  `json:"weekly_days,omitempty"`
	AsyncFriendly        bool     `json:"async_friendly,omitempty"`
	TimezoneOverlapHours int      `json:"timezone_overlap_hours,omitempty"`
	Country              string   `json:"country,omitempty"`
	EmploymentType       string   `json:"employment_type,omitempty"`
	Seniority            string   `json:"seniority,omitempty"`
	SalaryMin            int      `json:"salary_min,omitempty"`
	SalaryCurrency       string   `json:"salary_currency,omitempty"`
	SalaryDisclosed      bool     `json:"salary_disclosed,omitempty"`
	Equity               bool     `json:"equity,omitempty"`
	CompanyStage         string   `json:"company_stage,omitempty"`
	CompanySize          string   `json:"company_size,omitempty"`
	VisaSponsorship      bool     `json:"visa_sponsorship,omitempty"`
	Relocation           bool     `json:"relocation,omitempty"`
	EORFriendly          bool     `json:"eor_friendly,omitempty"`
	Archetype            string   `json:"archetype,omitempty"`
	Keywords             []string `json:"keywords,omitempty"`
	Benefits             []string `json:"benefits,omitempty"`

	FilterSkipReason string `json:"filter_skip_reason,omitempty"`

	Risk           int `json:"risk,omitempty"`
	ApplicantCount int `json:"applicant_count,omitempty"`
	FlagsJSON      string `json:"-"`
	Flags          []JobRiskFlag `json:"flags,omitempty"`
}

// JobRiskFlag mirrors stored scam-guard signals for API consumers.
type JobRiskFlag struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// JobView is returned from the jobs API with live filter evaluation.
type JobView struct {
	Job    Job `json:"job"`
	Filter struct {
		Keep    bool     `json:"keep"`
		Score   int      `json:"score"`
		Reasons []string `json:"reasons"`
	} `json:"filter"`
	Priority float64 `json:"priority,omitempty"`
}
