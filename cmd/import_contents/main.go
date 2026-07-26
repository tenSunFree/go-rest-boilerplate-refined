package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/lib/pq"
	"github.com/tenSunFree/luma-lang-go/internal/config"
	"github.com/tenSunFree/luma-lang-go/internal/constants"
	"github.com/tenSunFree/luma-lang-go/internal/datasources/drivers"
	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
	contentspostgres "github.com/tenSunFree/luma-lang-go/internal/datasources/repositories/postgres/contents"
	"github.com/tenSunFree/luma-lang-go/internal/http/datatransfers/responses"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

type ContentImportItem struct {
	Content         responses.ContentListItemResponse  `json:"content"`
	Playback        responses.ContentPlaybackResponse  `json:"playback"`
	CaptionsVersion int                                `json:"captionsVersion"`
	Captions        []responses.CaptionResponse        `json:"captions"`
	VocabularyItems []responses.VocabularyItemResponse `json:"vocabularyItems"`
}

func init() {
	if err := config.InitializeAppConfig(); err != nil {
		logger.Fatal(err.Error(), logger.Fields{
			constants.LoggerCategory: constants.LoggerCategoryConfig,
		})
	}
}

func main() {
	filePath := flag.String("file", "data/contents.json", "path to contents JSON file")
	flag.Parse()
	raw, err := os.ReadFile(*filePath)
	if err != nil {
		logger.Fatal("cannot read file: "+err.Error(), logger.Fields{
			constants.LoggerCategory: "import_contents",
			"path":                   *filePath,
		})
	}
	var items []ContentImportItem
	if err := json.Unmarshal(raw, &items); err != nil {
		logger.Fatal("cannot parse JSON: "+err.Error(), logger.Fields{
			constants.LoggerCategory: "import_contents",
		})
	}
	logger.Info("file loaded", logger.Fields{
		constants.LoggerCategory: "import_contents",
		"count":                  len(items),
	})
	db, err := drivers.SetupSQLXPostgres()
	if err != nil {
		logger.Fatal("db connect failed: "+err.Error(), logger.Fields{
			constants.LoggerCategory: "import_contents",
		})
	}
	defer func() { _ = db.Close() }()
	repo := contentspostgres.NewContentRepository(db)
	ctx := context.Background()
	for _, item := range items {
		rec, err := toRecord(item)
		if err != nil {
			logger.Fatal("convert failed: "+err.Error(), logger.Fields{
				constants.LoggerCategory: "import_contents",
				"content_id":             item.Content.ID,
			})
		}
		if err := repo.Upsert(ctx, rec); err != nil {
			logger.Fatal("upsert failed: "+err.Error(), logger.Fields{
				constants.LoggerCategory: "import_contents",
				"content_id":             item.Content.ID,
			})
		}
		logger.Info("upserted", logger.Fields{
			constants.LoggerCategory: "import_contents",
			"content_id":             item.Content.ID,
			"type":                   item.Content.Type,
		})
	}
	logger.Info("import complete", logger.Fields{
		constants.LoggerCategory: "import_contents",
		"total":                  len(items),
	})
}

func toRecord(item ContentImportItem) (records.Content, error) {
	playback, err := json.Marshal(item.Playback)
	if err != nil {
		return records.Content{}, err
	}
	captions, err := json.Marshal(item.Captions)
	if err != nil {
		return records.Content{}, err
	}
	vocabularyItems, err := json.Marshal(item.VocabularyItems)
	if err != nil {
		return records.Content{}, err
	}
	createdAt, err := time.Parse(time.RFC3339, item.Content.CreatedAt)
	if err != nil {
		return records.Content{}, err
	}
	updatedAt, err := time.Parse(time.RFC3339, item.Content.UpdatedAt)
	if err != nil {
		return records.Content{}, err
	}
	return records.Content{
		ID:              item.Content.ID,
		ContentType:     item.Content.Type,
		Title:           item.Content.Title,
		Subtitle:        item.Content.Subtitle,
		Description:     item.Content.Description,
		CoverURL:        item.Content.CoverURL,
		DurationMs:      item.Content.DurationMs,
		Level:           item.Content.Level,
		Category:        item.Content.Category,
		Tags:            pq.StringArray(item.Content.Tags),
		IsFree:          item.Content.IsFree,
		ViewCount:       item.Content.ViewCount,
		CaptionsVersion: item.CaptionsVersion,
		Playback:        json.RawMessage(playback),
		Captions:        json.RawMessage(captions),
		VocabularyItems: json.RawMessage(vocabularyItems),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}
