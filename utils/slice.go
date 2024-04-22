package utils

func GroupBySlice[T comparable, Kt comparable](slices []T, f func(T) Kt) (v [][]T) {

	var indexMap = make(map[Kt]int)
	for _, s := range slices {
		key := f(s)
		index, ok := indexMap[key]
		if !ok {
			v = append(v, []T{})
			index = len(v) - 1
			indexMap[key] = index
		}
		v[index] = append(v[index], s)
	}
	return
}
