package create_chat

import (
	"context"
	"fmt"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

type Handler struct {
	useCases     interfaces.UseCases
	errorsMapper interfaces.ErrorsMapper

	wailsCtx context.Context
}

func New(
	useCases interfaces.UseCases,
	errorsMapper interfaces.ErrorsMapper,
) *Handler {
	return &Handler{
		useCases:     useCases,
		errorsMapper: errorsMapper,
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.wailsCtx = ctx
}

func (h *Handler) CreateChat(
	in domains.CreateChatDTO,
) (*domains.Chat, error) {
	ctx := context.Background()

	if !in.IsValid() {
		return nil, h.errorsMapper.Map(
			fmt.Errorf("%w: chat is not valid: %v+", customerrors.ErrInvalidChat, in),
		)
	}

	members := make([]domains.User, 0, len(in.MemberIDs))
	for _, id := range in.MemberIDs {
		members = append(members, domains.User{ID: id})
	}

	chat := &domains.Chat{
		Type:        in.Type,
		Members:     members,
		Title:       in.Title,
		Description: in.Description,
	}

	return h.useCases.CreateChat(ctx, *chat)
}

func (h *Handler) StartListening() {}

func (h *Handler) StopListening() {}
