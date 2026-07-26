package contents

import (
	"encoding/json"

	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
	"github.com/tenSunFree/luma-lang-go/internal/http/datatransfers/responses"
)

func toListItem(r records.Content) responses.ContentListItemResponse {
	return responses.ContentListItemResponse{
		ID:          r.ID,
		Type:        r.ContentType,
		Title:       r.Title,
		Subtitle:    r.Subtitle,
		Description: r.Description,
		CoverURL:    r.CoverURL,
		DurationMs:  r.DurationMs,
		Level:       r.Level,
		Category:    r.Category,
		Tags:        []string(r.Tags),
		IsFree:      r.IsFree,
		ViewCount:   r.ViewCount,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toDetail(r records.Content) (responses.ContentDetailResponse, error) {
	var playback responses.ContentPlaybackResponse
	if err := json.Unmarshal(r.Playback, &playback); err != nil {
		return responses.ContentDetailResponse{}, err
	}
	var captions []responses.CaptionResponse
	if err := json.Unmarshal(r.Captions, &captions); err != nil {
		return responses.ContentDetailResponse{}, err
	}
	var vocabularyItems []responses.VocabularyItemResponse
	if err := json.Unmarshal(r.VocabularyItems, &vocabularyItems); err != nil {
		return responses.ContentDetailResponse{}, err
	}
	return responses.ContentDetailResponse{
		Content:         toListItem(r),
		Playback:        playback,
		CaptionsVersion: r.CaptionsVersion,
		Captions:        captions,
		VocabularyItems: vocabularyItems,
	}, nil
}
