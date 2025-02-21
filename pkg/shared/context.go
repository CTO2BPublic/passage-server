package shared

import (
	"context"

	"github.com/google/uuid"
)

// Define a custom type for context key
type contextKey string

// Define context keys
const (
	transactionIDKey contextKey = "transactionID"
	userIDKey        contextKey = "userID"
	lockKey          contextKey = "lock"
)

// Adds new transactionId to the context if one is not present
func WithTransactionID(ctx context.Context, transactionID ...string) context.Context {
	_, ok := GetTransactionID(ctx)
	if !ok {

		txid := uuid.New().String()
		if len(transactionID) > 0 {
			txid = transactionID[0]
		}
		ctx = context.WithValue(ctx, transactionIDKey, txid)
	}

	return ctx
}

// Adds new userID to the context if one is not present
func WithUserID(ctx context.Context, userID ...string) context.Context {
	_, ok := GetTransactionID(ctx)
	if !ok {

		uid := uuid.New().String()
		if len(userID) > 0 {
			uid = userID[0]
		}
		ctx = context.WithValue(ctx, userIDKey, uid)
	}

	return ctx
}

// GetTransactionID retrieves transaction ID from context
func GetTransactionID(ctx context.Context) (string, bool) {
	transactionID, ok := ctx.Value(transactionIDKey).(string)
	return transactionID, ok
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
