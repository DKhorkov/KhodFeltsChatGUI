package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/DKhorkov/kfcGUI/internal/common"
	"github.com/DKhorkov/kfcGUI/internal/domains"
	customerrors "github.com/DKhorkov/kfcGUI/internal/errors"
	"github.com/DKhorkov/kfcGUI/internal/interfaces"
)

type ChatsRepository struct {
	Repository

	httpClient interfaces.HTTPClient
	baseURL    string
	mu         sync.RWMutex
}

func NewChatsRepository(httpClient interfaces.HTTPClient, baseURL string) *ChatsRepository {
	return &ChatsRepository{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (r *ChatsRepository) GetUserChats(
	ctx context.Context,
	accessToken string,
	limit, offset int,
) ([]domains.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/chats?limit=%d&offset=%d", r.baseURL, limit, offset),
		http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer r.closeBody(ctx, resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(data))
	}

	var chats []domains.Chat
	if err = json.Unmarshal(data, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *ChatsRepository) CreateChat(
	ctx context.Context,
	accessToken string,
	chat domains.Chat,
) (*domains.Chat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !chat.IsValid() {
		return nil, fmt.Errorf("%w: chat is not valid: %v+", customerrors.ErrCreateChat, chat)
	}

	body, err := json.Marshal(chat)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.baseURL+"/chats",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(common.ContentTypeHeaderName, common.ApplicationJSONContentType)
	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer r.closeBody(ctx, resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(string(data))
	}

	var createdChat domains.Chat
	if err = json.Unmarshal(data, &createdChat); err != nil {
		return nil, err
	}

	return &createdChat, nil
}

func (r *ChatsRepository) GetChatMessages(
	ctx context.Context,
	accessToken string,
	chatID uint64,
	limit, offset int,
) ([]domains.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/chats/%d/messages?limit=%d&offset=%d", r.baseURL, chatID, limit, offset),
		http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	req.AddCookie(
		&http.Cookie{
			Name:  accessTokenCookieName,
			Value: accessToken,
		},
	)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.closeBody(ctx, resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(data))
	}

	var messages []domains.Message
	if err = json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
