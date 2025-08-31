package rpc

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/example/something-like-sns/apps/api/gen/sns/v1"
	"github.com/example/something-like-sns/apps/api/gen/sns/v1/v1connect"
	"github.com/example/something-like-sns/apps/api/internal/port"
)

type DMHandler struct {
	dmUsecase port.DMUsecase
}

func NewDMHandler(du port.DMUsecase) *DMHandler {
	return &DMHandler{dmUsecase: du}
}

func (s *DMHandler) MountHandler(authInterceptor connect.Interceptor) (string, http.Handler) {
	path, h := v1connect.NewDMServiceHandler(s, connect.WithInterceptors(authInterceptor))
	return path, h
}

func (s *DMHandler) GetOrCreateDM(ctx context.Context, req *connect.Request[v1.GetOrCreateDMRequest]) (*connect.Response[v1.GetOrCreateDMResponse], error) {
	scope := GetScopeFromContext(ctx)

	convID, err := s.dmUsecase.GetOrCreateDM(ctx, scope, req.Msg.GetOtherUserId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.GetOrCreateDMResponse{ConversationId: convID}), nil
}

func (s *DMHandler) ListConversations(ctx context.Context, req *connect.Request[v1.ListConversationsRequest]) (*connect.Response[v1.ListConversationsResponse], error) {
	scope := GetScopeFromContext(ctx)

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
	scope := GetScopeFromContext(ctx)

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
	scope := GetScopeFromContext(ctx)

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