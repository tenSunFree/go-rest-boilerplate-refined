package lives

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	livesuc "github.com/tenSunFree/luma-lang-go/internal/business/usecases/lives"
	httpauth "github.com/tenSunFree/luma-lang-go/internal/http/auth"
	"github.com/tenSunFree/luma-lang-go/internal/http/datatransfers/requests"
	"github.com/tenSunFree/luma-lang-go/internal/http/datatransfers/responses"
	v1 "github.com/tenSunFree/luma-lang-go/internal/http/handlers/v1"
)

type Handler struct {
	usecase livesuc.Usecase
}

func NewHandler(usecase livesuc.Usecase) Handler {
	return Handler{usecase: usecase}
}

// ListLiveCourses godoc
// @Summary      List live courses
// @Description  Returns scheduled and/or live courses. Pass status as comma-separated values, e.g. status=scheduled,live. Reminder flag reflects the current user's reminder state.
// @Tags         lives
// @Produce      json
// @Param        status  query     string  false  "scheduled | live | ended | cancelled  (comma-separated, default: scheduled,live)"
// @Success      200     {object}  v1.BaseResponse{data=object{data=[]responses.LiveCourseResponse}}
// @Failure      401     {object}  v1.BaseResponse
// @Security     BearerAuth
// @Router       /live-courses [get]
func (h Handler) ListLiveCourses(ctx *gin.Context) {
	user, _ := httpauth.CurrentUserFromContext(ctx)

	statusParam := ctx.Query("status")
	var statuses []string
	if statusParam != "" {
		statuses = strings.Split(statusParam, ",")
	}

	resp, err := h.usecase.ListLiveCourses(ctx.Request.Context(), livesuc.ListLiveCoursesRequest{
		UserID:   user.ID,
		Statuses: statuses,
	})
	if err != nil {
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "live courses fetched successfully", gin.H{
		"data": responses.FromLiveCourses(resp.Items),
	})
}

// JoinLive godoc
// @Summary      Student joins a live
// @Description  Allocates a student Agora UID (10000+), generates an RTC token, records the participant, and returns the full join config including teacher stream UIDs and WebSocket chat URL.
// @Tags         lives
// @Accept       json
// @Produce      json
// @Param        liveId   path      string                   true  "Live ID"
// @Param        request  body      requests.JoinLiveRequest false "Optional client metadata"
// @Success      200      {object}  v1.BaseResponse{data=object{liveId=string,courseId=string,agora=object,teacher=object,streamLayout=object,chat=object,features=object}}
// @Failure      400      {object}  v1.BaseResponse  "Malformed JSON"
// @Failure      401      {object}  v1.BaseResponse
// @Failure      403      {object}  v1.BaseResponse  "Live not started yet"
// @Failure      404      {object}  v1.BaseResponse  "Live not found"
// @Security     BearerAuth
// @Router       /lives/{liveId}/join [post]
func (h Handler) JoinLive(ctx *gin.Context) {
	user, err := httpauth.CurrentUserFromContext(ctx)
	if err != nil {
		v1.NewAbortResponse(ctx, err.Error())
		return
	}
	var req requests.JoinLiveRequest
	_ = ctx.ShouldBindJSON(&req)

	resp, err := h.usecase.JoinLive(ctx.Request.Context(), livesuc.JoinLiveRequest{
		LiveID:      ctx.Param("liveId"),
		UserID:      user.ID,
		DisplayName: user.Email,
		ClientType:  req.ClientType,
	})
	if err != nil {
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "joined successfully", toJoinResponse(resp))
}

// LeaveLive godoc
// @Summary      Student leaves a live
// @Description  Marks the participant as left and records watch duration.
// @Tags         lives
// @Accept       json
// @Produce      json
// @Param        liveId   path      string                    true  "Live ID"
// @Param        request  body      requests.LeaveLiveRequest true  "UID is required"
// @Success      200      {object}  v1.BaseResponse{data=object{liveId=string}}
// @Failure      400      {object}  v1.BaseResponse  "Missing uid"
// @Failure      401      {object}  v1.BaseResponse
// @Security     BearerAuth
// @Router       /lives/{liveId}/leave [post]
func (h Handler) LeaveLive(ctx *gin.Context) {
	user, err := httpauth.CurrentUserFromContext(ctx)
	if err != nil {
		v1.NewAbortResponse(ctx, err.Error())
		return
	}
	var req requests.LeaveLiveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.NewErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.usecase.LeaveLive(ctx.Request.Context(), livesuc.LeaveLiveRequest{
		LiveID: ctx.Param("liveId"),
		UserID: user.ID,
		UID:    req.UID,
	}); err != nil {
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "left successfully", gin.H{"liveId": ctx.Param("liveId")})
}

