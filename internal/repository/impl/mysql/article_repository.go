package mysql

import (
	"MikoNews/internal/model"
	"MikoNews/internal/repository"
	"context"

	"gorm.io/gorm"
)

// articleRepository 实现了 ArticleRepository 接口
type articleRepository struct {
	db *gorm.DB
}

// NewArticleRepository 创建一个新的 articleRepository 实例
func NewArticleRepository(db *gorm.DB) repository.ArticleRepository {
	return &articleRepository{db: db}
}

// Create 保存一篇新的文章投稿
func (r *articleRepository) Create(ctx context.Context, article *model.Article) error {
	result := r.db.WithContext(ctx).Create(article)
	return result.Error
}

// FindByID 根据ID查找文章
func (r *articleRepository) FindByID(ctx context.Context, id int64) (*model.Article, error) {
	var article model.Article
	result := r.db.WithContext(ctx).First(&article, id)
	if result.Error != nil {
		return nil, result.Error // GORM 会自动处理 ErrRecordNotFound
	}
	return &article, nil
}