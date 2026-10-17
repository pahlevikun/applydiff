package job

import (
	"net/url"
	"strings"
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

// DedupeKey collapses listings that share the same normalized URL and company+role.
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
