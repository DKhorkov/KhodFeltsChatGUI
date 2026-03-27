package repositories

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/DKhorkov/libs/logging/mocks"
	"go.uber.org/mock/gomock"
)

// MockReadCloser реализует io.ReadCloser для тестирования.
type MockReadCloser struct {
	closeError error
	closed     bool
}

func (m *MockReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *MockReadCloser) Close() error {
	m.closed = true

	return m.closeError
}

func TestRepository_closeBody(t *testing.T) {
	tests := []struct {
		name           string
		body           io.ReadCloser
		ctx            context.Context
		setupMocks     func(*mocks.MockLogger)
		shouldLogError bool
		expectedMsg    string
	}{
		{
			name: "successful close without error",
			body: &MockReadCloser{
				closeError: nil,
			},
			ctx: context.Background(),
			setupMocks: func(mockLogger *mocks.MockLogger) {
				// Не ожидаем вызовов логгера
			},
			shouldLogError: false,
		},
		{
			name: "close with error - should log error",
			body: &MockReadCloser{
				closeError: errors.New("connection closed unexpectedly"),
			},
			ctx: context.Background(),
			setupMocks: func(mockLogger *mocks.MockLogger) {
				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to close response body",
						gomock.Any(),
					).
					Times(1)
			},
			shouldLogError: true,
			expectedMsg:    "failed to close response body",
		},
		{
			name: "close with context containing values",
			body: &MockReadCloser{
				closeError: errors.New("timeout error"),
			},
			ctx: context.WithValue(context.Background(), "request_id", "12345"),
			setupMocks: func(mockLogger *mocks.MockLogger) {
				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to close response body",
						gomock.Any(),
					).
					Times(1)
			},
			shouldLogError: true,
			expectedMsg:    "failed to close response body",
		},
		{
			name: "close with nil body",
			body: nil,
			ctx:  context.Background(),
		},
		{
			name: "close with body that returns specific error type",
			body: &MockReadCloser{
				closeError: io.ErrClosedPipe,
			},
			ctx: context.Background(),
			setupMocks: func(mockLogger *mocks.MockLogger) {
				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to close response body",
						gomock.Any(),
					).
					Times(1)
			},
			shouldLogError: true,
			expectedMsg:    "failed to close response body",
		},
		{
			name: "close with canceled context",
			body: &MockReadCloser{
				closeError: errors.New("context canceled"),
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			setupMocks: func(mockLogger *mocks.MockLogger) {
				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to close response body",
						gomock.Any(),
					).
					Times(1)
			},
			shouldLogError: true,
			expectedMsg:    "failed to close response body",
		},
		{
			name: "close with multiple arguments in error log",
			body: &MockReadCloser{
				closeError: errors.New("close failed"),
			},
			ctx: context.Background(),
			setupMocks: func(mockLogger *mocks.MockLogger) {
				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to close response body",
						gomock.Any(), // args могут включать дополнительные параметры
					).
					Times(1)
			},
			shouldLogError: true,
			expectedMsg:    "failed to close response body",
		},
		{
			name: "close with error when body already closed",
			body: &MockReadCloser{
				closeError: errors.New("file already closed"),
			},
			ctx: context.Background(),
			setupMocks: func(mockLogger *mocks.MockLogger) {
				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to close response body",
						gomock.Any(),
					).
					Times(1)
			},
			shouldLogError: true,
			expectedMsg:    "failed to close response body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)

			mockLogger := mocks.NewMockLogger(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockLogger)
			}

			repo := NewRepository(mockLogger)

			// Execute
			repo.closeBody(tt.ctx, tt.body)

			// Дополнительные проверки для body
			if tt.body != nil {
				mockBody, ok := tt.body.(*MockReadCloser)
				if ok {
					if !tt.shouldLogError {
						if !mockBody.closed {
							t.Errorf("Expected body to be closed, but it wasn't")
						}
					} else {
						// Даже при ошибке, Close все равно вызывается
						if !mockBody.closed {
							t.Errorf("Body Close should be called even on error")
						}
					}
				}
			}
		})
	}
}

func TestNewRepository(t *testing.T) {
	tests := []struct {
		name   string
		logger *mocks.MockLogger
	}{
		{
			name:   "create repository with valid logger",
			logger: mocks.NewMockLogger(gomock.NewController(t)),
		},
		{
			name:   "create repository with nil logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewRepository(tt.logger)
			if repo == nil {
				t.Error("Expected non-nil repository")
			}

			if repo.logger != tt.logger {
				t.Errorf("Expected logger = %v, got %v", tt.logger, repo.logger)
			}
		})
	}
}

func TestRepository_closeBody_ErrorTypes(t *testing.T) {
	tests := []struct {
		name       string
		closeError error
		shouldLog  bool
	}{
		{
			name:       "io.EOF error",
			closeError: io.EOF,
			shouldLog:  true,
		},
		{
			name:       "io.ErrClosedPipe error",
			closeError: io.ErrClosedPipe,
			shouldLog:  true,
		},
		{
			name:       "io.ErrUnexpectedEOF error",
			closeError: io.ErrUnexpectedEOF,
			shouldLog:  true,
		},
		{
			name:       "context.Canceled error",
			closeError: context.Canceled,
			shouldLog:  true,
		},
		{
			name:       "context.DeadlineExceeded error",
			closeError: context.DeadlineExceeded,
			shouldLog:  true,
		},
		{
			name:       "custom error",
			closeError: errors.New("custom close error"),
			shouldLog:  true,
		},
		{
			name:       "nil error",
			closeError: nil,
			shouldLog:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockLogger := mocks.NewMockLogger(ctrl)

			body := &MockReadCloser{
				closeError: tt.closeError,
			}

			if tt.shouldLog {
				mockLogger.EXPECT().
					ErrorContext(
						gomock.Any(),
						"failed to close response body",
						gomock.Any(),
					).
					Times(1)
			}

			repo := NewRepository(mockLogger)
			repo.closeBody(context.Background(), body)
		})
	}
}
