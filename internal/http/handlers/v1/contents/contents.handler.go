package contents

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	contentsuc "github.com/tenSunFree/luma-lang-go/internal/business/usecases/contents"
	v1 "github.com/tenSunFree/luma-lang-go/internal/http/handlers/v1"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

type Handler struct {
	usecase contentsuc.Usecase
}

func NewHandler(usecase contentsuc.Usecase) Handler {
	return Handler{usecase: usecase}
}

// ListContents godoc
// @Summary Content list / Get content list
// @Description Distinguishes content types by type, supports pagination. Returns all types if type is empty.
// @Tags contents
// @Produce json
// @Param type query string false "video | music | fairy_tale | column | supplement"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 20, max: 100)"
// @Success 200 {object} v1.BaseResponse
// @Router /contents [get]
func (h Handler) ListContents(ctx *gin.Context) {
	req := contentsuc.ListRequest{
		Type:  ctx.Query("type"),
		Page:  parseIntQuery(ctx, "page", 1),
		Limit: parseIntQuery(ctx, "limit", 20),
	}
	result, err := h.usecase.List(ctx.Request.Context(), req)
	if err != nil {
		logger.ErrorWithContext(ctx.Request.Context(), "ListContents failed", logger.Fields{
			"error": err.Error(),
			"type":  req.Type,
		})
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "contents fetched successfully", gin.H{
		"contents": result.Items,
		"total":    result.Total,
		"page":     result.Page,
		"limit":    result.Limit,
	})
}

// SearchContents godoc
// @Summary Search contents
// @Description Keyword search across title/subtitle/description/category, with optional narrowing type.
// @Tags contents
// @Produce json
// @Param q query string true "Search keywords"
// @Param type query string false "Narrowing type (optional)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 20, max: 100)"
// @Success 200 {object} v1.BaseResponse
// @Failure 400 {object} v1.BaseResponse "q is required"
// @Router /contents/search [get]
func (h Handler) SearchContents(ctx *gin.Context) {
	req := contentsuc.SearchRequest{
		Query: ctx.Query("q"),
		Type:  ctx.Query("type"),
		Page:  parseIntQuery(ctx, "page", 1),
		Limit: parseIntQuery(ctx, "limit", 20),
	}
	result, err := h.usecase.Search(ctx.Request.Context(), req)
	if err != nil {
		logger.ErrorWithContext(ctx.Request.Context(), "SearchContents failed", logger.Fields{
			"error": err.Error(),
			"q":     req.Query,
			"type":  req.Type,
		})
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "contents searched successfully", gin.H{
		"contents": result.Items,
		"total":    result.Total,
		"page":     result.Page,
		"limit":    result.Limit,
	})
}

// GetContentDetail godoc
// @Summary Content details / Get content detail
// @Description Returns complete content information: basic information, playback settings, subtitles, and single words.
// @Tags contents
// @Produce json
// @Param contentId path string true "Content ID (e.g. video_starbucks_001)"
// @Success 200 {object} v1.BaseResponse
// @Failure 404 {object} v1.BaseResponse
// @Router /contents/{contentId} [get]
func (h Handler) GetContentDetail(ctx *gin.Context) {
	id := ctx.Param("contentId")
	detail, err := h.usecase.GetDetail(ctx.Request.Context(), id)
	if err != nil {
		logger.ErrorWithContext(ctx.Request.Context(), "GetContentDetail failed", logger.Fields{
			"error": err.Error(),
			"id":    id,
		})
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "content detail fetched successfully", detail)
}

func parseIntQuery(ctx *gin.Context, key string, fallback int) int {
	raw := ctx.Query(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return fallback
	}
	return v
}
