package grpc

import (
	"context"
	"fmt"
)

type contextKey string

const (
	// CtxKeyUsername is the context key for the authenticated username.
	CtxKeyUsername contextKey = "username"
	// CtxKeyUserID is the context key for the authenticated user ID.
	CtxKeyUserID contextKey = "userID"
	// CtxKeyRole is the context key for the authenticated user's role.
	CtxKeyRole contextKey = "role"
)

func userIDFromContext(ctx context.Context) (int64, error) {
	v := ctx.Value(CtxKeyUserID)
	if v == nil {
		return 0, fmt.Errorf("userID not found in context")
	}
	id, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("userID has unexpected type")
	}
	return id, nil
}

func roleFromContext(ctx context.Context) string {
	v := ctx.Value(CtxKeyRole)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func usernameFromContext(ctx context.Context) string {
	v := ctx.Value(CtxKeyUsername)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