// RenewToken godoc
// @Summary      Renew an Agora RTC token (student)
// @Description  Called when the Agora SDK fires token-privilege-will-expire. Pass the same uid that was returned by /join.
// @Tags         lives
// @Accept       json
// @Produce      json
// @Param        liveId   path      string                     true  "Live ID"
// @Param        request  body      requests.RenewTokenRequest true  "UID to renew"
// @Success      200      {object}  v1.BaseResponse{data=object{uid=int,rtcToken=string,tokenExpireAt=string}}
// @Failure      400      {object}  v1.BaseResponse  "Missing uid"
// @Failure      401      {object}  v1.BaseResponse
// @Failure      404      {object}  v1.BaseResponse  "Live not found"
// @Security     BearerAuth
// @Router       /lives/{liveId}/renew-token [post]
func (h Handler) RenewToken(ctx *gin.Context) {
	user, err := httpauth.CurrentUserFromContext(ctx)
	if err != nil {
		v1.NewAbortResponse(ctx, err.Error())
		return
	}
	_ = user
	var req requests.RenewTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.NewErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.usecase.RenewToken(ctx.Request.Context(), livesuc.RenewTokenRequest{
		LiveID: ctx.Param("liveId"),
		UID:    req.UID,
		Role:   livesuc.RoleAudience,
	})
	if err != nil {
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "token renewed", gin.H{
		"uid":           resp.UID,
		"rtcToken":      resp.RTCToken,
		"tokenExpireAt": resp.TokenExpireAt,
	})
}

// GetParticipants godoc
// @Summary      Get participants list
// @Description  Returns all online participants (left_at IS NULL). Teacher camera and screen streams are merged into a single participant entry. Teacher stream UIDs are returned separately in teacherStreams.
// @Tags         lives
// @Produce      json
// @Param        liveId  path      string  true  "Live ID"
// @Success      200     {object}  v1.BaseResponse{data=object{totalCount=int,teacherStreams=object,participants=array}}
// @Failure      401     {object}  v1.BaseResponse
// @Failure      404     {object}  v1.BaseResponse  "Live not found"
// @Security     BearerAuth
// @Router       /lives/{liveId}/participants [get]
func (h Handler) GetParticipants(ctx *gin.Context) {
	resp, err := h.usecase.GetParticipants(ctx.Request.Context(), livesuc.GetParticipantsRequest{
		LiveID: ctx.Param("liveId"),
	})
	if err != nil {
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "participants fetched", toParticipantsResponse(resp))
}

// SetReminder godoc
// @Summary      Enable reminder for a live
// @Description  Creates or updates a reminder set to 10 minutes before scheduledStartAt.
// @Tags         lives
// @Produce      json
// @Param        liveId  path      string  true  "Live ID"
// @Success      200     {object}  v1.BaseResponse{data=object{liveId=string,isReminderEnabled=bool}}
// @Failure      401     {object}  v1.BaseResponse
// @Failure      404     {object}  v1.BaseResponse  "Live not found"
// @Security     BearerAuth
// @Router       /lives/{liveId}/reminder [post]
func (h Handler) SetReminder(ctx *gin.Context) {
	user, err := httpauth.CurrentUserFromContext(ctx)
	if err != nil {
		v1.NewAbortResponse(ctx, err.Error())
		return
	}
	resp, err := h.usecase.SetReminder(ctx.Request.Context(), livesuc.SetReminderRequest{
		LiveID: ctx.Param("liveId"),
		UserID: user.ID,
	})
	if err != nil {
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "reminder enabled", toReminderResponse(resp))
}

// DeleteReminder godoc
// @Summary      Disable reminder for a live
// @Description  Removes the reminder for the current user.
// @Tags         lives
// @Produce      json
// @Param        liveId  path      string  true  "Live ID"
// @Success      200     {object}  v1.BaseResponse{data=object{liveId=string,isReminderEnabled=bool}}
// @Failure      401     {object}  v1.BaseResponse
// @Failure      404     {object}  v1.BaseResponse  "Live not found"
// @Security     BearerAuth
// @Router       /lives/{liveId}/reminder [delete]
func (h Handler) DeleteReminder(ctx *gin.Context) {
	user, err := httpauth.CurrentUserFromContext(ctx)
	if err != nil {
		v1.NewAbortResponse(ctx, err.Error())
		return
	}
	resp, err := h.usecase.DeleteReminder(ctx.Request.Context(), livesuc.DeleteReminderRequest{
		LiveID: ctx.Param("liveId"),
		UserID: user.ID,
	})
	if err != nil {
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "reminder disabled", toReminderResponse(resp))
}

// StartTeacherLive godoc
// @Summary      Teacher creates a live room and starts broadcasting
// @Description  One-shot API: creates the live_course record, sets status to live, and returns Agora RTC tokens for camera (uid=1000) and screen (uid=2000). If the teacher already has an active live, returns a new token for that live instead (idempotent).
// @Tags         lives-teacher
// @Accept       json
// @Produce      json
// @Param        request  body      requests.StartTeacherLiveRequest  true  "Live room info"
// @Success      200      {object}  v1.BaseResponse{data=object{liveId=string,courseId=string,agora=object,streams=object}}
// @Failure      400      {object}  v1.BaseResponse  "Missing required fields"
// @Failure      401      {object}  v1.BaseResponse
// @Security     BearerAuth
// @Router       /teacher/lives/start [post]
func (h Handler) StartTeacherLive(ctx *gin.Context) {
	user, err := httpauth.CurrentUserFromContext(ctx)
	if err != nil {
		v1.NewAbortResponse(ctx, err.Error())
		return
	}

	var req requests.StartTeacherLiveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.NewErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.usecase.StartTeacherLive(ctx.Request.Context(), livesuc.StartTeacherLiveRequest{
		TeacherID:    user.ID,
		TeacherName:  user.Email, // Use email for now, you can change it later to retrieve username from the users table.
		Title:        req.Title,
		Category:     req.Category,
		Level:        req.Level,
		CourseType:   req.CourseType,
		ThumbnailURL: req.ThumbnailURL,
		TextbookURL:  req.TextbookURL,
	})
	if err != nil {
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "live started", toStartResponse(resp))
}

