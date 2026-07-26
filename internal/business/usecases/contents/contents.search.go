package contents

import (
	"context"
	"fmt"
	"strings"

	"github.com/tenSunFree/luma-lang-go/internal/apperror"
	repointerface "github.com/tenSunFree/luma-lang-go/internal/datasources/repositories/interface"
	"github.com/tenSunFree/luma-lang-go/internal/http/datatransfers/responses"
)

func (uc *usecase) Search(ctx context.Context, req SearchRequest) (ListResult, error) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return ListResult{}, apperror.BadRequest("search query is required")
	}
	page := normalizePage(req.Page)
	limit := normalizeLimit(req.Limit)
	offset := (page - 1) * limit
	rows, total, err := uc.repo.Search(ctx, repointerface.ContentListFilter{
		Type:  req.Type,
		Query: query,
	}, offset, limit)
	if err != nil {
		return ListResult{}, apperror.InternalCause(fmt.Errorf("contents.Search: %w", err))
	}
	items := make([]responses.ContentListItemResponse, 0, len(rows))
	for _, r := range rows {
		items = append(items, toListItem(r))
	}
	return ListResult{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}
