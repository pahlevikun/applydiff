package store

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/risk"
)

func (s *Store) UpsertJob(ctx context.Context, j job.Job, dedupeKey string) error {
	if dedupeKey == "" {
		dedupeKey = job.DedupeKey(j)
	}
	roleKey := job.RoleKey(j.Company, j.Title)
	if roleKey != "" {
		existing, err := s.findJobByRoleKey(ctx, roleKey)
		if err == nil && existing != nil {
			if existing.DedupeKey != "" {
				dedupeKey = existing.DedupeKey
			}
			if existing.ID != "" {
				j.ID = existing.ID
			}
		}
	}
	if j.ID == "" {
		j.ID = dedupeKey
	}
	kw, _ := json.Marshal(j.Keywords)
	ben, _ := json.Marshal(j.Benefits)
	origins, _ := json.Marshal(j.Origins)
	mergedOrigins, _ := s.mergeJobOrigins(ctx, dedupeKey, origins)
	firstSeen := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO jobs (id, company, title, url, source, posted_at, first_seen, role_key, remote, region, status, dedupe_key,
			description, salary_post, salary_ats,
			arrangement, days_onsite, weekly_days, async_friendly, timezone_overlap_hours, country,
			employment_type, seniority, salary_min, salary_currency, salary_disclosed, equity,
			company_stage, company_size, visa_sponsorship, relocation, eor_friendly, archetype,
			keywords_json, benefits_json, filter_skip_reason, origins_json,
			risk, flags, applicant_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?)
		ON CONFLICT(dedupe_key) DO UPDATE SET
			company = CASE WHEN excluded.company != '' THEN excluded.company ELSE jobs.company END,
			title = CASE WHEN excluded.title != '' THEN excluded.title ELSE jobs.title END,
			url = CASE WHEN excluded.url != '' THEN excluded.url ELSE jobs.url END,
			source = CASE WHEN excluded.source != '' THEN excluded.source ELSE jobs.source END,
			posted_at = CASE WHEN excluded.posted_at != '' THEN excluded.posted_at ELSE jobs.posted_at END,
			role_key = CASE WHEN excluded.role_key != '' THEN excluded.role_key ELSE jobs.role_key END,
			remote = CASE WHEN excluded.remote != '' THEN excluded.remote ELSE jobs.remote END,
			region = CASE WHEN excluded.region != '' THEN excluded.region ELSE jobs.region END,
			description = CASE WHEN excluded.description != '' THEN excluded.description ELSE jobs.description END,
			salary_post = CASE WHEN excluded.salary_post != '' THEN excluded.salary_post ELSE jobs.salary_post END,
			salary_ats = CASE WHEN excluded.salary_ats != '' THEN excluded.salary_ats ELSE jobs.salary_ats END,
			arrangement = CASE WHEN excluded.arrangement != '' THEN excluded.arrangement ELSE jobs.arrangement END,
			days_onsite = CASE WHEN excluded.days_onsite > 0 THEN excluded.days_onsite ELSE jobs.days_onsite END,
			weekly_days = CASE WHEN excluded.weekly_days > 0 AND excluded.weekly_days != 5 THEN excluded.weekly_days ELSE jobs.weekly_days END,
			async_friendly = excluded.async_friendly,
			timezone_overlap_hours = CASE WHEN excluded.timezone_overlap_hours > 0 THEN excluded.timezone_overlap_hours ELSE jobs.timezone_overlap_hours END,
			country = CASE WHEN excluded.country != '' THEN excluded.country ELSE jobs.country END,
			employment_type = CASE WHEN excluded.employment_type != '' THEN excluded.employment_type ELSE jobs.employment_type END,
			seniority = CASE WHEN excluded.seniority != '' THEN excluded.seniority ELSE jobs.seniority END,
			salary_min = CASE WHEN excluded.salary_min > 0 THEN excluded.salary_min ELSE jobs.salary_min END,
			salary_currency = CASE WHEN excluded.salary_currency != '' THEN excluded.salary_currency ELSE jobs.salary_currency END,
			salary_disclosed = excluded.salary_disclosed,
			equity = excluded.equity,
			company_stage = CASE WHEN excluded.company_stage != '' THEN excluded.company_stage ELSE jobs.company_stage END,
			company_size = CASE WHEN excluded.company_size != '' THEN excluded.company_size ELSE jobs.company_size END,
			visa_sponsorship = excluded.visa_sponsorship,
			relocation = excluded.relocation,
			eor_friendly = excluded.eor_friendly,
			archetype = CASE WHEN excluded.archetype != '' THEN excluded.archetype ELSE jobs.archetype END,
			keywords_json = CASE WHEN excluded.keywords_json != '[]' THEN excluded.keywords_json ELSE jobs.keywords_json END,
			benefits_json = CASE WHEN excluded.benefits_json != '[]' THEN excluded.benefits_json ELSE jobs.benefits_json END,
			origins_json = excluded.origins_json,
			risk = excluded.risk,
			flags = excluded.flags,
			applicant_count = CASE WHEN excluded.applicant_count > 0 THEN excluded.applicant_count ELSE jobs.applicant_count END
	`, j.ID, j.Company, j.Title, j.URL, j.Source, j.PostedAt, firstSeen, roleKey, j.Remote, j.Region, j.Status, dedupeKey,
		j.Description, j.SalaryPost, j.SalaryATS,
		j.Arrangement, j.DaysOnsite, j.WeeklyDays, boolInt(j.AsyncFriendly), j.TimezoneOverlapHours, j.Country,
		j.EmploymentType, j.Seniority, j.SalaryMin, j.SalaryCurrency, boolInt(j.SalaryDisclosed), boolInt(j.Equity),
		j.CompanyStage, j.CompanySize, boolInt(j.VisaSponsorship), boolInt(j.Relocation), boolInt(j.EORFriendly), j.Archetype,
		string(kw), string(ben), j.FilterSkipReason, string(mergedOrigins),
		j.Risk, flagsJSON(j), j.ApplicantCount)
	return err
}

func flagsJSON(j job.Job) string {
	if j.FlagsJSON != "" {
		return j.FlagsJSON
	}
	if len(j.Flags) == 0 {
		return "[]"
	}
	raw, err := json.Marshal(j.Flags)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

type jobRowRef struct {
	ID        string
	DedupeKey string
}

func (s *Store) findJobByRoleKey(ctx context.Context, roleKey string) (*jobRowRef, error) {
	var ref jobRowRef
	err := s.db.QueryRowContext(ctx, `
		SELECT id, dedupe_key FROM jobs
		WHERE role_key = ? AND role_key != ''
		ORDER BY first_seen ASC, id ASC
		LIMIT 1
	`, roleKey).Scan(&ref.ID, &ref.DedupeKey)
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

func (s *Store) mergeJobOrigins(ctx context.Context, dedupeKey string, incoming []byte) ([]byte, error) {
	var existing []string
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT origins_json FROM jobs WHERE dedupe_key = ?`, dedupeKey).Scan(&raw)
	if err == nil && raw != "" {
		_ = json.Unmarshal([]byte(raw), &existing)
	}
	var add []string
	_ = json.Unmarshal(incoming, &add)
	merged := mergeOriginStrings(existing, add)
	out, err := json.Marshal(merged)
	return out, err
}

