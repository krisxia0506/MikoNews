package repository

import (
	"MikoNews/internal/model"
	"context"
)

// ArticleRepository 定义文章数据访问接口
type ArticleRepository interface {
	// Create 保存一篇新的文章投稿
	Create(ctx context.Context, article *model.Article) error

	// FindByID 根据ID查找文章
	FindByID(ctx context.Context, id int64) (*model.Article, error)
}
