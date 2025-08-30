package rpc

import (
	"context"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	"github.com/example/something-like-sns/apps/api/gen/sns/v1/v1connect"
	"github.com/example/something-like-sns/apps/api/internal/domain"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type DMHandler struct {
	authUsecase     port.AuthUsecase
	dmUsecase       port.DMUsecase
	allowDevHeaders bool
}

func NewDMHandler(au port.AuthUsecase, du port.DMUsecase, allowDev bool) *DMHandler {
	return &DMHandler{authUsecase: au, dmUsecase: du, allowDevHeaders: allowDev}
}

func (s *DMHandler) MountHandler() (string, http.Handler) {
	path, h := v1connect.NewDMServiceHandler(s)
	return path, h
}

func (s *DMHandler) getScope(ctx context.Context, h http.Header) (domain.Scope, error) {
	if !s.allowDevHeaders {
		return domain.Scope{}, connect.NewError(connect.CodeUnauthenticated, errors.New("dev headers disabled"))
	}
	tenantSlug := h.Get("X-Tenant")
	authSub := h.Get("X-User")
	scope, err := s.authUsecase.ResolveScope(ctx, tenantSlug, authSub)
	if err != nil {
		return domain.Scope{}, connect.NewError(connect.CodeUnauthenticated, err)
	}
	return *scope, nil
}

func (s *DMHandler) GetOrCreateDM(ctx context.Context, req *connect.Request[v1.GetOrCreateDMRequest]) (*connect.Response[v1.GetOrCreateDMResponse], error) {
	scope, err := s.getScope(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	convID, err := s.dmUsecase.GetOrCreateDM(ctx, scope, req.Msg.GetOtherUserId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.GetOrCreateDMResponse{ConversationId: convID}), nil
}

func (s *DMHandler) ListConversations(ctx context.Context, req *connect.Request[v1.ListConversationsRequest]) (*connect.Response[v1.ListConversationsResponse], error) {
	scope, err := s.getScope(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	convos, nextToken, err := s.dmUsecase.ListConversations(ctx, scope, req.Msg.GetCursor().GetToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*v1.Conversation, len(convos))
	for i, c := range convos {
		items[i] = &v1.Conversation{
			Id:            c.ID,
			CreatedAt:     c.CreatedAt.Format(time.RFC3339Nano),
			MemberUserIds: c.MemberUserIDs,
		}
	}

	res := &v1.ListConversationsResponse{Items: items}
	if nextToken != "" {
		res.Next = &v1.Cursor{Token: nextToken}
	}
	return connect.NewResponse(res), nil
}

func (s *DMHandler) ListMessages(ctx context.Context, req *connect.Request[v1.ListMessagesRequest]) (*connect.Response[v1.ListMessagesResponse], error) {
	scope, err := s.getScope(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	messages, nextToken, err := s.dmUsecase.ListMessages(ctx, scope, req.Msg.GetConversationId(), req.Msg.GetCursor().GetToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*v1.Message, len(messages))
	for i, m := range messages {
		items[i] = &v1.Message{
			Id:             m.ID,
			ConversationId: m.ConversationID,
			SenderUserId:   m.SenderUserID,
			Body:           m.Body,
			CreatedAt:      m.CreatedAt.Format(time.RFC3339Nano),
		}
	}

	res := &v1.ListMessagesResponse{Items: items}
	if nextToken != "" {
		res.Next = &v1.Cursor{Token: nextToken}
	}
	return connect.NewResponse(res), nil
}

func (s *DMHandler) SendMessage(ctx context.Context, req *connect.Request[v1.SendMessageRequest]) (*connect.Response[v1.SendMessageResponse], error) {
	scope, err := s.getScope(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	msg, err := s.dmUsecase.SendMessage(ctx, scope, req.Msg.GetConversationId(), req.Msg.GetBody())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&v1.SendMessageResponse{
		Message: &v1.Message{
			Id:             msg.ID,
			ConversationId: msg.ConversationID,
			SenderUserId:   msg.SenderUserID,
			Body:           msg.Body,
			CreatedAt:      msg.CreatedAt.Format(time.RFC3339Nano),
		},
	}), nil
}