package queries

func CollectTo[T any](receiver *[]T) func(T) error {
	return func(item T) error {
		*receiver = append(*receiver, item)
		return nil
	}
}
