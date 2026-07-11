package chats_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	chathandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/chats"
	mockerrors "github.com/DKhorkov/kfcGUI/mocks/errors"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	loggingmocks "github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/pointers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_GetUserChats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		pagination    *domains.Pagination
		setupMocks    func(*mockusecases.MockUseCases)
		expectedChats []domains.Chat
		expectedError error
	}{
		{
			name:       "successful get chats without pagination",
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetUserChats(gomock.Any(), (*domains.Pagination)(nil)).
					Return([]domains.Chat{{ID: 1}, {ID: 2}}, nil).
					Times(1)
			},
			expectedChats: []domains.Chat{{ID: 1}, {ID: 2}},
			expectedError: nil,
		},
		{
			name: "successful get chats with pagination",
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](0),
			},
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetUserChats(gomock.Any(), &domains.Pagination{Limit: pointers.New[uint64](10), Offset: pointers.New[uint64](0)}).
					Return([]domains.Chat{{ID: 1}}, nil).
					Times(1)
			},
			expectedChats: []domains.Chat{{ID: 1}},
			expectedError: nil,
		},
		{
			name:       "empty chats list",
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetUserChats(gomock.Any(), gomock.Any()).
					Return([]domains.Chat{}, nil).
					Times(1)
			},
			expectedChats: []domains.Chat{},
			expectedError: nil,
		},
		{
			name:       "use case error",
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetUserChats(gomock.Any(), gomock.Any()).
					Return(nil, customerrors.ErrGetUserChats).
					Times(1)
			},
			expectedChats: nil,
			expectedError: customerrors.ErrGetUserChats,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := chathandler.New(mockUseCases, mockMapper, nil)

			chats, err := h.GetUserChats(tt.pagination)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, chats)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedChats, chats)
			}
		})
	}
}

func TestHandler_CreateChat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		in            domains.CreateChatDTO
		setupMocks    func(*mockusecases.MockUseCases, *mockerrors.MockErrorsMapper)
		expectedChat  *domains.Chat
		expectedError error
	}{
		{
			name: "successful create private chat",
			in: domains.CreateChatDTO{
				Type:      domains.ChatTypePrivate,
				MemberIDs: []uint64{2},
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					CreateChat(gomock.Any(), gomock.Any()).
					Return(&domains.Chat{ID: 1, Type: domains.ChatTypePrivate}, nil).
					Times(1)
			},
			expectedChat:  &domains.Chat{ID: 1, Type: domains.ChatTypePrivate},
			expectedError: nil,
		},
		{
			name: "successful create group chat",
			in: domains.CreateChatDTO{
				Type:        domains.ChatTypeGroup,
				MemberIDs:   []uint64{2, 3},
				Title:       pointers.New("Test Group"),
				Description: pointers.New("Description"),
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					CreateChat(gomock.Any(), gomock.Any()).
					Return(&domains.Chat{ID: 2, Type: domains.ChatTypeGroup}, nil).
					Times(1)
			},
			expectedChat:  &domains.Chat{ID: 2, Type: domains.ChatTypeGroup},
			expectedError: nil,
		},
		{
			name: "invalid chat - private with no members",
			in: domains.CreateChatDTO{
				Type:      domains.ChatTypePrivate,
				MemberIDs: []uint64{},
			},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().
					Map(gomock.Any()).
					Return(customerrors.ErrInvalidChat).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: customerrors.ErrInvalidChat,
		},
		{
			name: "use case error",
			in: domains.CreateChatDTO{
				Type:      domains.ChatTypePrivate,
				MemberIDs: []uint64{2},
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					CreateChat(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("create failed")).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: errors.New("create failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases, mockMapper)
			}

			h := chathandler.New(mockUseCases, mockMapper, nil)

			chat, err := h.CreateChat(tt.in)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, chat)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedChat, chat)
			}
		})
	}
}

func TestHandler_SetContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := chathandler.New(mockUseCases, mockMapper, nil)
	h.SetContext(context.Background())
}

