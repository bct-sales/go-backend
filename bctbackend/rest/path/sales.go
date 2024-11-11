package path

type SalesPath struct{}

func Sales() *SalesPath {
	return &SalesPath{}
}

func (path *SalesPath) String() string {
	return "/api/v1/sales"
}
