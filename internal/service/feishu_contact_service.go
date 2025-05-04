package service

import (
	"context"

	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
)

// FeishuContactService defines the interface for interacting with Feishu contact APIs.
type FeishuContactService interface {
	// GetUserInfoByOpenID fetches user details using their Open ID.
	// It returns the User object from the SDK or an error.
	GetUserInfoByOpenID(ctx context.Context, openID string) (*larkcontact.User, error)
}
