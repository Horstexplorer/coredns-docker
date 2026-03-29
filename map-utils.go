package docker

func Merge[K comparable, V any](m1 map[K][]V, m2 map[K][]V) map[K][]V {
	result := make(map[K][]V)

	for k, v := range m1 {
		result[k] = append([]V{}, v...)
	}

	for k, srcSlice := range m2 {
		if destSlice, exists := result[k]; exists {
			result[k] = append(destSlice, srcSlice...)
		} else {
			result[k] = append([]V{}, srcSlice...)
		}
	}

	return result
}
