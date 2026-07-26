package contents

import (
	"context"
	"fmt"

	"github.com/tenSunFree/luma-lang-go/internal/apperror"
	repointerface "github.com/tenSunFree/luma-lang-go/internal/datasources/repositories/interface"
	"github.com/tenSunFree/luma-lang-go/internal/http/datatransfers/responses"
)

func (uc *usecase) List(ctx context.Context, req ListRequest) (ListResult, error) {
	page := normalizePage(req.Page)
	limit := normalizeLimit(req.Limit)
	offset := (page - 1) * limit
	rows, total, err := uc.repo.List(ctx, repointerface.ContentListFilter{
		Type: req.Type,
	}, offset, limit)
	if err != nil {
		return ListResult{}, apperror.InternalCause(fmt.Errorf("contents.List: %w", err))
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
