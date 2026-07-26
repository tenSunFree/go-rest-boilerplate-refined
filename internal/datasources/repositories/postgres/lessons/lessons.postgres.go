package postgres

import (
	"github.com/jmoiron/sqlx"
	repointerface "github.com/tenSunFree/luma-lang-go/internal/datasources/repositories/interface"
)

type postgreLessonRepository struct {
	conn *sqlx.DB
}

func NewLessonRepository(conn *sqlx.DB) repointerface.LessonRepository {
	return &postgreLessonRepository{conn: conn}
}