// EndTeacherLive godoc
// @Summary      Teacher ends current live session
// @Description  Ends the teacher's currently active live session (status=live). No body needed; the backend finds the active live by teacher identity from the JWT.
// @Tags         lives-teacher
// @Produce      json
// @Success      200  {object}  v1.BaseResponse{data=object{liveId=string}}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse  "No active live found"
// @Security     BearerAuth
// @Router       /teacher/lives/end [post]
func (h Handler) EndTeacherLive(ctx *gin.Context) {
	user, err := httpauth.CurrentUserFromContext(ctx)
	if err != nil {
		v1.NewAbortResponse(ctx, err.Error())
		return
	}

	liveID, err := h.usecase.EndTeacherLive(ctx.Request.Context(), livesuc.EndTeacherLiveRequest{
		TeacherID: user.ID,
	})
	if err != nil {
		v1.RespondWithError(ctx, err)
		return
	}
	v1.NewSuccessResponse(ctx, http.StatusOK, "live ended", gin.H{"liveId": liveID})
}

// Response conversion helpers
func toJoinResponse(r livesuc.JoinLiveResponse) gin.H {
	return gin.H{
		"liveId":   r.LiveID,
		"courseId": r.CourseID,
		"agora": gin.H{
			"appId":         r.Agora.AppID,
			"channelName":   r.Agora.ChannelName,
			"uid":           r.Agora.UID,
			"rtcToken":      r.Agora.RTCToken,
			"role":          r.Agora.Role,
			"tokenExpireAt": r.Agora.TokenExpireAt,
		},
		"teacher": gin.H{
			"teacherId": r.Teacher.TeacherID,
			"name":      r.Teacher.Name,
			"avatarUrl": r.Teacher.AvatarURL,
			"cameraUid": r.Teacher.CameraUID,
			"screenUid": r.Teacher.ScreenUID,
		},
		"streamLayout": gin.H{
			"mainUid":  r.StreamLayout.MainUID,
			"mainType": r.StreamLayout.MainType,
			"pipUid":   r.StreamLayout.PipUID,
			"pipType":  r.StreamLayout.PipType,
		},
		"chat": gin.H{
			"enabled":  r.Chat.Enabled,
			"provider": r.Chat.Provider,
			"roomId":   r.Chat.RoomID,
			"wsUrl":    r.Chat.WSUrl,
		},
		"features": gin.H{
			"canSendMessage":  r.Features.CanSendMessage,
			"canRaiseHand":    r.Features.CanRaiseHand,
			"canSendReaction": r.Features.CanSendReaction,
			"canPublishAudio": r.Features.CanPublishAudio,
			"canPublishVideo": r.Features.CanPublishVideo,
		},
	}
}

func toStartResponse(r livesuc.StartTeacherLiveResponse) gin.H {
	return gin.H{
		"liveId":   r.LiveID,
		"courseId": r.CourseID,
		"agora": gin.H{
			"appId":         r.Agora.AppID,
			"channelName":   r.Agora.ChannelName,
			"role":          r.Agora.Role,
			"tokenExpireAt": r.Agora.TokenExpireAt,
		},
		"streams": gin.H{
			"camera": gin.H{"uid": r.Streams.Camera.UID, "rtcToken": r.Streams.Camera.RTCToken},
			"screen": gin.H{"uid": r.Streams.Screen.UID, "rtcToken": r.Streams.Screen.RTCToken},
		},
	}
}

func toParticipantsResponse(r livesuc.GetParticipantsResponse) gin.H {
	ps := make([]gin.H, 0, len(r.Participants))
	for _, p := range r.Participants {
		ps = append(ps, gin.H{
			"userId":      p.UserID,
			"displayName": p.DisplayName,
			"avatarUrl":   p.AvatarURL,
			"role":        p.Role,
			"agoraUid":    p.AgoraUID,
			"isMuted":     p.IsMuted,
			"isCameraOn":  p.IsCameraOn,
		})
	}
	return gin.H{
		"totalCount": r.TotalCount,
		"teacherStreams": gin.H{
			"teacherId": r.TeacherStreams.TeacherID,
			"cameraUid": r.TeacherStreams.CameraUID,
			"screenUid": r.TeacherStreams.ScreenUID,
		},
		"participants": ps,
	}
}

func toReminderResponse(r livesuc.ReminderResponse) gin.H {
	return gin.H{"liveId": r.LiveID, "isReminderEnabled": r.IsReminderEnabled}
}
