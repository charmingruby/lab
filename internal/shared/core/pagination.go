package core

const (
	DefaultPageSize = 25
	MaxPageSize     = 100
)

type PaginationParams struct {
	Page  int
	Limit int
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

func (p PaginationParams) TotalPages(total int) int {
	if total == 0 {
		return 0
	}

	pages := total / p.Limit
	if total%p.Limit != 0 {
		pages++
	}

	return pages
}

func (p PaginationParams) Validate() PaginationParams {
	if p.Page < 1 {
		p.Page = 1
	}

	if p.Limit < 1 {
		p.Limit = DefaultPageSize
	}

	if p.Limit > MaxPageSize {
		p.Limit = MaxPageSize
	}

	return p
}
