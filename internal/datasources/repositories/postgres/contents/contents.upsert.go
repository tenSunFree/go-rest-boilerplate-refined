package postgres

import (
	"context"

	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

func (r *postgreContentRepository) Upsert(ctx context.Context, content records.Content) error {
	const fileName = "contents.upsert.go"
	_, err := r.conn.NamedExecContext(ctx, `
		INSERT INTO contents (
			id, content_type, title, subtitle, description, cover_url,
			duration_ms, level, category, tags, is_free,
			view_count, captions_version,
			playback, captions, vocabulary_items,
			created_at, updated_at
		) VALUES (
			:id, :content_type, :title, :subtitle, :description, :cover_url,
			:duration_ms, :level, :category, :tags, :is_free,
			:view_count, :captions_version,
			:playback, :captions, :vocabulary_items,
			:created_at, :updated_at
		)
		ON CONFLICT (id) DO UPDATE SET
			content_type     = EXCLUDED.content_type,
			title            = EXCLUDED.title,
			subtitle         = EXCLUDED.subtitle,
			description      = EXCLUDED.description,
			cover_url        = EXCLUDED.cover_url,
			duration_ms      = EXCLUDED.duration_ms,
			level            = EXCLUDED.level,
			category         = EXCLUDED.category,
			tags             = EXCLUDED.tags,
			is_free          = EXCLUDED.is_free,
			view_count       = EXCLUDED.view_count,
			captions_version = EXCLUDED.captions_version,
			playback         = EXCLUDED.playback,
			captions         = EXCLUDED.captions,
			vocabulary_items = EXCLUDED.vocabulary_items,
			updated_at       = EXCLUDED.updated_at
	`, content)
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to upsert content", logger.Fields{
			"file":  fileName,
			"error": err.Error(),
			"id":    content.ID,
			"type":  content.ContentType,
		})
	}
	return err
}
