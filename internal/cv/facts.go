package cv

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Facts struct {
	Facts map[string]string `json:"facts"`
}

func LoadFactsFile(path string) (*Facts, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFactsJSON(data)
}

func LoadFactsJSON(data []byte) (*Facts, error) {
	var f Facts
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.Facts == nil {
		f.Facts = map[string]string{}
	}
	return &f, nil
}

var numberTokenRe = regexp.MustCompile(`\d+(?:\.\d+)?%?`)

// GateNumberIn blocks any numeric token in text that is not present in approved facts values.
func (f *Facts) GateNumberIn(text string) error {
	if f == nil || len(f.Facts) == 0 {
		return fmt.Errorf("facts: no approved facts loaded")
	}
	allowed := make(map[string]struct{})
	for _, v := range f.Facts {
		for _, tok := range numberTokenRe.FindAllString(v, -1) {
			allowed[normalizeNumberToken(tok)] = struct{}{}
		}
	}
	for _, tok := range numberTokenRe.FindAllString(text, -1) {
		norm := normalizeNumberToken(tok)
		if _, ok := allowed[norm]; !ok {
			return fmt.Errorf("facts: unapproved number %q in text", tok)
		}
	}
	return nil
}

func normalizeNumberToken(tok string) string {
	tok = strings.TrimSpace(tok)
	if strings.HasSuffix(tok, "%") {
		return strings.TrimSuffix(tok, "%") + "%"
	}
	return tok
}
