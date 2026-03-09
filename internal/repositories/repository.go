package repositories

import (
	"context"
	"io"

	"github.com/DKhorkov/libs/logging"
)

type Repository struct {
	logger logging.Logger
}

func NewRepository(logger logging.Logger) *Repository {
	return &Repository{
		logger: logger,
	}
}

func (r *Repository) closeBody(ctx context.Context, body io.ReadCloser) {
	if err := body.Close(); err != nil {
		logging.LogErrorContext(
			ctx,
			r.logger,
			"failed to close response body",
			err,
		)
	}
}
