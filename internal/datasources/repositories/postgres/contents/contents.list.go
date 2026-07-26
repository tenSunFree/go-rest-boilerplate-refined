package postgres

import (
	"context"
	"fmt"

	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
	repointerface "github.com/tenSunFree/luma-lang-go/internal/datasources/repositories/interface"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

func (r *postgreContentRepository) List(
	ctx context.Context,
	filter repointerface.ContentListFilter,
	offset, limit int,
) ([]records.Content, int, error) {
	const fileName = "contents.list.go"
	// COUNT and SELECT share the same WHERE condition
	var total int
	if err := r.conn.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM contents
		WHERE ($1 = '' OR content_type = $1)
	`, filter.Type); err != nil {
		return nil, 0, fmt.Errorf("contents.List count: %w", err)
	}
	var contents []records.Content
	err := r.conn.SelectContext(ctx, &contents, `
		SELECT
			id, content_type, title, subtitle, description, cover_url,
			duration_ms, level, category, tags, is_free,
			view_count, captions_version,
			playback, captions, vocabulary_items,
			created_at, updated_at
		FROM contents
		WHERE ($1 = '' OR content_type = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, filter.Type, limit, offset)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list contents", logger.Fields{
			"file":  fileName,
			"error": err.Error(),
			"type":  filter.Type,
		})
		return nil, 0, err
	}
	return contents, total, nil
}
