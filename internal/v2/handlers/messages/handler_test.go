package messages_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	messageshandler "github.com/DKhorkov/kfcGUI/internal/v2/handlers/messages"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/DKhorkov/libs/pointers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_GetChatMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		chatID           uint64
		pagination       *domains.Pagination
		setupMocks       func(*mockusecases.MockUseCases)
		expectedMessages []domains.Message
		expectedError    error
	}{
		{
			name:   "successful get messages",
			chatID: 42,
			pagination: &domains.Pagination{
				Limit:  pointers.New[uint64](10),
				Offset: pointers.New[uint64](0),
			},
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetChatMessages(gomock.Any(), uint64(42), &domains.Pagination{Limit: pointers.New[uint64](10), Offset: pointers.New[uint64](0)}).
					Return([]domains.Message{{ID: 1, Text: "hello"}, {ID: 2, Text: "world"}}, nil).
					Times(1)
			},
			expectedMessages: []domains.Message{{ID: 1, Text: "hello"}, {ID: 2, Text: "world"}},
			expectedError:    nil,
		},
		{
			name:       "get messages without pagination",
			chatID:     42,
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetChatMessages(gomock.Any(), uint64(42), (*domains.Pagination)(nil)).
					Return([]domains.Message{}, nil).
					Times(1)
			},
			expectedMessages: []domains.Message{},
			expectedError:    nil,
		},
		{
			name:       "use case error",
			chatID:     42,
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetChatMessages(gomock.Any(), uint64(42), gomock.Any()).
					Return(nil, customerrors.ErrGetChatMessages).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    customerrors.ErrGetChatMessages,
		},
		{
			name:       "chat not found",
			chatID:     999,
			pagination: nil,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetChatMessages(gomock.Any(), uint64(999), gomock.Any()).
					Return(nil, customerrors.ErrChatNotFound).
					Times(1)
			},
			expectedMessages: nil,
			expectedError:    customerrors.ErrChatNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := messageshandler.New(mockUseCases)

			messages, err := h.GetChatMessages(tt.chatID, tt.pagination)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, messages)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMessages, messages)
			}
		})
	}
}

func TestHandler_SendMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		chatID           uint64
		text             string
		replyToMessageID *uint64
		setupMocks       func(*mockusecases.MockUseCases)
		expectedError    error
	}{
		{
			name:   "successful send message",
			chatID: 42,
			text:   "Hello!",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(&domains.User{ID: 1, Username: "john"}, nil).
					Times(1)
				uc.EXPECT().
					SendMessage(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:             "successful send reply message",
			chatID:           42,
			text:             "Reply!",
			replyToMessageID: pointers.New[uint64](10),
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(&domains.User{ID: 1, Username: "john"}, nil).
					Times(1)
				uc.EXPECT().
					SendMessage(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, msg domains.Message) error {
						assert.NotNil(t, msg.ReplyToMessage)
						assert.Equal(t, uint64(10), msg.ReplyToMessage.ID)

						return nil
					}).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:   "get current user error",
			chatID: 42,
			text:   "Hello!",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(nil, errors.New("session expired")).
					Times(1)
			},
			expectedError: errors.New("session expired"),
		},
		{
			name:   "send message use case error",
			chatID: 42,
			text:   "Hello!",
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetCurrentUser(gomock.Any()).
					Return(&domains.User{ID: 1}, nil).
					Times(1)
				uc.EXPECT().
					SendMessage(gomock.Any(), gomock.Any()).
					Return(errors.New("websocket closed")).
					Times(1)
			},
			expectedError: errors.New("websocket closed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := messageshandler.New(mockUseCases)

			err := h.SendMessage(tt.chatID, tt.text, tt.replyToMessageID)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_DeleteMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		messageID     uint64
		forAll        bool
		setupMocks    func(*mockusecases.MockUseCases)
		expectedError error
	}{
		{
			name:      "successful delete for self",
			messageID: 42,
			forAll:    false,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					DeleteMessage(gomock.Any(), domains.DeleteMessageDTO{
						MessageID: 42,
						ForAll:    false,
					}).
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:      "successful delete for all",
			messageID: 42,
			forAll:    true,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					DeleteMessage(gomock.Any(), domains.DeleteMessageDTO{
						MessageID: 42,
						ForAll:    true,
					}).
					Return(nil).
					Times(1)
			},
			expectedError: nil,
		},
		{
			name:      "use case error",
			messageID: 42,
			forAll:    false,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					DeleteMessage(gomock.Any(), domains.DeleteMessageDTO{
						MessageID: 42,
						ForAll:    false,
					}).
					Return(errors.New("delete failed")).
					Times(1)
			},
			expectedError: errors.New("delete failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := messageshandler.New(mockUseCases)

			err := h.DeleteMessage(tt.messageID, tt.forAll)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_GetMessageByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		messageID       uint64
		setupMocks      func(*mockusecases.MockUseCases)
		expectedMessage *domains.Message
		expectedError   error
	}{
		{
			name:      "successful get message",
			messageID: 42,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetMessageByID(gomock.Any(), uint64(42)).
					Return(&domains.Message{ID: 42, Text: "hello", ChatID: 1}, nil).
					Times(1)
			},
			expectedMessage: &domains.Message{ID: 42, Text: "hello", ChatID: 1},
			expectedError:   nil,
		},
		{
			name:      "message not found",
			messageID: 999,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetMessageByID(gomock.Any(), uint64(999)).
					Return(nil, customerrors.ErrMessageNotFound).
					Times(1)
			},
			expectedMessage: nil,
			expectedError:   customerrors.ErrMessageNotFound,
		},
		{
			name:      "use case error",
			messageID: 42,
			setupMocks: func(uc *mockusecases.MockUseCases) {
				uc.EXPECT().
					GetMessageByID(gomock.Any(), uint64(42)).
					Return(nil, errors.New("internal error")).
					Times(1)
			},
			expectedMessage: nil,
			expectedError:   errors.New("internal error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUseCases := mockusecases.NewMockUseCases(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUseCases)
			}

			h := messageshandler.New(mockUseCases)

			message, err := h.GetMessageByID(tt.messageID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, message)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMessage, message)
			}
		})
	}
}

func TestHandler_SetContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := messageshandler.New(mockUseCases)
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := messageshandler.New(mockUseCases)
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)

	h := messageshandler.New(mockUseCases)
	h.StopListening()
}
