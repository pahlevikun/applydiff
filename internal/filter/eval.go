package filter

import (
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/job"
)

func softBoost(ruleBoost int) int {
	if ruleBoost > 0 {
		return ruleBoost
	}
	return 5
}

// Evaluate applies spec to j. AND across categories; OR within category values.
func Evaluate(j job.Job, spec Spec) Result {
	res := Result{Keep: true, Score: 0, Reasons: []string{}}
	text := strings.ToLower(j.Description + " " + j.Title)

	evalCategory := func(name string, rule *CategoryRule, match func(v string) bool, jobVal string, optional bool) {
		if rule == nil || len(rule.Values) == 0 {
			return
		}
		matched := false
		var hit string
		for _, v := range rule.Values {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			if match(v) {
				matched = true
				hit = v
				break
			}
		}
		if rule.Hard {
			if !matched {
				if optional && strings.TrimSpace(jobVal) == "" {
					return
				}
				res.Keep = false
				res.Reasons = append(res.Reasons, skipReason(name, jobVal, rule.Values))
			} else {
				res.Reasons = append(res.Reasons, "matched: "+name+" "+hit)
			}
			return
		}
		if matched {
			res.Score += softBoost(rule.Boost)
			res.Reasons = append(res.Reasons, "boost: "+name+" "+hit)
		}
	}

	// Arrangement
	evalCategory("arrangement", spec.Arrangement, func(v string) bool {
		return strings.EqualFold(j.Arrangement, v) || legacyRemoteMatch(j, v)
	}, j.Arrangement, false)

	// Regions (also match legacy Region field)
	evalCategory("region", spec.Regions, func(v string) bool {
		return regionMatch(j.Region, v) || regionMatch(j.Country, v)
	}, j.Region, false)

	evalCategory("country", spec.Countries, func(v string) bool {
		return regionMatch(j.Country, v)
	}, j.Country, true)

	if spec.ExcludeRegions != nil && len(spec.ExcludeRegions.Values) > 0 {
		for _, ex := range spec.ExcludeRegions.Values {
			if regionMatch(j.Region, ex) || regionMatch(j.Country, ex) {
				if spec.ExcludeRegions.Hard {
					res.Keep = false
					res.Reasons = append(res.Reasons, "skipped: excluded region "+ex)
				}
				break
			}
		}
	}

	// Days on-site (hybrid): job days must be <= max in rule values
	if spec.DaysOnsite != nil && (spec.DaysOnsite.Max > 0 || spec.DaysOnsite.Min > 0) {
		ok := true
		if spec.DaysOnsite.Max > 0 && j.DaysOnsite > spec.DaysOnsite.Max {
			ok = false
		}
		if spec.DaysOnsite.Min > 0 && j.DaysOnsite < spec.DaysOnsite.Min {
			ok = false
		}
		if spec.DaysOnsite.Hard && !ok {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: days on-site "+itoa(j.DaysOnsite))
		} else if !spec.DaysOnsite.Hard && ok && j.DaysOnsite > 0 {
			res.Score += softBoost(spec.DaysOnsite.Boost)
			res.Reasons = append(res.Reasons, "boost: days on-site")
		}
	}

	evalCategory("weekly_days", spec.WeeklyDays, func(v string) bool {
		return weeklyDaysMatch(j.WeeklyDays, v)
	}, formatWeeklyDays(j.WeeklyDays), true)

	if spec.AsyncFriendly != nil && spec.AsyncFriendly.Require {
		if spec.AsyncFriendly.Hard && !j.AsyncFriendly {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: not async-friendly")
		} else if !spec.AsyncFriendly.Hard && j.AsyncFriendly {
			res.Score += softBoost(spec.AsyncFriendly.Boost)
			res.Reasons = append(res.Reasons, "boost: async-friendly")
		}
	}

	if spec.TimezoneOverlapHours != nil && spec.TimezoneOverlapHours.Min > 0 {
		ok := j.TimezoneOverlapHours >= spec.TimezoneOverlapHours.Min
		if spec.TimezoneOverlapHours.Hard && !ok {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: timezone overlap < "+itoa(spec.TimezoneOverlapHours.Min)+"h")
		} else if !spec.TimezoneOverlapHours.Hard && ok {
			res.Score += softBoost(spec.TimezoneOverlapHours.Boost)
			res.Reasons = append(res.Reasons, "boost: timezone overlap")
		}
	}

	evalCategory("employment_type", spec.EmploymentType, func(v string) bool {
		return strings.EqualFold(j.EmploymentType, v)
	}, j.EmploymentType, true)

	if spec.ExcludeEmployment != nil {
		for _, ex := range spec.ExcludeEmployment.Values {
			if strings.EqualFold(j.EmploymentType, ex) {
				if spec.ExcludeEmployment.Hard {
					res.Keep = false
					res.Reasons = append(res.Reasons, "skipped: excluded employment "+ex)
				}
				break
			}
		}
	}

	evalCategory("seniority", spec.Seniority, func(v string) bool {
		return seniorityMatch(j.Seniority, v) || seniorityMatch(j.Title, v)
	}, j.Seniority, true)

	if spec.MinSeniority != nil && len(spec.MinSeniority.Values) > 0 {
		min := spec.MinSeniority.Values[0]
		ok := seniorityAtLeast(j.Seniority, j.Title, min)
		if spec.MinSeniority.Hard && !ok {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: below min seniority "+min)
		}
	}

	if spec.SalaryMin != nil && spec.SalaryMin.Min > 0 {
		ok := j.SalaryDisclosed && j.SalaryMin >= spec.SalaryMin.Min
		if spec.SalaryMin.Currency != "" {
			ok = ok && strings.EqualFold(j.SalaryCurrency, spec.SalaryMin.Currency)
		}
		if spec.SalaryMin.Hard && !ok {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: salary below minimum")
		} else if !spec.SalaryMin.Hard && ok {
			res.Score += softBoost(spec.SalaryMin.Boost)
			res.Reasons = append(res.Reasons, "boost: meets salary minimum")
		}
	}

	if spec.SalaryDisclosed != nil && spec.SalaryDisclosed.Require {
		if spec.SalaryDisclosed.Hard && !j.SalaryDisclosed {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: salary not disclosed")
		} else if !spec.SalaryDisclosed.Hard && j.SalaryDisclosed {
			res.Score += softBoost(spec.SalaryDisclosed.Boost)
			res.Reasons = append(res.Reasons, "boost: salary disclosed")
		}
	}

	if spec.Equity != nil && spec.Equity.Require {
		if spec.Equity.Hard && !j.Equity {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: no equity offered")
		} else if !spec.Equity.Hard && j.Equity {
			res.Score += softBoost(spec.Equity.Boost)
			res.Reasons = append(res.Reasons, "boost: equity offered")
		}
	}

	evalCategory("company_stage", spec.CompanyStage, func(v string) bool {
		return strings.EqualFold(j.CompanyStage, v)
	}, j.CompanyStage, true)

	evalCategory("company_size", spec.CompanySize, func(v string) bool {
		return strings.EqualFold(j.CompanySize, v)
	}, j.CompanySize, true)

	if spec.VisaSponsorship != nil && spec.VisaSponsorship.Require {
		if spec.VisaSponsorship.Hard && !j.VisaSponsorship {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: no visa sponsorship")
		} else if !spec.VisaSponsorship.Hard && j.VisaSponsorship {
			res.Score += softBoost(spec.VisaSponsorship.Boost)
			res.Reasons = append(res.Reasons, "boost: visa sponsorship")
		}
	}

	if spec.Relocation != nil && spec.Relocation.Require {
		if spec.Relocation.Hard && !j.Relocation {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: no relocation")
		} else if !spec.Relocation.Hard && j.Relocation {
			res.Score += softBoost(spec.Relocation.Boost)
			res.Reasons = append(res.Reasons, "boost: relocation offered")
		}
	}

	if spec.EORFriendly != nil && spec.EORFriendly.Require {
		if spec.EORFriendly.Hard && !j.EORFriendly {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: not EOR-friendly")
		} else if !spec.EORFriendly.Hard && j.EORFriendly {
			res.Score += softBoost(spec.EORFriendly.Boost)
			res.Reasons = append(res.Reasons, "boost: EOR-friendly")
		}
	}

	evalCategory("archetype", spec.Archetype, func(v string) bool {
		return strings.EqualFold(j.Archetype, v)
	}, j.Archetype, true)

	evalKeywords(&res, spec.MustHaveKeywords, j.Keywords, text, true)
	evalKeywords(&res, spec.NiceToHaveKeywords, j.Keywords, text, false)
	evalKeywordsExclude(&res, spec.ExcludeKeywords, j.Keywords, text)

	evalCategory("benefits", spec.Benefits, func(v string) bool {
		return sliceContainsFold(j.Benefits, v)
	}, "", true)

	if spec.MaxPostedDays != nil && spec.MaxPostedDays.Max > 0 {
		cutoff := time.Now().AddDate(0, 0, -spec.MaxPostedDays.Max)
		if posted, ok := parsePosted(j.PostedAt); ok && posted.Before(cutoff) {
			if spec.MaxPostedDays.Hard {
				res.Keep = false
				res.Reasons = append(res.Reasons, "skipped: posted too long ago")
			}
		}
	}

	if spec.PostedWithin != nil && spec.PostedWithin.Max > 0 {
		if posted, ok := parsePosted(j.PostedAt); ok {
			age := time.Since(posted)
			limit := time.Duration(spec.PostedWithin.Max) * time.Hour
			if age > limit {
				if spec.PostedWithin.Hard {
					res.Keep = false
					res.Reasons = append(res.Reasons, "skipped: posted outside "+itoa(spec.PostedWithin.Max)+"h window")
				}
			} else if !spec.PostedWithin.Hard {
				res.Score += softBoost(spec.PostedWithin.Boost)
				res.Reasons = append(res.Reasons, "boost: fresh posting")
			}
		} else if spec.PostedWithin.Hard {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: posted time unknown")
		}
	}

	if spec.MaxApplicants != nil && spec.MaxApplicants.Max > 0 && j.ApplicantCount > 0 {
		ok := j.ApplicantCount <= spec.MaxApplicants.Max
		if spec.MaxApplicants.Hard && !ok {
			res.Keep = false
			res.Reasons = append(res.Reasons, "skipped: applicant count "+itoa(j.ApplicantCount))
		} else if !spec.MaxApplicants.Hard && ok {
			res.Score += softBoost(spec.MaxApplicants.Boost)
			res.Reasons = append(res.Reasons, "boost: low competition")
		}
	}

	if res.Keep && len(res.Reasons) == 0 {
		res.Reasons = append(res.Reasons, "matched: no hard filters failed")
	}
	return res
}

