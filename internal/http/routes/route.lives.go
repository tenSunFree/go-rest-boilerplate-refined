package routes

import (
	"github.com/gin-gonic/gin"
	livesuc "github.com/tenSunFree/luma-lang-go/internal/business/usecases/lives"
	liveshandler "github.com/tenSunFree/luma-lang-go/internal/http/handlers/v1/lives"
)

type livesRoute struct {
	handler        liveshandler.Handler
	router         *gin.RouterGroup
	authMiddleware gin.HandlerFunc
}

func NewLivesRoute(router *gin.RouterGroup, livesUC livesuc.Usecase, authMiddleware gin.HandlerFunc) *livesRoute {
	return &livesRoute{
		handler:        liveshandler.NewHandler(livesUC),
		router:         router,
		authMiddleware: authMiddleware,
	}
}

func (r *livesRoute) Routes() {
	v1 := r.router.Group("/v1")

	// Live stream list (reminder status can only be viewed after logging in)
	v1.GET("/live-courses", r.authMiddleware, r.handler.ListLiveCourses)

	// Student side
	lives := v1.Group("/lives/:liveId")
	lives.Use(r.authMiddleware)
	lives.POST("/join", r.handler.JoinLive)
	lives.POST("/leave", r.handler.LeaveLive)
	lives.POST("/renew-token", r.handler.RenewToken)
	lives.GET("/participants", r.handler.GetParticipants)
	lives.POST("/reminder", r.handler.SetReminder)
	lives.DELETE("/reminder", r.handler.DeleteReminder)

	// Teacher side (liveId path param is no longer needed)
	teacher := v1.Group("/teacher/lives")
	teacher.Use(r.authMiddleware)
	teacher.POST("/start", r.handler.StartTeacherLive)
	teacher.POST("/end", r.handler.EndTeacherLive)
}
