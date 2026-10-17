package tui

import "github.com/pahlevikun/applydiff/internal/features"

func visibleScreens() []screen {
	list := []screen{screenFeed, screenJob, screenDiff, screenSrc, screenApps}
	if features.InterviewPrepEnabled {
		list = append(list, screenInterview)
	}
	return list
}

func stepScreen(cur screen, delta int) screen {
	list := visibleScreens()
	idx := 0
	for i, s := range list {
		if s == cur {
			idx = i
			break
		}
	}
	n := len(list)
	return list[(idx+delta%n+n)%n]
}
