package cv

// FieldView is one editable CV field for host-agent tailoring.
type FieldView struct {
	FieldID string `json:"field_id"`
	Section string `json:"section"`
	Text    string `json:"text"`
}

// ListFieldViews returns stable field ids with current text for tailoring context.
func (doc *Document) ListFieldViews() []FieldView {
	if doc == nil {
		return nil
	}
	var out []FieldView
	add := func(id, section, text string) {
		if text == "" && id == "" {
			return
		}
		out = append(out, FieldView{FieldID: id, Section: section, Text: text})
	}
	add("contact.name", "contact", doc.Contact.Name)
	add("contact.headline", "contact", doc.Contact.Headline)
	add("contact.email", "contact", doc.Contact.Email)
	add("contact.site", "contact", doc.Contact.Site)
	add("contact.linkedin", "contact", doc.Contact.LinkedIn)
	add("contact.github", "contact", doc.Contact.GitHub)
	add("contact.location", "contact", doc.Contact.Location)
	add("contact.timezone_line", "contact", doc.Contact.TimezoneLine)
	if doc.Summary.ID != "" {
		add(doc.Summary.ID, "summary", doc.Summary.Text)
	}
	for _, sg := range doc.Skills {
		add(sg.ID, "skills", joinItems(sg.Items))
	}
	for _, exp := range doc.Experience {
		for _, b := range exp.Bullets {
			add(b.ID, "experience", b.Text)
		}
	}
	for _, ach := range doc.Achievements {
		add(ach.ID, "achievements", ach.Text)
	}
	return out
}
