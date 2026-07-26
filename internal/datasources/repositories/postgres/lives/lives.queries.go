package lives

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/tenSunFree/luma-lang-go/internal/apperror"
	livesuc "github.com/tenSunFree/luma-lang-go/internal/business/usecases/lives"
	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
)

func (r *postgreLiveRepository) ListLiveCourses(ctx context.Context, statuses []string, userID string) ([]livesuc.LiveCourse, error) {
	if len(statuses) == 0 {
		statuses = []string{livesuc.StatusScheduled, livesuc.StatusLive}
	}
	// Expand comma-separated strings (supports the `?status=scheduled,live` format)
	statuses = normalizeStatuses(statuses)

	var rows []records.LiveCourse
	err := r.conn.SelectContext(ctx, &rows, `
		SELECT
			lc.*,
			(
				SELECT COUNT(*)
				FROM live_participants lp
				WHERE lp.live_id = lc.id AND lp.left_at IS NULL
			)::int AS viewer_count,
			EXISTS (
				SELECT 1 FROM live_reminders lr
				WHERE lr.live_id = lc.id AND lr.user_id = $2
			) AS is_reminder_enabled
		FROM live_courses lc
		WHERE lc.status = ANY($1)
		ORDER BY lc.scheduled_start_at ASC
	`, pq.Array(statuses), userID)
	if err != nil {
		return nil, err
	}

	result := make([]livesuc.LiveCourse, 0, len(rows))
	for _, row := range rows {
		result = append(result, toUsecaseLiveCourse(row))
	}
	return result, nil
}

func (r *postgreLiveRepository) GetLiveByID(ctx context.Context, liveID string) (livesuc.LiveCourse, error) {
	var row records.LiveCourse
	err := r.conn.GetContext(ctx, &row, `
		SELECT
			lc.*,
			(
				SELECT COUNT(*)
				FROM live_participants lp
				WHERE lp.live_id = lc.id AND lp.left_at IS NULL
			)::int AS viewer_count,
			false AS is_reminder_enabled
		FROM live_courses lc
		WHERE lc.id = $1
	`, liveID)
	if err != nil {
		if err == sql.ErrNoRows {
			return livesuc.LiveCourse{}, apperror.NotFound("live not found")
		}
		return livesuc.LiveCourse{}, err
	}
	return toUsecaseLiveCourse(row), nil
}

func (r *postgreLiveRepository) GetNextStudentUID(ctx context.Context, liveID string) (int, error) {
	var nextUID int
	err := r.conn.GetContext(ctx, &nextUID, `
		SELECT GREATEST(
			$2,
			COALESCE(MAX(agora_uid), $2 - 1) + 1
		)
		FROM live_participants
		WHERE live_id = $1 AND agora_uid >= $2
	`, liveID, livesuc.StudentUIDBase)
	if err != nil {
		return 0, fmt.Errorf("get next student uid: %w", err)
	}
	return nextUID, nil
}

func (r *postgreLiveRepository) UpsertParticipant(ctx context.Context, in livesuc.UpsertParticipantInput) error {
	_, err := r.conn.ExecContext(ctx, `
		INSERT INTO live_participants (live_id, user_id, display_name, avatar_url, agora_uid, role, joined_at, left_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), NULL, now())
		ON CONFLICT (live_id, user_id, agora_uid)
		DO UPDATE SET
			display_name = EXCLUDED.display_name,
			avatar_url   = EXCLUDED.avatar_url,
			role         = EXCLUDED.role,
			left_at      = NULL,
			last_seen_at = now()
	`, in.LiveID, in.UserID, in.DisplayName, in.AvatarURL, in.AgoraUID, in.Role)
	return err
}

func (r *postgreLiveRepository) LeaveParticipant(ctx context.Context, liveID, userID string, uid int) error {
	_, err := r.conn.ExecContext(ctx, `
		UPDATE live_participants
		SET left_at = now(), last_seen_at = now()
		WHERE live_id = $1 AND user_id = $2 AND agora_uid = $3 AND left_at IS NULL
	`, liveID, userID, uid)
	return err
}

