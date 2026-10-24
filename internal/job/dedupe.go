package job

import (
	"net/url"
	"strings"
	"time"
)

func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return strings.ToLower(raw)
	}
	u.Fragment = ""
	u.RawQuery = ""
	host := strings.ToLower(u.Host)
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}
	u.Host = host
	u.Scheme = strings.ToLower(u.Scheme)
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}
	u.Path = path
	return u.String()
}

func normalizePart(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}

// DedupeKey is the per-URL upsert identity (normalized URL + company + title).
// Cross-board collapse uses RoleKey via MergeByRoleKey instead.
func DedupeKey(j Job) string {
	return NormalizeURL(j.URL) + "\x00" + normalizePart(j.Company) + "\x00" + normalizePart(j.Title)
}

// Dedupe keeps the first occurrence of each dedupe key.
func Dedupe(jobs []Job) []Job {
	return MergeByDedupeKey(jobs)
}

// MergeByDedupeKey collapses duplicate listings and records every contributing source in Origins.
func MergeByDedupeKey(jobs []Job) []Job {
	byKey := make(map[string]Job, len(jobs))
	order := make([]string, 0, len(jobs))
	for _, j := range jobs {
		k := DedupeKey(j)
		if existing, ok := byKey[k]; ok {
			existing.Origins = mergeOrigins(existing.Origins, j.Source)
			byKey[k] = preferRicher(existing, j)
			continue
		}
		if j.Source != "" {
			j.Origins = []string{j.Source}
		}
		byKey[k] = j
		order = append(order, k)
	}
	out := make([]Job, 0, len(order))
	for _, k := range order {
		out = append(out, byKey[k])
	}
	return out
}

// MergeByRoleKey collapses listings with the same RoleKey (empty keys are left separate).
func MergeByRoleKey(jobs []Job) []Job {
	byKey := make(map[string]Job, len(jobs))
	order := make([]string, 0, len(jobs))
	for _, j := range jobs {
		k := RoleKey(j.Company, j.Title)
		if k == "" {
			order = append(order, "solo:"+DedupeKey(j))
			byKey["solo:"+DedupeKey(j)] = j
			continue
		}
		if existing, ok := byKey[k]; ok {
			merged := mergeRoleDuplicates(existing, j)
			byKey[k] = merged
			continue
		}
		if j.Source != "" && len(j.Origins) == 0 {
			j.Origins = []string{j.Source}
		}
		byKey[k] = j
		order = append(order, k)
	}
	out := make([]Job, 0, len(order))
	seen := make(map[string]struct{}, len(order))
	for _, k := range order {
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, byKey[k])
	}
	return out
}

func mergeRoleDuplicates(a, b Job) Job {
	canon := preferCanonicalJob(a, b)
	var other Job
	if sameJobIdentity(canon, a) {
		other = b
	} else {
		other = a
	}
	canon.Origins = mergeOrigins(canon.Origins, other.Source)
	for _, o := range other.Origins {
		canon.Origins = mergeOrigins(canon.Origins, o)
	}
	return preferRicher(canon, other)
}

func sameJobIdentity(a, b Job) bool {
	if a.ID != "" && b.ID != "" {
		return a.ID == b.ID
	}
	return DedupeKey(a) == DedupeKey(b)
}

func preferCanonicalJob(a, b Job) Job {
	ta, oka := firstSeenOrPosted(a)
	tb, okb := firstSeenOrPosted(b)
	if oka && okb {
		if ta.Before(tb) {
			return a
		}
		if tb.Before(ta) {
			return b
		}
	} else if oka && !okb {
		return a
	} else if okb && !oka {
		return b
	}
	if a.ID != "" && b.ID != "" {
		if a.ID < b.ID {
			return a
		}
		if b.ID < a.ID {
			return b
		}
	}
	if DedupeKey(a) < DedupeKey(b) {
		return a
	}
	return b
}

func firstSeenOrPosted(j Job) (time.Time, bool) {
	if j.FirstSeen != "" {
		if t, err := parseJobTime(j.FirstSeen); err == nil {
			return t, true
		}
	}
	if j.PostedAt != "" {
		if t, err := parseJobTime(j.PostedAt); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func mergeOrigins(existing []string, source string) []string {
	if source == "" {
		return existing
	}
	for _, o := range existing {
		if o == source {
			return existing
		}
	}
	return append(existing, source)
}

func preferRicher(a, b Job) Job {
	if len(b.Description) > len(a.Description) {
		a.Description = b.Description
	}
	if a.SalaryPost == "" && b.SalaryPost != "" {
		a.SalaryPost = b.SalaryPost
	}
	if len(b.Keywords) > len(a.Keywords) {
		a.Keywords = b.Keywords
	}
	a.Origins = mergeOrigins(a.Origins, b.Source)
	return a
}
