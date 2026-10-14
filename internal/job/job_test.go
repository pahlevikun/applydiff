package job

import "testing"

func TestJobFields(t *testing.T) {
	j := Job{ID: "1", Title: "Engineer"}
	if j.ID != "1" || j.Title != "Engineer" {
		t.Fatalf("unexpected job: %+v", j)
	}
}
