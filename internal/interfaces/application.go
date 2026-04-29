package interfaces

import (
	"context"
)

//go:generate mockgen -source=application.go -destination=../../mocks/application/application.go -package=mockapplication -exclude_interfaces=
type Application interface {
	Startup(ctx context.Context)
	Shutdown(ctx context.Context)
	BindHandlers() []any
}
