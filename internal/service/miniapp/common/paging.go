package common

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 50
)

func NormalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = DefaultPage
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

func Skip(page, pageSize int) int64 {
	return int64((page - 1) * pageSize)
}
