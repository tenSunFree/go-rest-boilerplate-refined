package lives

import (
	"context"
	"fmt"
	"time"

	"github.com/tenSunFree/luma-lang-go/internal/apperror"
)

type usecase struct {
	repo         Repository
	tokenService TokenService
	appID        string
	wsBaseURL    string
}

func NewUsecase(repo Repository, tokenService TokenService, appID, wsBaseURL string) Usecase {
	return &usecase{
		repo:         repo,
		tokenService: tokenService,
		appID:        appID,
		wsBaseURL:    wsBaseURL,
	}
}

func (uc *usecase) ListLiveCourses(ctx context.Context, req ListLiveCoursesRequest) (ListLiveCoursesResponse, error) {
	items, err := uc.repo.ListLiveCourses(ctx, req.Statuses, req.UserID)
	if err != nil {
		return ListLiveCoursesResponse{}, err
	}
	return ListLiveCoursesResponse{Items: items}, nil
}

func (uc *usecase) JoinLive(ctx context.Context, req JoinLiveRequest) (JoinLiveResponse, error) {
	live, err := uc.repo.GetLiveByID(ctx, req.LiveID)
	if err != nil {
		return JoinLiveResponse{}, err
	}
	if live.Status != StatusLive {
		return JoinLiveResponse{}, apperror.Forbidden("live is not joinable")
	}

	uid, err := uc.repo.GetNextStudentUID(ctx, req.LiveID)
	if err != nil {
		return JoinLiveResponse{}, apperror.InternalCause(fmt.Errorf("allocate uid: %w", err))
	}

	token, err := uc.tokenService.BuildRTCToken(live.AgoraChannelName, uid, RoleAudience)
	if err != nil {
		return JoinLiveResponse{}, apperror.InternalCause(fmt.Errorf("build student rtc token: %w", err))
	}

	if err := uc.repo.UpsertParticipant(ctx, UpsertParticipantInput{
		LiveID:      req.LiveID,
		UserID:      req.UserID,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		AgoraUID:    uid,
		Role:        RoleAudience,
	}); err != nil {
		return JoinLiveResponse{}, err
	}

	return JoinLiveResponse{
		LiveID:   live.LiveID,
		CourseID: live.CourseID,
		Agora: AgoraJoinConfig{
			AppID:         uc.appID,
			ChannelName:   live.AgoraChannelName,
			UID:           uid,
			RTCToken:      token.Token,
			Role:          RoleAudience,
			TokenExpireAt: token.ExpireAt,
		},
		Teacher: TeacherConfig{
			TeacherID: live.TeacherID,
			Name:      live.TeacherName,
			AvatarURL: live.TeacherAvatarURL,
			CameraUID: live.TeacherCameraUID,
			ScreenUID: live.TeacherScreenUID,
		},
		StreamLayout: StreamLayout{
			MainUID:  live.TeacherScreenUID,
			MainType: "teacher_screen",
			PipUID:   live.TeacherCameraUID,
			PipType:  "teacher_camera",
		},
		Chat: ChatConfig{
			Enabled:  true,
			Provider: "websocket",
			RoomID:   "chat_" + live.LiveID,
			WSUrl:    fmt.Sprintf("%s/ws/lives/%s", uc.wsBaseURL, live.LiveID),
		},
		Features: LiveFeatures{
			CanSendMessage:  true,
			CanRaiseHand:    true,
			CanSendReaction: true,
			CanPublishAudio: false,
			CanPublishVideo: false,
		},
	}, nil
}

func (uc *usecase) LeaveLive(ctx context.Context, req LeaveLiveRequest) error {
	return uc.repo.LeaveParticipant(ctx, req.LiveID, req.UserID, req.UID)
}

func (uc *usecase) RenewToken(ctx context.Context, req RenewTokenRequest) (RenewTokenResponse, error) {
	live, err := uc.repo.GetLiveByID(ctx, req.LiveID)
	if err != nil {
		return RenewTokenResponse{}, err
	}
	role := req.Role
	if role == "" {
		role = RoleAudience
	}
	token, err := uc.tokenService.BuildRTCToken(live.AgoraChannelName, req.UID, role)
	if err != nil {
		return RenewTokenResponse{}, apperror.InternalCause(fmt.Errorf("renew rtc token: %w", err))
	}
	return RenewTokenResponse{
		UID:           req.UID,
		RTCToken:      token.Token,
		TokenExpireAt: token.ExpireAt,
	}, nil
}

func (uc *usecase) GetParticipants(ctx context.Context, req GetParticipantsRequest) (GetParticipantsResponse, error) {
	live, err := uc.repo.GetLiveByID(ctx, req.LiveID)
	if err != nil {
		return GetParticipantsResponse{}, err
	}
	items, err := uc.repo.ListParticipants(ctx, req.LiveID)
	if err != nil {
		return GetParticipantsResponse{}, err
	}
	return GetParticipantsResponse{
		TotalCount: len(items),
		TeacherStreams: TeacherStreams{
			TeacherID: live.TeacherID,
			CameraUID: live.TeacherCameraUID,
			ScreenUID: live.TeacherScreenUID,
		},
		Participants: items,
	}, nil
}

