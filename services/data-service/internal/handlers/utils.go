package handlers

import "strconv"

// parseIntParam parses an integer parameter with a default value
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(param)
	if err != nil {
		return defaultValue
	}
	
	return value
}