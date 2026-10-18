package common

// PageResponse 通用分页结构
// PageResponse[T any] 使用了 Go 1.18 引入的泛型语法
// T any 表示 T 可以是任意类型，和 Java 的 PageResponse<T> 效果一样
type PageResponse[T any] struct {
	Records            []T   `json:"records"`
	PageNumber         int64 `json:"pageNumber,string"`
	PageSize           int64 `json:"pageSize,string"`
	TotalPage          int64 `json:"totalPage,string"`
	TotalRow           int64 `json:"totalRow,string"`
	OptimizeCountQuery bool  `json:"optimizeCountQuery"`
}

func BuildPageResponse[T any](records []T, pageNum, pageSize, totalRow int64) PageResponse[T] {
	totalPage := int64(0)
	if pageSize > 0 {
		totalPage = (totalRow + pageSize - 1) / pageSize
	}
	return PageResponse[T]{
		Records:    records,
		PageNumber: pageNum,
		PageSize:   pageSize,
		TotalPage:  totalPage,
		TotalRow:   totalRow,
	}
}
