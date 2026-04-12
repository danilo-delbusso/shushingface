package migrations

// V1LegacyToConnections migrates old provider formats to the connections slice.
func V1LegacyToConnections(data map[string]any) error {
	// Already has connections? Nothing to do.
	if conns, ok := data["connections"].([]any); ok && len(conns) > 0 {
		return nil
	}

	// Step 1: old multi-provider map → flat provider fields
	if providers, ok := data["providers"].(map[string]any); ok && len(providers) > 0 {
		provID := strOr(data, "transcriptionProviderId", "")
		if provID == "" {
			provID = strOr(data, "refinementProviderId", "")
		}
		if old, ok := providers[provID].(map[string]any); ok {
			name := strOr(old, "name", "groq")
			if name == "" {
				name = "groq"
			}
			data["providerId"] = name
			data["providerApiKey"] = strOr(old, "apiKey", "")
			data["providerBaseUrl"] = strOr(old, "baseUrl", "")
		}
		delete(data, "providers")
		delete(data, "transcriptionProviderId")
		delete(data, "refinementProviderId")
	}

	// Step 2: flat provider fields → connections slice
	provID := strOr(data, "providerId", "")
	if provID != "" {
		displayName := provID
		switch provID {
		case "groq":
			displayName = "Groq"
		}
		conn := map[string]any{
			"id":         "default",
			"name":       displayName,
			"providerId": provID,
			"apiKey":     strOr(data, "providerApiKey", ""),
			"baseUrl":    strOr(data, "providerBaseUrl", ""),
		}
		data["connections"] = []any{conn}
		data["transcriptionConnectionId"] = "default"
		data["refinementConnectionId"] = "default"
		delete(data, "providerId")
		delete(data, "providerApiKey")
		delete(data, "providerBaseUrl")
	}

	return nil
}
