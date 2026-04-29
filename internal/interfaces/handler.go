package interfaces

import "context"

//go:generate mockgen -source=handler.go -destination=../../mocks/handler/handler.go -package=mockhandler -exclude_interfaces=
type Handler interface {
	SetContext(ctx context.Context)
	StartListening()
	StopListening()
}
