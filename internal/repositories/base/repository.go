package base

import (
	"context"
	"io"

	"github.com/DKhorkov/libs/logging"
)

type Repository struct {
	logger logging.Logger
}

func New(logger logging.Logger) *Repository {
	return &Repository{
		logger: logger,
	}
}

func (r *Repository) CloseBody(ctx context.Context, body io.ReadCloser) {
	if body == nil {
		return
	}

	if err := body.Close(); err != nil {
		logging.LogErrorContext(
			ctx,
			r.logger,
			"failed to close response body",
			err,
		)
	}
}
