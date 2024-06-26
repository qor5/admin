package pregexp

const patternApplyPathValues = `\{\s*([^\}]+?)\s*\}`

func ApplyPathValues(pat string, values map[string]string, onceIfDuplicated bool) string {
	used := map[string]bool{}
	return ReplaceAllSubmatchFunc(patternApplyPathValues, pat, func(match string, groups [][]int) string {
		key := match[groups[1][0]:groups[1][1]]
		if onceIfDuplicated && used[key] {
			return match
		}
		repl, ok := values[key]
		if ok {
			used[key] = true
			return repl
		}
		return match
	})
}
