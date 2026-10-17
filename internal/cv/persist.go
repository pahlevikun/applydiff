package cv

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const masterJSONComment = "Structured master CV (cv/master_cv.json, cv/master_cv.tex, cv/facts.json)."

// WriteMasterJSON writes the document to path with a stable _comment field.
func WriteMasterJSON(path string, doc *Document) error {
	if doc == nil {
		return fmt.Errorf("cv: document is nil")
	}
	if err := Validate(doc); err != nil {
		return err
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	out := make(map[string]json.RawMessage, len(fields)+1)
	comment, err := json.Marshal(masterJSONComment)
	if err != nil {
		return err
	}
	out["_comment"] = comment
	for k, v := range fields {
		out[k] = v
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
