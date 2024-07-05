package testflow

import "github.com/qor5/admin/v3/utils/pregexp"

func RemoveTime(input string) string {
	// 2024-07-04T21:30:04.807186+08:00
	re := pregexp.MustCompile(`"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(|\.\d{1,6})\+\d{2}:\d{2}"`)
	return re.ReplaceAllString(input, "")
}

func EqualIgnoreTime(str1, str2 string) bool {
	return RemoveTime(str1) == RemoveTime(str2)
}
