package path

type categoryCountsPath struct{}

func CategoryCounts() *categoryCountsPath {
	return &categoryCountsPath{}
}

func (path *categoryCountsPath) String() string {
	return "/api/v1/category-counts"
}