func TestHandler_StartListening_StopListening(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(*mockusecases.MockUseCases, *loggingmocks.MockLogger)
		skipStart  bool
	}{
		{
			name: "start and stop with ReadEvent returning ErrWebsocketClosed",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				// readMessages goroutine: exits on ErrWebsocketClosed
				uc.EXPECT().
					ReadEvent(gomock.Any()).
					Return((*domains.WSEvent)(nil), customerrors.ErrWebsocketClosed).
					AnyTimes()
				// logging.LogInfo is called inside readMessages on ErrWebsocketClosed
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				// logging.LogErrorContext may also be called
				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name:       "stop without start does not panic",
			setupMocks: func(_ *mockusecases.MockUseCases, _ *loggingmocks.MockLogger) {},
			skipStart:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)
			mockLogger := loggingmocks.NewMockLogger(ctrl)

			tt.setupMocks(mockUseCases, mockLogger)

			h := chathandler.New(mockUseCases, mockMapper, mockLogger)

			if !tt.skipStart {
				h.StartListening()
			}

			h.StopListening()
		})
	}
}

func TestHandler_ReadEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(*mockusecases.MockUseCases, *loggingmocks.MockLogger)
	}{
		{
			name: "ReadEvent returns generic error then exits on ErrWebsocketClosed",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							return nil, errors.New("connection timeout")
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "ReadEvent returns NewMessage with invalid JSON then exits",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							return &domains.WSEvent{
								Type:    domains.WSEventNewMessage,
								Payload: json.RawMessage(`{invalid json`),
							}, nil
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "ReadEvent returns MessageDeleted with invalid JSON then exits",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							return &domains.WSEvent{
								Type:    domains.WSEventMessageDeleted,
								Payload: json.RawMessage(`not a json`),
							}, nil
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "ReadEvent returns MessageEdited with valid JSON then exits",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							payload, _ := json.Marshal(domains.MessageEditedPayload{
								MessageID: 42,
								ChatID:    1,
							})

							return &domains.WSEvent{
								Type:    domains.WSEventMessageEdited,
								Payload: payload,
							}, nil
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "ReadEvent returns MessageEdited with invalid JSON then exits",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							return &domains.WSEvent{
								Type:    domains.WSEventMessageEdited,
								Payload: json.RawMessage(`not a json`),
							}, nil
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "ReadEvent returns ReactionAdded with valid JSON then exits",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							payload, _ := json.Marshal(domains.ReactionAddedPayload{
								MessageID:  10,
								ChatID:     1,
								UserID:     7,
								ReactionID: 2,
								Emoji:      "👍",
							})

							return &domains.WSEvent{
								Type:    domains.WSEventReactionAdded,
								Payload: payload,
							}, nil
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "ReadEvent returns ReactionAdded with invalid JSON then exits",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							return &domains.WSEvent{
								Type:    domains.WSEventReactionAdded,
								Payload: json.RawMessage(`not a json`),
							}, nil
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "ReadEvent returns ReactionRemoved with valid JSON then exits",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							payload, _ := json.Marshal(domains.ReactionRemovedPayload{
								MessageID:  10,
								ChatID:     1,
								UserID:     7,
								ReactionID: 2,
							})

							return &domains.WSEvent{
								Type:    domains.WSEventReactionRemoved,
								Payload: payload,
							}, nil
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "ReadEvent returns ReactionRemoved with invalid JSON then exits",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							return &domains.WSEvent{
								Type:    domains.WSEventReactionRemoved,
								Payload: json.RawMessage(`not a json`),
							}, nil
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "ReadEvent returns unknown event type then exits",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				var callCount atomic.Int32

				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(_ context.Context) (*domains.WSEvent, error) {
						if callCount.Add(1) == 1 {
							return &domains.WSEvent{
								Type:    domains.WSEventType("unknown_event"),
								Payload: json.RawMessage(`{}`),
							}, nil
						}

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
		{
			name: "context cancellation stops readEvents goroutine",
			setupMocks: func(uc *mockusecases.MockUseCases, logger *loggingmocks.MockLogger) {
				uc.EXPECT().
					ReadEvent(gomock.Any()).
					DoAndReturn(func(ctx context.Context) (*domains.WSEvent, error) {
						<-ctx.Done()

						return nil, customerrors.ErrWebsocketClosed
					}).
					AnyTimes()

				logger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				logger.EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)
			mockMapper := mockerrors.NewMockErrorsMapper(ctrl)
			mockLogger := loggingmocks.NewMockLogger(ctrl)

			tt.setupMocks(mockUseCases, mockLogger)

			h := chathandler.New(mockUseCases, mockMapper, mockLogger)
			h.StartListening()
			h.StopListening()
		})
	}
}
