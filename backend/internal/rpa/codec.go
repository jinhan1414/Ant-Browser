package rpa

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func mustJSON(value any, fallback string) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fallback
	}
	return string(data)
}

func decodeJSON(input string, target any) {
	if input == "" {
		return
	}
	_ = json.Unmarshal([]byte(input), target)
}

func parseFloat(input string) float64 {
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0
	}
	return value
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func mustJSONString(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func prettyJSON(value any) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}
