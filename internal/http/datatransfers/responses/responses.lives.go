package responses

import (
	"time"

	livesuc "github.com/tenSunFree/luma-lang-go/internal/business/usecases/lives"
)

// LiveCourseResponse is the item returned by GET /live-courses.
type LiveCourseResponse struct {
	CourseID          string          `json:"courseId"          example:"course_ielts_001"`
	LiveID            string          `json:"liveId"            example:"live_20260618_1900_001"`
	Title             string          `json:"title"             example:"雅思閱讀[一般訓練組] - 段落標題-複習篇"`
	Category          string          `json:"category"          example:"語言學習"`
	Level             string          `json:"level"             example:"A2"`
	CourseType        string          `json:"courseType"        example:"required"`
	Status            string          `json:"status"            example:"scheduled"`
	ScheduledStartAt  string          `json:"scheduledStartAt"  example:"2026-06-18T19:00:00+08:00"`
	StartedAt         *string         `json:"startedAt"         example:"2026-06-18T19:02:00+08:00"`
	EndedAt           *string         `json:"endedAt"`
	Teacher           TeacherResponse `json:"teacher"`
	ThumbnailURL      *string         `json:"thumbnailUrl"      example:"https://cdn.example.com/courses/ielts_001.jpg"`
	TextbookURL       *string         `json:"textbookUrl"       example:"https://cdn.example.com/textbooks/ielts_001.pdf"`
	ViewerCount       int             `json:"viewerCount"       example:"42"`
	IsReminderEnabled bool            `json:"isReminderEnabled" example:"false"`
	CanJoin           bool            `json:"canJoin"           example:"false"`
}

// TeacherResponse is the teacher block inside LiveCourseResponse.
type TeacherResponse struct {
	TeacherID string  `json:"teacherId" example:"teacher_ben"`
	Name      string  `json:"name"      example:"Ben"`
	AvatarURL *string `json:"avatarUrl" example:"https://cdn.example.com/teachers/ben.jpg"`
}

func FromLiveCourse(in livesuc.LiveCourse) LiveCourseResponse {
	return LiveCourseResponse{
		CourseID:          in.CourseID,
		LiveID:            in.LiveID,
		Title:             in.Title,
		Category:          in.Category,
		Level:             in.Level,
		CourseType:        in.CourseType,
		Status:            in.Status,
		ScheduledStartAt:  in.ScheduledStartAt.Format(time.RFC3339),
		StartedAt:         formatOptionalTime(in.StartedAt),
		EndedAt:           formatOptionalTime(in.EndedAt),
		Teacher:           TeacherResponse{TeacherID: in.TeacherID, Name: in.TeacherName, AvatarURL: in.TeacherAvatarURL},
		ThumbnailURL:      in.ThumbnailURL,
		TextbookURL:       in.TextbookURL,
		ViewerCount:       in.ViewerCount,
		IsReminderEnabled: in.IsReminderEnabled,
		CanJoin:           in.Status == livesuc.StatusLive,
	}
}

func FromLiveCourses(items []livesuc.LiveCourse) []LiveCourseResponse {
	out := make([]LiveCourseResponse, 0, len(items))
	for _, item := range items {
		out = append(out, FromLiveCourse(item))
	}
	return out
}

func formatOptionalTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	v := t.Format(time.RFC3339)
	return &v
}
