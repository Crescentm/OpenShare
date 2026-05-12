package pagination

type Page struct {
	Page     int
	PageSize int
	Offset   int
	Limit    int
}

func Normalize(page, pageSize, defaultPage, defaultPageSize, maxPageSize int) (Page, bool) {
	return NormalizeWindow(page, pageSize, defaultPage, defaultPageSize, maxPageSize, 0)
}

func NormalizeWindow(page, pageSize, defaultPage, defaultPageSize, maxPageSize, resultWindow int) (Page, bool) {
	if page == 0 {
		page = defaultPage
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	if page < 1 || pageSize < 1 || pageSize > maxPageSize {
		return Page{}, false
	}
	if resultWindow > 0 && page*pageSize > resultWindow {
		return Page{}, false
	}
	return Page{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
		Limit:    pageSize,
	}, true
}
