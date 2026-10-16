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
	seen := make(map[string]struct{}, len(jobs))
	out := make([]Job, 0, len(jobs))
	for _, j := range jobs {
		k := DedupeKey(j)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, j)
	}
	return out
}
