package set

type Set[T comparable] map[T]bool

func NewSet[T comparable]() Set[T] {
	return Set[T]{}
}

func FromSlice[T comparable](slc []T) Set[T] {
	s := Set[T]{}
	for _, item := range slc {
		s[item] = true
	}

	return s
}