func evalKeywords(res *Result, rule *KeywordRule, keywords []string, text string, must bool) {
	if rule == nil || len(rule.Values) == 0 {
		return
	}
	if rule.All || must {
		for _, kw := range rule.Values {
			if !keywordHit(kw, keywords, text) {
				if rule.Hard {
					res.Keep = false
					res.Reasons = append(res.Reasons, "skipped: missing keyword "+kw)
				}
				return
			}
		}
		if rule.Hard {
			res.Reasons = append(res.Reasons, "matched: must-have keywords")
		} else {
			res.Score += softBoost(rule.Boost)
			res.Reasons = append(res.Reasons, "boost: must-have keywords")
		}
		return
	}
	matched := false
	for _, kw := range rule.Values {
		if keywordHit(kw, keywords, text) {
			matched = true
			break
		}
	}
	if rule.Hard && !matched {
		res.Keep = false
		res.Reasons = append(res.Reasons, "skipped: no nice-to-have keyword match")
	} else if !rule.Hard && matched {
		res.Score += softBoost(rule.Boost)
		res.Reasons = append(res.Reasons, "boost: keyword match")
	}
}

func evalKeywordsExclude(res *Result, rule *KeywordRule, keywords []string, text string) {
	if rule == nil || len(rule.Values) == 0 {
		return
	}
	for _, kw := range rule.Values {
		if keywordHit(kw, keywords, text) {
			if rule.Hard {
				res.Keep = false
				res.Reasons = append(res.Reasons, "skipped: excluded keyword "+kw)
			}
			return
		}
	}
}

