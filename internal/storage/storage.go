package storage

import (
	"context"
	"github.com/hunterkritik-byte/FlakeHawk/internal/model"
)

type TestStore interface {
	SaveExecutions(context.Context, []model.TestExecution) error
	History(context.Context, string) ([]model.TestExecution, error)
}
