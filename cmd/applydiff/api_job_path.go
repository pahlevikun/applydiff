package main

import (
	"fmt"
	"net/url"
	"strings"
)

// jobIDFromAPIPath returns the job id after a fixed API prefix (e.g. "/api/match/").
// Ids are often URL-based dedupe keys and may contain slashes; clients must path-escape them.
func jobIDFromAPIPath(prefix, requestPath string) string {
	rest := strings.TrimPrefix(requestPath, prefix)
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return ""
	}
	decoded, err := url.PathUnescape(rest)
	if err != nil {
		return rest
	}
	return decoded
}

type jobAPISubroute int

const (
	jobAPISubNone   jobAPISubroute = 0
	jobAPISubHunks  jobAPISubroute = 1
	jobAPISubBuild  jobAPISubroute = 2
)

// parseJobAPIPath splits /api/jobs/{key} and optional /hunks, /hunks/{n}, or /build.
func parseJobAPIPath(requestPath string) (jobKey string, sub jobAPISubroute, hunkIndex int, ok bool) {
	rest := strings.TrimPrefix(requestPath, "/api/jobs/")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return "", jobAPISubNone, 0, false
	}
	if strings.HasSuffix(rest, "/build") {
		key := strings.TrimSuffix(rest, "/build")
		key = strings.Trim(key, "/")
		if key == "" {
			return "", jobAPISubNone, 0, false
		}
		return pathUnescapeJobKey(key), jobAPISubBuild, 0, true
	}
	const marker = "/hunks"
	if idx := strings.Index(rest, marker); idx >= 0 {
		jobKey = strings.Trim(rest[:idx], "/")
		tail := strings.TrimPrefix(rest[idx:], marker)
		tail = strings.Trim(tail, "/")
		if tail == "" {
			return pathUnescapeJobKey(jobKey), jobAPISubHunks, 0, true
		}
		n, err := parseHunkIndexSegment(tail)
		if err != nil {
			return "", jobAPISubNone, 0, false
		}
		return pathUnescapeJobKey(jobKey), jobAPISubHunks, n, true
	}
	return pathUnescapeJobKey(rest), jobAPISubNone, 0, true
}

func pathUnescapeJobKey(raw string) string {
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return raw
	}
	return decoded
}

func parseHunkIndexSegment(seg string) (int, error) {
	if seg == "" {
		return 0, fmt.Errorf("empty hunk index")
	}
	var n int
	for _, c := range seg {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid hunk index")
		}
		n = n*10 + int(c-'0')
	}
	if n < 1 {
		return 0, fmt.Errorf("invalid hunk index")
	}
	return n, nil
}

type interviewSubroute int

const (
	interviewRouteWorkspace interviewSubroute = 0
	interviewRouteKit       interviewSubroute = 1
	interviewRouteMockSave  interviewSubroute = 2
	interviewRouteStories   interviewSubroute = 3
)

func parseInterviewPath(requestPath string) (jobKey string, route interviewSubroute, ok bool) {
	rest := strings.TrimPrefix(requestPath, "/api/interview/")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return "", 0, false
	}
	switch {
	case strings.HasSuffix(rest, "/mock/save"):
		jobKey = strings.TrimSuffix(rest, "/mock/save")
		route = interviewRouteMockSave
	case strings.HasSuffix(rest, "/stories"):
		jobKey = strings.TrimSuffix(rest, "/stories")
		route = interviewRouteStories
	case strings.HasSuffix(rest, "/kit"):
		jobKey = strings.TrimSuffix(rest, "/kit")
		route = interviewRouteKit
	default:
		jobKey = rest
		route = interviewRouteWorkspace
	}
	jobKey = strings.Trim(jobKey, "/")
	if jobKey == "" {
		return "", 0, false
	}
	decoded, err := url.PathUnescape(jobKey)
	if err != nil {
		decoded = jobKey
	}
	return decoded, route, true
}
