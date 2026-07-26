package seeders

import (
	"github.com/tenSunFree/luma-lang-go/pkg/helpers"
)

const LiveTeacherEmail = "teacher_ben@gmail.com"

// LiveCourseSeeder ensures the teacher account exists.
// Live courses are now created by the teacher via POST /teacher/lives/start.
func (s *seeder) LiveCourseSeeder() error {
	teacherPass, err := helpers.GenerateHash("12345")
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO users (id, username, email, password, active, role_id, created_at)
		VALUES (uuid_generate_v4(), 'teacher_ben', $1, $2, true, 2, now())
		ON CONFLICT (email) WHERE deleted_at IS NULL
		DO UPDATE SET email = EXCLUDED.email
	`, LiveTeacherEmail, teacherPass)
	return err
}
