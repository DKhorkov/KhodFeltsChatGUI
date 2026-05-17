package notification_test

import (
	"context"
	"testing"

	notificationhandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/notification"
	"github.com/DKhorkov/libs/logging"
)

func TestHandler_SetContext(t *testing.T) {
	t.Parallel()

	h := notificationhandler.New(logging.New(logging.Levels.DEBUG, ""))
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	h := notificationhandler.New(logging.New(logging.Levels.DEBUG, ""))
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	h := notificationhandler.New(logging.New(logging.Levels.DEBUG, ""))
	h.StopListening()
}
