package job

import "testing"

func TestJobFields(t *testing.T) {
	j := Job{ID: "1", Company: "Acme", Title: "Engineer"}
	if j.ID != "1" || j.Company != "Acme" {
		t.Fatalf("unexpected job: %+v", j)
	}
}
