package postgres

import (
	"context"

	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

func (r *postgreLessonRepository) List(ctx context.Context) ([]records.Lesson, error) {
	const fileName = "lessons.list.go"

	var lessons []records.Lesson
	err := r.conn.SelectContext(ctx, &lessons, `
		SELECT
			id, title, subtitle, description, cover_url,
			duration_ms, level, category, tags, is_free,
			view_count, captions_version,
			playback, captions, vocabulary_items,
			created_at, updated_at
		FROM lessons
		ORDER BY created_at DESC
	`)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list lessons", logger.Fields{
			"file":  fileName,
			"error": err.Error(),
		})
	}
	return lessons, err
}
