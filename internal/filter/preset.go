package filter

// AggressivePreset is a built-in filter for very fresh, lower-competition postings.
func AggressivePreset() Spec {
	return Spec{
		PostedWithin: &IntRule{Hard: true, Max: 5},
		MaxApplicants: &IntRule{Hard: true, Max: 50},
	}
}
