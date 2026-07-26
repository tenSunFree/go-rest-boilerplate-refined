package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/tenSunFree/luma-lang-go/internal/apperror"
	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

func (r *postgreContentRepository) GetByID(ctx context.Context, id string) (records.Content, error) {
	const fileName = "contents.get_by_id.go"
	var content records.Content
	err := r.conn.GetContext(ctx, &content, `
		SELECT
			id, content_type, title, subtitle, description, cover_url,
			duration_ms, level, category, tags, is_free,
			view_count, captions_version,
			playback, captions, vocabulary_items,
			created_at, updated_at
		FROM contents
		WHERE id = $1
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return records.Content{}, apperror.NotFound("content not found")
	}
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to get content by id", logger.Fields{
			"file":  fileName,
			"error": err.Error(),
			"id":    id,
		})
		return records.Content{}, err
	}
	return content, nil
}
