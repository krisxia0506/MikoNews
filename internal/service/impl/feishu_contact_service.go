package impl

import (
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/service"
	"context"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"go.uber.org/zap"
)

// feishuContactServiceImpl implements the FeishuContactService interface.
type feishuContactServiceImpl struct {
	client *lark.Client
	// No logger field, using global logger functions
}

// NewFeishuContactService creates a new Feishu contact service implementation.
func NewFeishuContactService(client *lark.Client) service.FeishuContactService {
	return &feishuContactServiceImpl{
		client: client,
	}
}

// GetUserInfoByOpenID fetches user details using their Open ID.
func (s *feishuContactServiceImpl) GetUserInfoByOpenID(ctx context.Context, openID string) (*larkcontact.User, error) {
	// 1. Create the request object
	req := larkcontact.NewGetUserReqBuilder().
		UserId(openID).
		UserIdType(larkcontact.UserIdTypeOpenId). // Specify we are using open_id
		// DepartmentIdType could be set if needed, but often not necessary when querying by user ID directly
		// DepartmentIdType(larkcontact.DepartmentIdTypeOpenDepartmentId).
		Build()

	// 2. Make the API call
	resp, err := s.client.Contact.V3.User.Get(ctx, req)

	// 3. Handle potential errors during the API call
	if err != nil {
		logger.Error("Failed to call Feishu Contact API", zap.String("openID", openID), zap.Error(err))
		return nil, fmt.Errorf("飞书联系人 API 调用失败: %w", err)
	}

	// 4. Handle unsuccessful responses from the Feishu server
	if !resp.Success() {
		logger.Error("Feishu Contact API call unsuccessful",
			zap.String("openID", openID),
			zap.String("requestID", resp.RequestId()),
			zap.Int("code", resp.Code),
			zap.String("msg", resp.Msg),
			// zap.Any("errorDetail", resp.CodeError), // Uncomment for more detail if needed
		)
		return nil, fmt.Errorf("获取飞书用户信息失败: %s (code: %d, request_id: %s)", resp.Msg, resp.Code, resp.RequestId())
	}

	// 5. Handle success case
	if resp.Data == nil || resp.Data.User == nil {
		logger.Warn("Feishu Contact API successful but user data is nil", zap.String("openID", openID), zap.String("requestID", resp.RequestId()))
		return nil, fmt.Errorf("获取飞书用户信息成功，但用户数据为空 (request_id: %s)", resp.RequestId())
	}

	logger.Debug("Successfully fetched Feishu user info", zap.String("openID", openID), zap.String("userName", *resp.Data.User.Name))
	return resp.Data.User, nil
}

// Ensure feishuContactServiceImpl implements FeishuContactService
var _ service.FeishuContactService = (*feishuContactServiceImpl)(nil)
