package context

import (
	"context"

	"github.com/eitah/lenslocked/src/lenslocked.com/models"
)

type privateKey string

const (
	// Our privateKey type, while backed by a string, is not actu­ally
	// the same as a string, and the keys used for the context package
	// take both the type and the value into consideration.
	userKey privateKey = "user"
)

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	if temp := ctx.Value(userKey); temp != nil {
		if user, ok := temp.(*models.User); ok {
			return user
		}
	}
	return nil
}
