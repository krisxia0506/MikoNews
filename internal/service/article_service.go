package service

import (
	"MikoNews/internal/model"
	"context"
)

// ArticleService 定义文章业务逻辑接口
type ArticleService interface {
	// SaveSubmission 处理并保存用户通过飞书发送的投稿
	// 参数包括消息的唯一ID、来源会话ID、作者ID、作者名、标题、纯文本内容、原始富文本内容(JSON)
	// 返回创建的文章对象（如果成功）和错误
	SaveSubmission(ctx context.Context, authorID, authorName, title string, textContent string) (*model.Article, error)

	// FindArticleByID 根据ID查找文章 (如果需要此功能)
	FindArticleByID(ctx context.Context, id int64) (*model.Article, error)
}
