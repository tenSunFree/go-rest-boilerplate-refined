package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/tenSunFree/luma-lang-go/internal/apperror"
	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

func (r *postgreLessonRepository) GetByID(ctx context.Context, id string) (records.Lesson, error) {
	const fileName = "lessons.get_by_id.go"

	var lesson records.Lesson
	err := r.conn.GetContext(ctx, &lesson, `
		SELECT
			id, title, subtitle, description, cover_url,
			duration_ms, level, category, tags, is_free,
			view_count, captions_version,
			playback, captions, vocabulary_items,
			created_at, updated_at
		FROM lessons
		WHERE id = $1
	`, id)

	if errors.Is(err, sql.ErrNoRows) {
		return records.Lesson{}, apperror.NotFound("lesson not found")
	}
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to get lesson by id", logger.Fields{
			"file":  fileName,
			"error": err.Error(),
			"id":    id,
		})
	}
	return lesson, err
}
