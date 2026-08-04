package core

const pageSize = 25

type PaginationParams struct {
	Page     int
	PageSize int
}

func DefaultPaginationParams(page int) PaginationParams {
	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}
