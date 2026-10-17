package filter

import (
	"encoding/json"
)

// Spec is the full filter configuration (config yaml or saved preset).
// AND across non-empty categories; OR within each category's values.
type Spec struct {
	Arrangement           *CategoryRule `json:"arrangement,omitempty" yaml:"arrangement,omitempty"`
	DaysOnsite            *IntRule      `json:"days_onsite,omitempty" yaml:"days_onsite,omitempty"`
	WeeklyDays            *CategoryRule `json:"weekly_days,omitempty" yaml:"weekly_days,omitempty"`
	AsyncFriendly         *BoolRule     `json:"async_friendly,omitempty" yaml:"async_friendly,omitempty"`
	TimezoneOverlapHours  *IntRule      `json:"timezone_overlap_hours,omitempty" yaml:"timezone_overlap_hours,omitempty"`
	Regions               *CategoryRule `json:"regions,omitempty" yaml:"regions,omitempty"`
	Countries             *CategoryRule `json:"countries,omitempty" yaml:"countries,omitempty"`
	ExcludeRegions        *CategoryRule `json:"exclude_regions,omitempty" yaml:"exclude_regions,omitempty"`
	EmploymentType        *CategoryRule `json:"employment_type,omitempty" yaml:"employment_type,omitempty"`
	ExcludeEmployment     *CategoryRule `json:"exclude_employment,omitempty" yaml:"exclude_employment,omitempty"`
	Seniority             *CategoryRule `json:"seniority,omitempty" yaml:"seniority,omitempty"`
	MinSeniority          *CategoryRule `json:"min_seniority,omitempty" yaml:"min_seniority,omitempty"`
	SalaryMin             *SalaryRule   `json:"salary_min,omitempty" yaml:"salary_min,omitempty"`
	SalaryDisclosed       *BoolRule     `json:"salary_disclosed,omitempty" yaml:"salary_disclosed,omitempty"`
	Equity                *BoolRule     `json:"equity,omitempty" yaml:"equity,omitempty"`
	CompanyStage          *CategoryRule `json:"company_stage,omitempty" yaml:"company_stage,omitempty"`
	CompanySize           *CategoryRule `json:"company_size,omitempty" yaml:"company_size,omitempty"`
	VisaSponsorship       *BoolRule     `json:"visa_sponsorship,omitempty" yaml:"visa_sponsorship,omitempty"`
	Relocation            *BoolRule     `json:"relocation,omitempty" yaml:"relocation,omitempty"`
	EORFriendly           *BoolRule     `json:"eor_friendly,omitempty" yaml:"eor_friendly,omitempty"`
	Archetype             *CategoryRule `json:"archetype,omitempty" yaml:"archetype,omitempty"`
	MustHaveKeywords      *KeywordRule  `json:"must_have_keywords,omitempty" yaml:"must_have_keywords,omitempty"`
	NiceToHaveKeywords    *KeywordRule  `json:"nice_to_have_keywords,omitempty" yaml:"nice_to_have_keywords,omitempty"`
	ExcludeKeywords       *KeywordRule  `json:"exclude_keywords,omitempty" yaml:"exclude_keywords,omitempty"`
	Benefits              *CategoryRule `json:"benefits,omitempty" yaml:"benefits,omitempty"`
	MaxPostedDays         *IntRule      `json:"max_posted_days,omitempty" yaml:"max_posted_days,omitempty"`
	RemoteOnly            bool          `json:"remote_only,omitempty" yaml:"remote_only,omitempty"`
}

type CategoryRule struct {
	Hard   bool     `json:"hard" yaml:"hard"`
	Values []string `json:"values" yaml:"values"`
	Boost  int      `json:"boost,omitempty" yaml:"boost,omitempty"`
}

type IntRule struct {
	Hard  bool `json:"hard" yaml:"hard"`
	Min   int  `json:"min,omitempty" yaml:"min,omitempty"`
	Max   int  `json:"max,omitempty" yaml:"max,omitempty"`
	Boost int  `json:"boost,omitempty" yaml:"boost,omitempty"`
}

type BoolRule struct {
	Hard    bool `json:"hard" yaml:"hard"`
	Require bool `json:"require" yaml:"require"`
	Boost   int  `json:"boost,omitempty" yaml:"boost,omitempty"`
}

type SalaryRule struct {
	Hard     bool   `json:"hard" yaml:"hard"`
	Min      int    `json:"min" yaml:"min"`
	Currency string `json:"currency,omitempty" yaml:"currency,omitempty"`
	Boost    int    `json:"boost,omitempty" yaml:"boost,omitempty"`
}

type KeywordRule struct {
	Hard   bool     `json:"hard" yaml:"hard"`
	Values []string `json:"values" yaml:"values"`
	All    bool     `json:"all,omitempty" yaml:"all,omitempty"` // AND if true, else OR within
	Boost  int      `json:"boost,omitempty" yaml:"boost,omitempty"`
}

// Result of evaluating one job against a spec.
type Result struct {
	Keep    bool     `json:"keep"`
	Score   int      `json:"score"`
	Reasons []string `json:"reasons"`
}

func (s Spec) Clone() Spec {
	raw, _ := json.Marshal(s)
	var out Spec
	_ = json.Unmarshal(raw, &out)
	return out
}

// NormalizeLegacy maps flat jobconfig-style fields into category rules when unset.
func NormalizeLegacy(s *Spec, regions []string, remoteOnly bool, minSeniority string, maxPostedDays int) {
	if remoteOnly && s.Arrangement == nil {
		s.Arrangement = &CategoryRule{Hard: true, Values: []string{"remote"}}
	}
	if len(regions) > 0 && s.Regions == nil {
		s.Regions = &CategoryRule{Hard: true, Values: regions}
	}
	if minSeniority != "" && s.MinSeniority == nil {
		s.MinSeniority = &CategoryRule{Hard: true, Values: []string{minSeniority}}
	}
	if maxPostedDays > 0 && s.MaxPostedDays == nil {
		s.MaxPostedDays = &IntRule{Hard: true, Max: maxPostedDays}
	}
}