func (r *postgreLiveRepository) ListParticipants(ctx context.Context, liveID string) ([]livesuc.Participant, error) {
	var rows []records.LiveParticipant
	err := r.conn.SelectContext(ctx, &rows, `
		SELECT
			user_id,
			display_name,
			avatar_url,
			role,
			agora_uid,
			CASE WHEN role = 'teacher' AND agora_uid = 1000 THEN false ELSE true  END AS is_muted,
			CASE WHEN role = 'teacher' AND agora_uid = 1000 THEN true  ELSE false END AS is_camera_on
		FROM live_participants
		WHERE live_id = $1 AND left_at IS NULL
		ORDER BY
			CASE WHEN role = 'teacher' THEN 0 ELSE 1 END,
			joined_at ASC
	`, liveID)
	if err != nil {
		return nil, err
	}

	// Merge multiple agora uids for the same user (the teacher has two uids: camera and screen).
	merged := make(map[string]*livesuc.Participant)
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		key := row.UserID
		if _, ok := merged[key]; !ok {
			p := &livesuc.Participant{
				UserID:      row.UserID,
				DisplayName: row.DisplayName,
				AvatarURL:   row.AvatarURL,
				Role:        row.Role,
				AgoraUID:    row.AgoraUID,
				IsMuted:     row.IsMuted,
				IsCameraOn:  row.IsCameraOn,
			}
			merged[key] = p
			order = append(order, key)
		} else {
			// Teacher's second uid: merge camera/muted state
			if row.IsCameraOn {
				merged[key].IsCameraOn = true
			}
			if !row.IsMuted {
				merged[key].IsMuted = false
			}
		}
	}

	result := make([]livesuc.Participant, 0, len(order))
	for _, key := range order {
		result = append(result, *merged[key])
	}
	return result, nil
}

func (r *postgreLiveRepository) MarkLiveStarted(ctx context.Context, liveID string) error {
	_, err := r.conn.ExecContext(ctx, `
		UPDATE live_courses
		SET status = 'live', started_at = COALESCE(started_at, now()), updated_at = now()
		WHERE id = $1 AND status IN ('scheduled', 'live')
	`, liveID)
	return err
}

func (r *postgreLiveRepository) MarkLiveEnded(ctx context.Context, liveID string) error {
	_, err := r.conn.ExecContext(ctx, `
		UPDATE live_courses
		SET status = 'ended', ended_at = COALESCE(ended_at, now()), updated_at = now()
		WHERE id = $1
	`, liveID)
	return err
}

func (r *postgreLiveRepository) SetReminder(ctx context.Context, liveID, userID string, remindAt time.Time) error {
	_, err := r.conn.ExecContext(ctx, `
		INSERT INTO live_reminders (live_id, user_id, remind_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (live_id, user_id) DO UPDATE SET remind_at = EXCLUDED.remind_at
	`, liveID, userID, remindAt)
	return err
}

func (r *postgreLiveRepository) DeleteReminder(ctx context.Context, liveID, userID string) error {
	_, err := r.conn.ExecContext(ctx, `
		DELETE FROM live_reminders WHERE live_id = $1 AND user_id = $2
	`, liveID, userID)
	return err
}

// toUsecaseLiveCourse converts record to usecase domain.
func toUsecaseLiveCourse(row records.LiveCourse) livesuc.LiveCourse {
	return livesuc.LiveCourse{
		LiveID:            row.ID,
		CourseID:          row.CourseID,
		Title:             row.Title,
		Category:          row.Category,
		Level:             row.Level,
		CourseType:        row.CourseType,
		Status:            row.Status,
		ScheduledStartAt:  row.ScheduledStartAt,
		StartedAt:         row.StartedAt,
		EndedAt:           row.EndedAt,
		TeacherID:         row.TeacherID,
		TeacherName:       row.TeacherName,
		TeacherAvatarURL:  row.TeacherAvatarURL,
		ThumbnailURL:      row.ThumbnailURL,
		TextbookURL:       row.TextbookURL,
		AgoraChannelName:  row.AgoraChannelName,
		TeacherCameraUID:  row.TeacherCameraUID,
		TeacherScreenUID:  row.TeacherScreenUID,
		ViewerCount:       row.ViewerCount,
		IsReminderEnabled: row.IsReminderEnabled,
	}
}

