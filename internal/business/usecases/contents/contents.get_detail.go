package contents

import (
	"context"
	"fmt"

	"github.com/tenSunFree/luma-lang-go/internal/apperror"
	"github.com/tenSunFree/luma-lang-go/internal/http/datatransfers/responses"
)

func (uc *usecase) GetDetail(ctx context.Context, id string) (responses.ContentDetailResponse, error) {
	row, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		// If NotFound is sent directly, the handler will return a 404 error.
		return responses.ContentDetailResponse{}, err
	}
	detail, err := toDetail(row)
	if err != nil {
		return responses.ContentDetailResponse{}, apperror.InternalCause(
			fmt.Errorf("contents.GetDetail marshal %s: %w", id, err),
		)
	}
	return detail, nil
}
