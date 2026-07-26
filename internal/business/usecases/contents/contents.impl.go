package contents

import (
	repointerface "github.com/tenSunFree/luma-lang-go/internal/datasources/repositories/interface"
)

type usecase struct {
	repo repointerface.ContentRepository
}

func NewUsecase(repo repointerface.ContentRepository) Usecase {
	return &usecase{repo: repo}
}
