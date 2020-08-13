package admin

// Page contain pagination information
type Page struct {
	Page       int
	Current    bool
	IsPrevious bool
	IsNext     bool
	IsFirst    bool
	IsLast     bool
}

type PaginationResult struct {
	Pagination     Pagination
	Pages          []Page
	ShowAllEnabled bool
}

const visiblePageCount = 8

// Pagination return pagination information
// Keep visiblePageCount's pages visible, exclude prev and next link
// Assume there are 12 pages in total.
// When current page is 1
// [current, 2, 3, 4, 5, 6, 7, 8, next]
// When current page is 6
// [prev, 2, 3, 4, 5, current, 7, 8, 9, 10, next]
// When current page is 10
// [prev, 5, 6, 7, 8, 9, current, 11, 12]
// If total page count less than VISIBLE_PAGE_COUNT, always show all pages
func (this *Context) Pagination() *PaginationResult {
	var (
		pages      []Page
		pagination = this.Searcher.Pagination
		pageCount  = pagination.PerPage
	)

	if pageCount == 0 {
		if this.Resource != nil && this.Resource.Config.PageCount != 0 {
			pageCount = this.Resource.Config.PageCount
		} else {
			pageCount = PaginationPageCount
		}
	}

	if pagination.Total <= pageCount && pagination.CurrentPage <= 1 {
		return nil
	}

	start := pagination.CurrentPage - visiblePageCount/2
	if start < 1 {
		start = 1
	}

	end := start + visiblePageCount - 1 // -1 for "start page" itself
	if end > pagination.Pages {
		end = pagination.Pages
	}

	if (end-start) < visiblePageCount && start != 1 {
		start = end - visiblePageCount + 1
	}
	if start < 1 {
		start = 1
	}

	// Append prev link
	if start > 1 {
		pages = append(pages, Page{Page: 1, IsFirst: true})
		pages = append(pages, Page{Page: pagination.CurrentPage - 1, IsPrevious: true})
	}

	for i := start; i <= end; i++ {
		pages = append(pages, Page{Page: i, Current: pagination.CurrentPage == i})
	}

	// Append next link
	if end < pagination.Pages {
		pages = append(pages, Page{Page: pagination.CurrentPage + 1, IsNext: true})
		pages = append(pages, Page{Page: pagination.Pages, IsLast: true})
	}

	return &PaginationResult{Pagination: pagination, Pages: pages}
}
