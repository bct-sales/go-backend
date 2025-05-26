package algorithms

type Set[T comparable] struct {
	elements map[T]struct{}
}

func NewSet[T comparable](initialElements ...T) Set[T] {
	result := Set[T]{elements: make(map[T]struct{})}

	for _, element := range initialElements {
		result.Add(element)
	}

	return result
}

func (s Set[T]) Add(element T) {
	s.elements[element] = struct{}{}
}

func (s Set[T]) Remove(element T) {
	delete(s.elements, element)
}

func (s Set[T]) Contains(element T) bool {
	_, exists := s.elements[element]
	return exists
}

func (s Set[T]) Len() int {
	return len(s.elements)
}

func (s Set[T]) ForEach(callback func(element T)) {
	for element := range s.elements {
		callback(element)
	}
}

func (s Set[T]) ToSlice() []T {
	result := make([]T, 0, len(s.elements))
	s.ForEach(func(element T) {
		result = append(result, element)
	})
	return result
}
