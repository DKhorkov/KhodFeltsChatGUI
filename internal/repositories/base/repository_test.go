package base

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/DKhorkov/libs/logging/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// MockReadCloser реализует io.ReadCloser для тестирования.
type MockReadCloser struct {
	closeError error
	closed     bool
}

func (_ *MockReadCloser) Read(_ []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *MockReadCloser) Close() error {
	m.closed = true

	return m.closeError
}

func TestRepository_CloseBody(t *testing.T) {
	t.Parallel()

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
			ctx:            context.Background(),
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
			t.Parallel()

			// Setup
			ctrl := gomock.NewController(t)

			mockLogger := mocks.NewMockLogger(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mockLogger)
			}

			repo := New(mockLogger)

			// Execute
			repo.CloseBody(tt.ctx, tt.body)

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

func TestNew(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			repo := New(tt.logger)
			assert.NotNil(t, repo)
		})
	}
}

func TestRepository_closeBody_ErrorTypes(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

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

			repo := New(mockLogger)
			repo.CloseBody(context.Background(), body)
		})
	}
}