// StartTeacherLive: Teachers directly create and start a live stream room.
// If the teacher already has a live stream in progress, they directly send back a new token for that session (without creating a duplicate token).
func (uc *usecase) StartTeacherLive(ctx context.Context, req StartTeacherLiveRequest) (StartTeacherLiveResponse, error) {
	// First check if the teacher already has a live stream in progress.
	live, err := uc.repo.GetActiveLiveByTeacherID(ctx, req.TeacherID)
	if err != nil {
		// No live stream in progress → Create a new one
		if !apperror.IsNotFound(err) {
			return StartTeacherLiveResponse{}, err
		}

		now := time.Now()
		suffix := fmt.Sprintf("%d%d%d_%d%d%d_%d",
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second(),
			now.UnixNano()%1000000,
		)
		liveID := "live_" + suffix
		courseID := "course_live_" + suffix
		channelName := "ch_" + suffix

		courseType := req.CourseType
		if courseType == "" {
			courseType = "live"
		}
		category := req.Category
		if category == "" {
			category = "直播課"
		}

		live, err = uc.repo.CreateTeacherLive(ctx, CreateTeacherLiveInput{
			LiveID:           liveID,
			CourseID:         courseID,
			Title:            req.Title,
			Category:         category,
			Level:            req.Level,
			CourseType:       courseType,
			TeacherID:        req.TeacherID,
			TeacherName:      req.TeacherName,
			AvatarURL:        req.AvatarURL,
			ThumbnailURL:     req.ThumbnailURL,
			TextbookURL:      req.TextbookURL,
			AgoraChannelName: channelName,
			TeacherCameraUID: TeacherCameraUID,
			TeacherScreenUID: TeacherScreenUID,
		})
		if err != nil {
			return StartTeacherLiveResponse{}, err
		}
	}
	// If a live stream is already in progress, generate a new token for that stream and send it back directly (idempotent design).
	// Generate a universal token using uid=0; both camera (uid=1000) and screen (uid=2000) can use it.
	broadcasterToken, err := uc.tokenService.BuildRTCToken(live.AgoraChannelName, 0, RoleBroadcaster)
	if err != nil {
		return StartTeacherLiveResponse{}, apperror.InternalCause(fmt.Errorf("build broadcaster token: %w", err))
	}

	_ = uc.repo.UpsertParticipant(ctx, UpsertParticipantInput{
		LiveID:      live.LiveID,
		UserID:      live.TeacherID,
		DisplayName: live.TeacherName,
		AvatarURL:   live.TeacherAvatarURL,
		AgoraUID:    live.TeacherCameraUID,
		Role:        RoleTeacher,
	})

	return StartTeacherLiveResponse{
		LiveID:   live.LiveID,
		CourseID: live.CourseID,
		Agora: TeacherAgoraConfig{
			AppID:         uc.appID,
			ChannelName:   live.AgoraChannelName,
			Role:          RoleBroadcaster,
			TokenExpireAt: broadcasterToken.ExpireAt,
		},
		// For the same token, each uid corresponds to a different one
		Streams: TeacherStreamsToken{
			Camera: TeacherStreamToken{UID: live.TeacherCameraUID, RTCToken: broadcasterToken.Token},
			Screen: TeacherStreamToken{UID: live.TeacherScreenUID, RTCToken: broadcasterToken.Token},
		},
	}, nil
}

// EndTeacherLive Ends the teacher's current live stream and returns the ended liveId.
func (uc *usecase) EndTeacherLive(ctx context.Context, req EndTeacherLiveRequest) (string, error) {
	return uc.repo.MarkTeacherActiveLiveEnded(ctx, req.TeacherID)
}

func (uc *usecase) SetReminder(ctx context.Context, req SetReminderRequest) (ReminderResponse, error) {
	live, err := uc.repo.GetLiveByID(ctx, req.LiveID)
	if err != nil {
		return ReminderResponse{}, err
	}
	remindAt := live.ScheduledStartAt.Add(-10 * time.Minute)
	if err := uc.repo.SetReminder(ctx, req.LiveID, req.UserID, remindAt); err != nil {
		return ReminderResponse{}, err
	}
	return ReminderResponse{LiveID: req.LiveID, IsReminderEnabled: true}, nil
}

func (uc *usecase) DeleteReminder(ctx context.Context, req DeleteReminderRequest) (ReminderResponse, error) {
	if err := uc.repo.DeleteReminder(ctx, req.LiveID, req.UserID); err != nil {
		return ReminderResponse{}, err
	}
	return ReminderResponse{LiveID: req.LiveID, IsReminderEnabled: false}, nil
}
