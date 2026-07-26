package contents

import (
	"context"

	"github.com/tenSunFree/luma-lang-go/internal/http/datatransfers/responses"
)

// ListRequest list request parameters.
type ListRequest struct {
	Type  string
	Page  int
	Limit int
}

// SearchRequest parameters.
type SearchRequest struct {
	Query string
	Type  string
	Page  int
	Limit int
}

// ListResult is a list result with paginated information.
type ListResult struct {
	Items []responses.ContentListItemResponse
	Total int
	Page  int
	Limit int
}

// Usecase is the input boundary for contents.
type Usecase interface {
	List(ctx context.Context, req ListRequest) (ListResult, error)
	Search(ctx context.Context, req SearchRequest) (ListResult, error)
	GetDetail(ctx context.Context, id string) (responses.ContentDetailResponse, error)
}
