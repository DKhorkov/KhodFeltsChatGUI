package application

import (
	"context"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	mockhandler "github.com/DKhorkov/kfcGUI/mocks/handler"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	loggingmocks "github.com/DKhorkov/libs/logging/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestApp_Startup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handlerCount int
		setupMocks   func(*loggingmocks.MockLogger, []*mockhandler.MockHandler)
	}{
		{
			name:         "startup with no handlers",
			handlerCount: 0,
			setupMocks: func(logger *loggingmocks.MockLogger, handlers []*mockhandler.MockHandler) {
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
			},
		},
		{
			name:         "startup with single handler calls SetContext",
			handlerCount: 1,
			setupMocks: func(logger *loggingmocks.MockLogger, handlers []*mockhandler.MockHandler) {
				handlers[0].EXPECT().SetContext(gomock.Any()).Times(1)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
			},
		},
		{
			name:         "startup with multiple handlers calls SetContext on each",
			handlerCount: 3,
			setupMocks: func(logger *loggingmocks.MockLogger, handlers []*mockhandler.MockHandler) {
				for _, h := range handlers {
					h.EXPECT().SetContext(gomock.Any()).Times(1)
				}
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockLogger := loggingmocks.NewMockLogger(ctrl)

			mockHandlers := make([]*mockhandler.MockHandler, tt.handlerCount)
			for i := range mockHandlers {
				mockHandlers[i] = mockhandler.NewMockHandler(ctrl)
			}

			tt.setupMocks(mockLogger, mockHandlers)

			handlers := make([]interfaces.Handler, len(mockHandlers))
			for i, h := range mockHandlers {
				handlers[i] = h
			}

			app := New(mockUseCases, mockLogger, nil, handlers)
			app.Startup(context.Background())
		})
	}
}

func TestApp_Shutdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handlerCount int
		setupMocks   func(*loggingmocks.MockLogger, []*mockhandler.MockHandler)
	}{
		{
			name:         "shutdown with no handlers",
			handlerCount: 0,
			setupMocks: func(logger *loggingmocks.MockLogger, handlers []*mockhandler.MockHandler) {
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
			},
		},
		{
			name:         "shutdown with single handler calls StopListening",
			handlerCount: 1,
			setupMocks: func(logger *loggingmocks.MockLogger, handlers []*mockhandler.MockHandler) {
				handlers[0].EXPECT().StopListening().Times(1)
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
			},
		},
		{
			name:         "shutdown with multiple handlers calls StopListening on each",
			handlerCount: 3,
			setupMocks: func(logger *loggingmocks.MockLogger, handlers []*mockhandler.MockHandler) {
				for _, h := range handlers {
					h.EXPECT().StopListening().Times(1)
				}
				logger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockLogger := loggingmocks.NewMockLogger(ctrl)

			mockHandlers := make([]*mockhandler.MockHandler, tt.handlerCount)
			for i := range mockHandlers {
				mockHandlers[i] = mockhandler.NewMockHandler(ctrl)
			}

			tt.setupMocks(mockLogger, mockHandlers)

			handlers := make([]interfaces.Handler, len(mockHandlers))
			for i, h := range mockHandlers {
				handlers[i] = h
			}

			app := New(mockUseCases, mockLogger, nil, handlers)
			app.Shutdown(context.Background())
		})
	}
}

func TestApp_BindHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		handlerCount   int
		expectedLength int
	}{
		{
			name:           "no handlers returns empty slice",
			handlerCount:   0,
			expectedLength: 0,
		},
		{
			name:           "single handler",
			handlerCount:   1,
			expectedLength: 1,
		},
		{
			name:           "multiple handlers",
			handlerCount:   3,
			expectedLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			mockHandlers := make([]*mockhandler.MockHandler, tt.handlerCount)
			for i := range mockHandlers {
				mockHandlers[i] = mockhandler.NewMockHandler(ctrl)
			}

			handlers := make([]interfaces.Handler, len(mockHandlers))
			for i, h := range mockHandlers {
				handlers[i] = h
			}

			app := New(mockUseCases, nil, nil, handlers)

			result := app.BindHandlers()

			assert.Len(t, result, tt.expectedLength)
			for i, h := range result {
				assert.Equal(t, handlers[i], h)
			}
		})
	}
}
