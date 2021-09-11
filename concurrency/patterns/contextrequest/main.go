package main

import (
	"context"
	"fmt"
)

func main() {
	processRequest("jane", "abc123")
}

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxAuthToken
)

func userID(ctx context.Context) string {
	return ctx.Value(ctxUserID).(string)
}

func authToken(ctx context.Context) string {
	return ctx.Value(ctxAuthToken).(string)
}

func processRequest(userID, authToken string) {
	ctx := context.WithValue(context.Background(), ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxAuthToken, authToken)
	handleResponse(ctx)
}

func handleResponse(ctx context.Context) {
	fmt.Printf(
		"handling response for %v (%v)",
		userID(ctx),
		authToken(ctx),
	)
}
