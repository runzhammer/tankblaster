package helpers

type Number interface {
	int64 | float64 | int | uint64
}

func MakeRange[T Number](min, max T) []T {
	a := make([]T, int(max-min+1))
	for i := range a {
		a[i] = min + T(i)
	}
	return a
}
