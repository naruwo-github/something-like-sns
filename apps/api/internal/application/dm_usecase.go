package application

import (
	"context"
	"errors"
	"strings"

	"github.com/example/something-like-sns/apps/api/internal/domain"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type dmUsecase struct {
	dmRepo        port.DMRepository
	cursorEncoder port.CursorEncoder
}

func NewDMUsecase(dr port.DMRepository, ce port.CursorEncoder) port.DMUsecase {
	return &dmUsecase{dmRepo: dr, cursorEncoder: ce}
}

func (u *dmUsecase) GetOrCreateDM(ctx context.Context, scope domain.Scope, otherUserID uint64) (uint64, error) {
	if otherUserID == 0 || otherUserID == scope.UserID {
		return 0, errors.New("invalid other_user_id")
	}

	convID, err := u.dmRepo.FindDMConversation(ctx, scope.TenantID, scope.UserID, otherUserID)
	if err != nil {
		return 0, err
	}
	if convID != 0 {
		return convID, nil
	}

	return u.dmRepo.CreateDMConversation(ctx, scope.TenantID, scope.UserID, otherUserID)
}

func (u *dmUsecase) ListConversations(ctx context.Context, scope domain.Scope, token string) ([]*domain.Conversation, string, error) {
	const limit = 20
	cursorTime, cursorID, err := u.cursorEncoder.Decode(token)
	if err != nil {
		return nil, "", err
	}

	convos, err := u.dmRepo.FindConversations(ctx, scope.TenantID, scope.UserID, limit, cursorTime, cursorID)
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

	messages, err := u.dmRepo.FindMessages(ctx, scope.TenantID, conversationID, limit, cursorTime, cursorID)
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
	return u.dmRepo.CreateMessage(ctx, scope.TenantID, conversationID, scope.UserID, body)
}
