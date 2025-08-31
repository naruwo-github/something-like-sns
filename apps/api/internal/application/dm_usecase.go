package application

import (
	"context"
	"errors"
	"strings"

	"github.com/example/something-like-sns/apps/api/internal/domain"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type dmUsecase struct {
	store         port.Store
	cursorEncoder port.CursorEncoder
}

func NewDMUsecase(store port.Store, ce port.CursorEncoder) port.DMUsecase {
	return &dmUsecase{store: store, cursorEncoder: ce}
}

func (u *dmUsecase) GetOrCreateDM(ctx context.Context, scope domain.Scope, otherUserID uint64) (uint64, error) {
	if otherUserID == 0 || otherUserID == scope.UserID {
		return 0, errors.New("invalid other_user_id")
	}

	var convID uint64
	err := u.store.ExecTx(ctx, func(s port.Store) error {
		var err error
		convID, err = s.DMRepository().FindDMConversation(ctx, scope.TenantID, scope.UserID, otherUserID)
		if err != nil {
			return err
		}
		if convID != 0 {
			return nil
		}

		convID, err = s.DMRepository().CreateDMConversation(ctx, scope.TenantID, scope.UserID, otherUserID)
		return err
	})

	return convID, err
}

func (u *dmUsecase) ListConversations(ctx context.Context, scope domain.Scope, token string) ([]*domain.Conversation, string, error) {
	const limit = 20
	cursorTime, cursorID, err := u.cursorEncoder.Decode(token)
	if err != nil {
		return nil, "", err
	}

	convos, err := u.store.DMRepository().FindConversations(ctx, scope.TenantID, scope.UserID, limit, cursorTime, cursorID)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(convos) == limit {
		lastConvo := convos[len(convos)-1]
		nextToken = u.cursorEncoder.Encode(lastConvo.CreatedAt, lastConvo.ID)
	}

	return convos, nextToken, nil
}

func (u *dmUsecase) ListMessages(ctx context.Context, scope domain.Scope, conversationID uint64, token string) ([]*domain.Message, string, error) {
	const limit = 50
	cursorTime, cursorID, err := u.cursorEncoder.Decode(token)
	if err != nil {
		return nil, "", err
	}

	messages, err := u.store.DMRepository().FindMessages(ctx, scope.TenantID, conversationID, limit, cursorTime, cursorID)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(messages) == limit {
		lastMessage := messages[len(messages)-1]
		nextToken = u.cursorEncoder.Encode(lastMessage.CreatedAt, lastMessage.ID)
	}

	return messages, nextToken, nil
}

func (u *dmUsecase) SendMessage(ctx context.Context, scope domain.Scope, conversationID uint64, body string) (*domain.Message, error) {
	body = strings.TrimSpace(body)
	if body == "" || len(body) > 2000 {
		return nil, errors.New("invalid body")
	}
	// TODO: Check if user is a member of the conversation
	return u.store.DMRepository().CreateMessage(ctx, scope.TenantID, conversationID, scope.UserID, body)
}
