package postgres

import (
	"context"

	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

func (r *postgreLessonRepository) Upsert(ctx context.Context, lesson records.Lesson) error {
	const fileName = "lessons.upsert.go"
	_, err := r.conn.NamedExecContext(ctx, `
		INSERT INTO lessons (
			id, title, subtitle, description, cover_url,
			duration_ms, level, category, tags, is_free,
			view_count, captions_version,
			playback, captions, vocabulary_items,
			created_at, updated_at
		) VALUES (
			:id, :title, :subtitle, :description, :cover_url,
			:duration_ms, :level, :category, :tags, :is_free,
			:view_count, :captions_version,
			:playback, :captions, :vocabulary_items,
			:created_at, :updated_at
		)
		ON CONFLICT (id) DO UPDATE SET
			title            = EXCLUDED.title,
			subtitle         = EXCLUDED.subtitle,
			description      = EXCLUDED.description,
			cover_url        = EXCLUDED.cover_url,
			duration_ms      = EXCLUDED.duration_ms,
			level            = EXCLUDED.level,
			category         = EXCLUDED.category,
			tags             = EXCLUDED.tags,
			is_free          = EXCLUDED.is_free,
			captions_version = EXCLUDED.captions_version,
			playback         = EXCLUDED.playback,
			captions         = EXCLUDED.captions,
			vocabulary_items = EXCLUDED.vocabulary_items,
			updated_at       = EXCLUDED.updated_at
	`, lesson)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to upsert lesson", logger.Fields{
			"file":      fileName,
			"error":     err.Error(),
			"lesson_id": lesson.ID,
		})
	}
	return err
}
