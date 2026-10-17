package main

import (
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
