package admin

// PaginationPageCount default pagination page count
var PaginationPageCount = 20

// Pagination is used to hold pagination related information when rendering tables
type Pagination struct {
	Total            int
	Pages            int
	CurrentPage      int
	PerPage          int
	UnlimitedEnabled bool
	// TODO: example (paginath by month)
	PageLabels       []string
}

func (this *Pagination) Unlimited() bool {
	return this.PerPage == -1
}
