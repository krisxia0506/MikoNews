package impl

import (
	"MikoNews/internal/model"
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/repository"
	"MikoNews/internal/service"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// articleService 实现了 ArticleService 接口
type articleService struct {
	repo repository.ArticleRepository
}

// NewArticleService 创建一个新的 articleService 实例
// 移除了 bot 参数
func NewArticleService(repo repository.ArticleRepository) service.ArticleService {
	return &articleService{
		repo: repo,
	}
}

// SaveSubmission 处理并保存用户通过飞书发送的投稿
func (s *articleService) SaveSubmission(ctx context.Context, authorID, authorName, title string, textContent string) (*model.Article, error) {
	now := time.Now()
	article := &model.Article{
		AuthorID:    authorID,
		AuthorName:  authorName,
		Title:       title,
		Content:     textContent,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.repo.Create(ctx, article) // 调用更新后的 Create 方法
	if err != nil {
		logger.Error("Failed to save submission to repository",
			zap.String("authorID", authorID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("保存投稿失败: %w", err)
	}

	logger.Info("Submission saved successfully", zap.Int64("articleID", article.ID))
	return article, nil
}

// FindArticleByID 根据ID查找文章
func (s *articleService) FindArticleByID(ctx context.Context, id int64) (*model.Article, error) {
	article, err := s.repo.FindByID(ctx, id) // 调用更新后的 FindByID 方法
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文章未找到 (ID: %d)", id)
		}
		logger.Error("Failed to find article by ID", zap.Int64("id", id), zap.Error(err))
		return nil, fmt.Errorf("查找文章失败: %w", err)
	}
	return article, nil
}
