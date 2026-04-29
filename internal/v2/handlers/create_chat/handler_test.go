package create_chat_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	createchat "github.com/DKhorkov/kfcGUI/internal/v2/handlers/create_chat"
	mockerrors "github.com/DKhorkov/kfcGUI/mocks/errors"
	mockusecases "github.com/DKhorkov/kfcGUI/mocks/usecases"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_CreateChat(t *testing.T) {
	t.Parallel()

	title := "My Group"

	tests := []struct {
		name          string
		in            domains.CreateChatDTO
		setupMocks    func(*mockusecases.MockUseCases, *mockerrors.MockErrorsMapper)
		expectedChat  *domains.Chat
		expectedError error
	}{
		{
			name: "successful private chat creation",
			in: domains.CreateChatDTO{
				Type:      domains.ChatTypePrivate,
				MemberIDs: []uint64{2},
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					CreateChat(gomock.Any(), domains.Chat{
						Type:    domains.ChatTypePrivate,
						Members: []domains.User{{ID: 2}},
					}).
					Return(&domains.Chat{ID: 1, Type: domains.ChatTypePrivate}, nil).
					Times(1)
			},
			expectedChat:  &domains.Chat{ID: 1, Type: domains.ChatTypePrivate},
			expectedError: nil,
		},
		{
			name: "successful group chat creation",
			in: domains.CreateChatDTO{
				Type:      domains.ChatTypeGroup,
				MemberIDs: []uint64{2, 3},
				Title:     &title,
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					CreateChat(gomock.Any(), domains.Chat{
						Type:    domains.ChatTypeGroup,
						Members: []domains.User{{ID: 2}, {ID: 3}},
						Title:   &title,
					}).
					Return(&domains.Chat{ID: 2, Type: domains.ChatTypeGroup, Title: &title}, nil).
					Times(1)
			},
			expectedChat:  &domains.Chat{ID: 2, Type: domains.ChatTypeGroup, Title: &title},
			expectedError: nil,
		},
		{
			name: "invalid chat - unknown type",
			in: domains.CreateChatDTO{
				Type:      "unknown",
				MemberIDs: []uint64{2},
			},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().Map(gomock.Any()).Return(customerrors.ErrInvalidChat).Times(1)
			},
			expectedChat:  nil,
			expectedError: customerrors.ErrInvalidChat,
		},
		{
			name: "invalid private chat - no members",
			in: domains.CreateChatDTO{
				Type:      domains.ChatTypePrivate,
				MemberIDs: []uint64{},
			},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().Map(gomock.Any()).Return(customerrors.ErrInvalidChat).Times(1)
			},
			expectedChat:  nil,
			expectedError: customerrors.ErrInvalidChat,
		},
		{
			name: "invalid private chat - too many members",
			in: domains.CreateChatDTO{
				Type:      domains.ChatTypePrivate,
				MemberIDs: []uint64{2, 3},
			},
			setupMocks: func(_ *mockusecases.MockUseCases, em *mockerrors.MockErrorsMapper) {
				em.EXPECT().Map(gomock.Any()).Return(customerrors.ErrInvalidChat).Times(1)
			},
			expectedChat:  nil,
			expectedError: customerrors.ErrInvalidChat,
		},
		{
			name: "chat already exists",
			in: domains.CreateChatDTO{
				Type:      domains.ChatTypePrivate,
				MemberIDs: []uint64{2},
			},
			setupMocks: func(uc *mockusecases.MockUseCases, _ *mockerrors.MockErrorsMapper) {
				uc.EXPECT().
					CreateChat(gomock.Any(), gomock.Any()).
					Return(nil, customerrors.ErrChatAlreadyExists).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: customerrors.ErrChatAlreadyExists,
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
					Return(nil, errors.New("database unavailable")).
					Times(1)
			},
			expectedChat:  nil,
			expectedError: errors.New("database unavailable"),
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

			h := createchat.New(mockUseCases, mockMapper)

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

	h := createchat.New(mockUseCases, mockMapper)
	h.SetContext(context.Background())
}

func TestHandler_StartListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := createchat.New(mockUseCases, mockMapper)
	h.StartListening()
}

func TestHandler_StopListening(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	mockUseCases := mockusecases.NewMockUseCases(ctrl)
	mockMapper := mockerrors.NewMockErrorsMapper(ctrl)

	h := createchat.New(mockUseCases, mockMapper)
	h.StopListening()
}
