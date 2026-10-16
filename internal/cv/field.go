package cv

import "fmt"

// Field returns the string value for a stable field id (for diff targeting).
func (doc *Document) Field(id string) (string, bool) {
	if doc == nil {
		return "", false
	}
	switch id {
	case "contact.name":
		return doc.Contact.Name, true
	case "contact.headline":
		return doc.Contact.Headline, true
	case "contact.email":
		return doc.Contact.Email, true
	case "contact.site":
		return doc.Contact.Site, true
	case "contact.linkedin":
		return doc.Contact.LinkedIn, true
	case "contact.github":
		return doc.Contact.GitHub, true
	case "contact.location":
		return doc.Contact.Location, true
	case "contact.timezone_line":
		return doc.Contact.TimezoneLine, true
	}
	if doc.Summary.ID == id {
		return doc.Summary.Text, true
	}
	for _, sg := range doc.Skills {
		if sg.ID == id {
			return joinItems(sg.Items), true
		}
	}
	for _, exp := range doc.Experience {
		if exp.ID == id {
			return fmt.Sprintf("%s — %s", exp.Company, exp.Title), true
		}
		for _, b := range exp.Bullets {
			if b.ID == id {
				return b.Text, true
			}
		}
	}
	for _, ed := range doc.Education {
		if ed.ID == id {
			return fmt.Sprintf("%s, %s (%s)", ed.Degree, ed.School, ed.Year), true
		}
	}
	for _, ach := range doc.Achievements {
		if ach.ID == id {
			return ach.Text, true
		}
	}
	return "", false
}

// SetField updates a field by id. Used when applying accepted hunks.
func (doc *Document) SetField(id, value string) error {
	switch id {
	case "contact.name":
		doc.Contact.Name = value
		return nil
	case "contact.headline":
		doc.Contact.Headline = value
		return nil
	case "contact.email":
		doc.Contact.Email = value
		return nil
	case "contact.site":
		doc.Contact.Site = value
		return nil
	case "contact.linkedin":
		doc.Contact.LinkedIn = value
		return nil
	case "contact.github":
		doc.Contact.GitHub = value
		return nil
	case "contact.location":
		doc.Contact.Location = value
		return nil
	case "contact.timezone_line":
		doc.Contact.TimezoneLine = value
		return nil
	}
	if doc.Summary.ID == id {
		doc.Summary.Text = value
		return nil
	}
	for i := range doc.Skills {
		if doc.Skills[i].ID == id {
			doc.Skills[i].Items = splitItems(value)
			return nil
		}
	}
	for i := range doc.Experience {
		if doc.Experience[i].ID == id {
			return fmt.Errorf("cv: cannot set composite experience field %q", id)
		}
		for j := range doc.Experience[i].Bullets {
			if doc.Experience[i].Bullets[j].ID == id {
				doc.Experience[i].Bullets[j].Text = value
				return nil
			}
		}
	}
	for i := range doc.Education {
		if doc.Education[i].ID == id {
			return fmt.Errorf("cv: cannot set composite education field %q", id)
		}
	}
	for i := range doc.Achievements {
		if doc.Achievements[i].ID == id {
			doc.Achievements[i].Text = value
			return nil
		}
	}
	return fmt.Errorf("cv: unknown field id %q", id)
}

func joinItems(items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for i := 1; i < len(items); i++ {
		out += ", " + items[i]
	}
	return out
}

func splitItems(s string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	for _, p := range splitOnComma(s) {
		p = trimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitOnComma(s string) []string {
	var out []string
	var cur string
	for _, r := range s {
		if r == ',' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	out = append(out, cur)
	return out
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
