package profile

import "strings"

func ShortName(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	result := strings.ToUpper(string(runes[0:1]))
	if len(runes) > 2 {
		for i := 2; i < len(runes); i++ {
			if runes[i-1] == ' ' && runes[i] != ' ' {
				result += strings.ToUpper(string(runes[i : i+1]))
				break
			}
		}
	}
	return result
}