// normalizeStatuses expand "scheduled,live" -> ["scheduled", "live"]
func normalizeStatuses(raw []string) []string {
	var out []string
	for _, s := range raw {
		for _, part := range strings.Split(s, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

func (r *postgreLiveRepository) CreateTeacherLive(ctx context.Context, in livesuc.CreateTeacherLiveInput) (livesuc.LiveCourse, error) {
	var row records.LiveCourse
	err := r.conn.GetContext(ctx, &row, `
		INSERT INTO live_courses (
			id, course_id, title, category, level, course_type,
			status, scheduled_start_at, started_at,
			teacher_id, teacher_name, teacher_avatar_url,
			thumbnail_url, textbook_url,
			agora_channel_name, teacher_camera_uid, teacher_screen_uid,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			'live', now(), now(),
			$7, $8, $9,
			$10, $11,
			$12, $13, $14,
			now(), now()
		)
		RETURNING
			id, course_id, title, category, level, course_type,
			status, scheduled_start_at, started_at, ended_at,
			teacher_id, teacher_name, teacher_avatar_url,
			thumbnail_url, textbook_url,
			agora_channel_name, teacher_camera_uid, teacher_screen_uid,
			created_at, updated_at,
			0::int AS viewer_count,
			false AS is_reminder_enabled
	`,
		in.LiveID, in.CourseID, in.Title, in.Category, in.Level, in.CourseType,
		in.TeacherID, in.TeacherName, in.AvatarURL,
		in.ThumbnailURL, in.TextbookURL,
		in.AgoraChannelName, in.TeacherCameraUID, in.TeacherScreenUID,
	)
	if err != nil {
		return livesuc.LiveCourse{}, fmt.Errorf("create teacher live: %w", err)
	}
	return toUsecaseLiveCourse(row), nil
}

func (r *postgreLiveRepository) GetActiveLiveByTeacherID(ctx context.Context, teacherID string) (livesuc.LiveCourse, error) {
	var row records.LiveCourse
	err := r.conn.GetContext(ctx, &row, `
		SELECT
			lc.*,
			(
				SELECT COUNT(*)
				FROM live_participants lp
				WHERE lp.live_id = lc.id AND lp.left_at IS NULL
			)::int AS viewer_count,
			false AS is_reminder_enabled
		FROM live_courses lc
		WHERE lc.teacher_id = $1
		  AND lc.status = 'live'
		ORDER BY lc.started_at DESC
		LIMIT 1
	`, teacherID)
	if err != nil {
		if err == sql.ErrNoRows {
			return livesuc.LiveCourse{}, apperror.NotFound("no active live found for teacher")
		}
		return livesuc.LiveCourse{}, err
	}
	return toUsecaseLiveCourse(row), nil
}

func (r *postgreLiveRepository) MarkTeacherActiveLiveEnded(ctx context.Context, teacherID string) (string, error) {
	var liveID string
	err := r.conn.GetContext(ctx, &liveID, `
		UPDATE live_courses
		SET status = 'ended',
		    ended_at = COALESCE(ended_at, now()),
		    updated_at = now()
		WHERE id = (
			SELECT id FROM live_courses
			WHERE teacher_id = $1 AND status = 'live'
			ORDER BY started_at DESC
			LIMIT 1
		)
		RETURNING id
	`, teacherID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", apperror.NotFound("no active live found for teacher")
		}
		return "", err
	}
	return liveID, nil
}
