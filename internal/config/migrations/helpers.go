package migrations

// strOr returns the string value for key in the map, or the fallback.
func strOr(m map[string]any, key, fallback string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return fallback
}
