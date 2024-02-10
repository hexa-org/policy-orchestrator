package functionalsupport

func Map[T any, R any](list []T, transform func(T) R) []R {
	var results []R
	for _, t := range list {
		results = append(results, transform(t))
	}
	return results
}