func mergeOriginStrings(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	var out []string
	for _, list := range [][]string{a, b} {
		for _, v := range list {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

func (s *Store) ListJobs(ctx context.Context) ([]job.Job, error) {
	rows, err := s.db.QueryContext(ctx, jobSelectSQL+`
		WHERE status != 'dead'
		ORDER BY posted_at DESC, title ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []job.Job
	for rows.Next() {
		j, err := scanJobRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, enrichJob(j))
	}
	return out, rows.Err()
}

func enrichJob(j job.Job) job.Job {
	j.Restamped = job.Restamped(j.PostedAt, j.FirstSeen)
	return j
}

func (s *Store) InsertJob(ctx context.Context, j job.Job) error {
	return s.UpsertJob(ctx, j, job.DedupeKey(j))
}

func (s *Store) Close() error {
	return s.db.Close()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

const jobSelectSQL = `
	SELECT id, company, title, url, source, posted_at, first_seen, role_key, remote, region, status,
		description, salary_post, salary_ats,
		arrangement, days_onsite, weekly_days, async_friendly, timezone_overlap_hours, country,
		employment_type, seniority, salary_min, salary_currency, salary_disclosed, equity,
		company_stage, company_size, visa_sponsorship, relocation, eor_friendly, archetype,
		keywords_json, benefits_json, filter_skip_reason, origins_json,
		risk, flags, applicant_count, dedupe_key
	FROM jobs`

func scanJobRow(scanner interface {
	Scan(dest ...any) error
}) (job.Job, error) {
	var j job.Job
	var desc, salaryPost, salaryATS string
	var async, salDisc, equity, visa, reloc, eor int
	var kwJSON, benJSON, originsJSON, flagsRaw string
	var dedupeKey string
	err := scanner.Scan(
		&j.ID, &j.Company, &j.Title, &j.URL, &j.Source, &j.PostedAt, &j.FirstSeen, &j.RoleKey, &j.Remote, &j.Region, &j.Status,
		&desc, &salaryPost, &salaryATS,
		&j.Arrangement, &j.DaysOnsite, &j.WeeklyDays, &async, &j.TimezoneOverlapHours, &j.Country,
		&j.EmploymentType, &j.Seniority, &j.SalaryMin, &j.SalaryCurrency, &salDisc, &equity,
		&j.CompanyStage, &j.CompanySize, &visa, &reloc, &eor, &j.Archetype,
		&kwJSON, &benJSON, &j.FilterSkipReason, &originsJSON,
		&j.Risk, &flagsRaw, &j.ApplicantCount, &dedupeKey,
	)
	if err != nil {
		return j, err
	}
	j.Description = desc
	j.SalaryPost = salaryPost
	j.SalaryATS = salaryATS
	j.AsyncFriendly = async != 0
	j.SalaryDisclosed = salDisc != 0
	j.Equity = equity != 0
	j.VisaSponsorship = visa != 0
	j.Relocation = reloc != 0
	j.EORFriendly = eor != 0
	_ = json.Unmarshal([]byte(kwJSON), &j.Keywords)
	_ = json.Unmarshal([]byte(benJSON), &j.Benefits)
	_ = json.Unmarshal([]byte(originsJSON), &j.Origins)
	j.FlagsJSON = flagsRaw
	j.Flags = decodeJobFlags(flagsRaw)
	_ = dedupeKey
	return j, nil
}

func decodeJobFlags(raw string) []job.JobRiskFlag {
	parsed := risk.ParseFlagsJSON(raw)
	out := make([]job.JobRiskFlag, 0, len(parsed))
	for _, f := range parsed {
		out = append(out, job.JobRiskFlag{Code: f.Code, Severity: f.Severity, Message: f.Message})
	}
	return out
}

// UpdateJobRisk persists scam-guard fields for an existing row.
func (s *Store) UpdateJobRisk(ctx context.Context, id string, score int, flagsJSON string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE jobs SET risk = ?, flags = ? WHERE id = ?`, score, flagsJSON, id)
	return err
}

// ListJobDescriptionHashes returns normalized description hashes for duplicate checks.
func (s *Store) ListJobDescriptionHashes(ctx context.Context) (map[string][]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, description FROM jobs WHERE status != 'dead' AND description != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var id, desc string
		if err := rows.Scan(&id, &desc); err != nil {
			return nil, err
		}
		h := risk.DescriptionHash(desc)
		out[h] = append(out[h], id)
	}
	return out, rows.Err()
}

// CountJobs returns the number of job rows (including dead).
func (s *Store) CountJobs(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM jobs`).Scan(&n)
	return n, err
}
