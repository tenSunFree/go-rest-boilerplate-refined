package routes

import (
	"github.com/gin-gonic/gin"
	contentsuc "github.com/tenSunFree/luma-lang-go/internal/business/usecases/contents"
	contentshandler "github.com/tenSunFree/luma-lang-go/internal/http/handlers/v1/contents"
)

type contentsRoute struct {
	handler contentshandler.Handler
	router  *gin.RouterGroup
}

func NewContentsRoute(router *gin.RouterGroup, contentsUC contentsuc.Usecase) *contentsRoute {
	return &contentsRoute{
		handler: contentshandler.NewHandler(contentsUC),
		router:  router,
	}
}

func (r *contentsRoute) Routes() {
	v1 := r.router.Group("/v1")
	grp := v1.Group("/contents")
	// /search must precede /:contentId, otherwise Gin will treat "search" as contentId.
	grp.GET("", r.handler.ListContents)
	grp.GET("/search", r.handler.SearchContents)
	grp.GET("/:contentId", r.handler.GetContentDetail)
}
