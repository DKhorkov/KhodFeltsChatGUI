package chats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
	"github.com/DKhorkov/libs/logging"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	refreshTokensInterval = 1 * time.Minute
	updateChatsInterval   = 5 * time.Second

	chatsUpdatedEventName    = "chats_updated"
	newMessageEventName      = "new_message"
	messageDeletedEventName  = "message_deleted"
	messageEditedEventName   = "message_edited"
	reactionAddedEventName   = "reaction_added"
	reactionRemovedEventName = "reaction_removed"
)

type Handler struct {
	useCases     interfaces.UseCases
	errorsMapper interfaces.ErrorsMapper
	logger       logging.Logger

	wailsCtx                context.Context
	goroutinesCtx           context.Context
	goroutinesCtxCancelFunc context.CancelFunc
	wg                      sync.WaitGroup
}

func New(
	useCases interfaces.UseCases,
	errorsMapper interfaces.ErrorsMapper,
	logger logging.Logger,
) *Handler {
	return &Handler{
		useCases:     useCases,
		errorsMapper: errorsMapper,
		logger:       logger,
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.wailsCtx = ctx
}

func (h *Handler) GetUserChats(pagination *domains.Pagination) ([]domains.Chat, error) {
	ctx := context.Background()

	return h.useCases.GetUserChats(ctx, pagination)
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

func (h *Handler) StartListening() {
	h.goroutinesCtx, h.goroutinesCtxCancelFunc = context.WithCancel(context.Background())

	// TODO Wails CLI v2.12.0 не умеет в wg.Go(), поэтому надо будет обновить в дальнейшем

	// Запускаем горутину для чтения сообщений
	h.wg.Add(1)

	go h.readEvents()

	// Запускаем обновление токенов
	h.wg.Add(1)

	go h.refreshTokens()

	// Запускаем обновление чатов
	h.wg.Add(1)

	go h.updateChats()
}

func (h *Handler) StopListening() {
	if h.goroutinesCtxCancelFunc != nil {
		h.goroutinesCtxCancelFunc()
	}

	h.wg.Wait()
}

func (h *Handler) readEvents() {
	defer h.wg.Done()

	for {
		select {
		case <-h.goroutinesCtx.Done():
			return
		default:
			event, err := h.useCases.ReadEvent(h.goroutinesCtx)
			if err != nil {
				// Соккет закрыт, отключаем горутину
				if errors.Is(err, customerrors.ErrWebsocketClosed) {
					logging.LogInfo(
						h.logger,
						"startReadMessagesGoroutine завершена из-за закрытия соединения",
					)

					return
				}

				logging.LogErrorContext(
					h.goroutinesCtx,
					h.logger,
					"Не удалось прочитать событие",
					err,
				)

				continue
			}

			switch event.Type {
			case domains.WSEventNewMessage:
				var message domains.Message
				if err = json.Unmarshal(event.Payload, &message); err != nil {
					logging.LogErrorContext(
						h.goroutinesCtx,
						h.logger,
						"Не удалось распарсить сообщение из WS-события",
						err,
					)

					continue
				}

				runtime.EventsEmit(h.wailsCtx, newMessageEventName, message)
			case domains.WSEventMessageDeleted:
				var dto domains.MessageDeletedPayload
				if err = json.Unmarshal(event.Payload, &dto); err != nil {
					logging.LogErrorContext(
						h.goroutinesCtx,
						h.logger,
						"Не удалось распарсить payload удаления из WS-события",
						err,
					)

					continue
				}

				runtime.EventsEmit(h.wailsCtx, messageDeletedEventName, dto)
			case domains.WSEventMessageEdited:
				var dto domains.MessageEditedPayload
				if err = json.Unmarshal(event.Payload, &dto); err != nil {
					logging.LogErrorContext(
						h.goroutinesCtx,
						h.logger,
						"Не удалось распарсить payload редактирования из WS-события",
						err,
					)

					continue
				}

				runtime.EventsEmit(h.wailsCtx, messageEditedEventName, dto)
			case domains.WSEventReactionAdded:
				var dto domains.ReactionAddedPayload
				if err = json.Unmarshal(event.Payload, &dto); err != nil {
					logging.LogErrorContext(
						h.goroutinesCtx,
						h.logger,
						"Не удалось распарсить payload reaction_added из WS-события",
						err,
					)

					continue
				}

				runtime.EventsEmit(h.wailsCtx, reactionAddedEventName, dto)
			case domains.WSEventReactionRemoved:
				var dto domains.ReactionRemovedPayload
				if err = json.Unmarshal(event.Payload, &dto); err != nil {
					logging.LogErrorContext(
						h.goroutinesCtx,
						h.logger,
						"Не удалось распарсить payload reaction_removed из WS-события",
						err,
					)

					continue
				}

				runtime.EventsEmit(h.wailsCtx, reactionRemovedEventName, dto)
			default:
				logging.LogInfoContext(
					h.goroutinesCtx,
					h.logger,
					"Получен неизвестный тип WS-события",
					"type", event.Type,
				)
			}
		}
	}
}

func (h *Handler) refreshTokens() {
	defer h.wg.Done()

	ticker := time.NewTicker(refreshTokensInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.goroutinesCtx.Done():
			return
		case <-ticker.C:
			_, _ = h.useCases.RefreshTokens(h.goroutinesCtx)
		}
	}
}

func (h *Handler) updateChats() {
	defer h.wg.Done()

	ticker := time.NewTicker(updateChatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.goroutinesCtx.Done():
			return
		case <-ticker.C:
			chats, err := h.useCases.GetUserChats(h.goroutinesCtx, nil)
			if err != nil {
				logging.LogErrorContext(
					h.goroutinesCtx,
					h.logger,
					"Не удалось обновить список чатов",
					err,
				)
			}

			// Отправляем событие на фронтенд
			runtime.EventsEmit(h.wailsCtx, chatsUpdatedEventName, chats)
		}
	}
}