func keywordHit(kw string, keywords []string, text string) bool {
	kw = strings.TrimSpace(strings.ToLower(kw))
	if kw == "" {
		return false
	}
	if strings.Contains(text, kw) {
		return true
	}
	return sliceContainsFold(keywords, kw)
}

func skipReason(name, jobVal string, allowed []string) string {
	if jobVal != "" {
		return "skipped: " + name + " " + jobVal + " not in " + strings.Join(allowed, "/")
	}
	return "skipped: " + name + " not in " + strings.Join(allowed, "/")
}

func legacyRemoteMatch(j job.Job, v string) bool {
	if strings.EqualFold(v, "remote") {
		return j.Remote == "yes" || strings.EqualFold(j.Arrangement, "remote")
	}
	return false
}

func regionMatch(jobRegion, want string) bool {
	jobRegion = strings.ToLower(strings.TrimSpace(jobRegion))
	want = strings.ToLower(strings.TrimSpace(want))
	if jobRegion == "" {
		return want == "global" || want == "worldwide"
	}
	return jobRegion == want || strings.Contains(jobRegion, want)
}

func weeklyDaysMatch(jobDays float64, want string) bool {
	switch strings.TrimSpace(want) {
	case "4", "4-day", "4_day":
		return jobDays >= 3.75 && jobDays <= 4.25
	case "4.5", "4.5-day":
		return jobDays >= 4.25 && jobDays <= 4.75
	case "5", "5-day":
		return jobDays >= 4.75
	default:
		return formatWeeklyDays(jobDays) == want
	}
}

func formatWeeklyDays(d float64) string {
	if d <= 0 {
		return ""
	}
	if d >= 3.75 && d <= 4.25 {
		return "4"
	}
	if d >= 4.25 && d <= 4.75 {
		return "4.5"
	}
	return "5"
}

var seniorityRank = map[string]int{
	"intern": 0, "junior": 1, "mid": 2, "middle": 2,
	"senior": 3, "staff": 4, "principal": 5, "lead": 4,
	"manager": 5, "director": 6,
}

func seniorityMatch(field, level string) bool {
	return strings.Contains(strings.ToLower(field), strings.ToLower(level))
}

func seniorityAtLeast(seniority, title, min string) bool {
	min = strings.ToLower(strings.TrimSpace(min))
	minRank, ok := seniorityRank[min]
	if !ok {
		return seniorityMatch(title, min) || seniorityMatch(seniority, min)
	}
	for name, rank := range seniorityRank {
		if rank >= minRank && (seniorityMatch(seniority, name) || seniorityMatch(title, name)) {
			return true
		}
	}
	return false
}

func sliceContainsFold(slice []string, v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	for _, s := range slice {
		if strings.ToLower(strings.TrimSpace(s)) == v {
			return true
		}
	}
	return false
}

func parsePosted(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05-07:00"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
