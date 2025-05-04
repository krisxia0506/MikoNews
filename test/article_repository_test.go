package test

import (
	"MikoNews/internal/config"
	"MikoNews/internal/database"
	"MikoNews/internal/model"
	"MikoNews/internal/pkg/logger"
	"MikoNews/internal/repository"
	"MikoNews/internal/repository/impl/mysql"
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
)

// TestMain 用于测试前的初始化工作
func TestMain(m *testing.M) {
	setupTestEnv()
	exitCode := m.Run()
	teardownTestEnv()
	os.Exit(exitCode)
}

// 全局变量，测试期间使用
var (
	dbConfig *config.DatabaseConfig
	db       *database.DB
	repo     repository.ArticleRepository
	testCtx  context.Context
)

// setupTestEnv 设置测试环境
func setupTestEnv() {
	// 1. 初始化 Logger (必须在 database.New 之前)
	// logger.InitLogger 需要 logPath 和 logLevel 两个字符串参数
	testLogPath := "../test-logs/test.log" // 指定测试日志文件路径
	testLogLevel := "debug"                // 测试时使用 debug 级别
	// 确保日志目录存在
	if _, err := os.Stat("../test-logs"); os.IsNotExist(err) {
		_ = os.Mkdir("../test-logs", 0755)
	}
	logger.InitLogger(testLogPath, testLogLevel)

	// 2. 设置数据库配置
	dbConfig = &config.DatabaseConfig{
		Host:         "58.87.95.28",
		Port:         3306,
		User:         "miko_news",
		Password:     "mJTpfLp8s4382TWt",
		DBName:       "miko_news",
		MaxOpenConns: 5,
		MaxIdleConns: 3,
	}

	// 3. 连接数据库 (现在 logger 已初始化)
	var err error
	db, err = database.New(dbConfig)
	if err != nil {
		log.Fatalf("连接测试数据库失败: %v", err)
	}

	// // 4. 自动迁移
	// err = db.DB.AutoMigrate(&model.Article{})
	// if err != nil {
	// 	log.Fatalf("自动迁移失败: %v", err)
	// }

	// 5. 创建仓库实例
	repo = mysql.NewArticleRepository(db.DB)

	// 6. 创建上下文
	testCtx = context.Background()

	// 7. 清理测试数据
	cleanTestData()
}

// teardownTestEnv 清理测试环境
func teardownTestEnv() {
	if db != nil {
		cleanTestData()
		if err := db.Close(); err != nil {
			log.Printf("关闭测试数据库连接时出错: %v", err)
		}
	}
}

// cleanTestData 清理测试数据
func cleanTestData() {
	db.DB.Unscoped().Where("title LIKE ?", "Test Article%").Delete(&model.Article{})
}

// TestArticleRepository_CreateAndFind 测试文章的创建和查找
func TestArticleRepository_CreateAndFind(t *testing.T) {
	// 1. 创建测试文章
	testArticle := createTestArticle(t, "Test Article CRUD")

	// 2. 按 ID 读取文章
	readArticleByID(t, testArticle.ID, testArticle.Title)

	// 3. 测试查找不存在的记录
	testFindNotFound(t)
}

// createTestArticle 创建测试文章
func createTestArticle(t *testing.T, title string) *model.Article {
	now := time.Now()
	testContent := fmt.Sprintf("这是 %s 的测试内容。", title)

	article := &model.Article{
		AuthorID:   "test_user_1231",
		AuthorName: "测试用户",
		Title:      title,
		Content:    testContent,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := repo.Create(testCtx, article)
	if err != nil {
		t.Fatalf("创建文章失败 (Title: %s): %v", title, err)
	}

	if article.ID <= 0 {
		t.Fatalf("创建文章后返回的ID无效: %d", article.ID)
	}

	t.Logf("创建文章成功，ID: %d, Title: %s", article.ID, article.Title)
	return article
}

// readArticleByID 按 ID 读取文章并验证
func readArticleByID(t *testing.T, id int64, expectedTitle string) *model.Article {
	article, err := repo.FindByID(testCtx, id)
	if err != nil {
		t.Fatalf("按 ID 读取文章失败 (ID: %d): %v", id, err)
	}

	if article == nil {
		t.Fatalf("没有找到 ID 为 %d 的文章", id)
	}

	if article.ID != id {
		t.Errorf("文章 ID 不匹配，期望: %d, 实际: %d", id, article.ID)
	}
	if article.Title != expectedTitle {
		t.Errorf("文章 Title 不匹配，期望: %s, 实际: %s", expectedTitle, article.Title)
	}
	if article.Content == "" {
		t.Errorf("文章 Content 不应为空字符串")
	}

	t.Logf("按 ID 读取文章成功 (ID: %d)，标题: %s", article.ID, article.Title)
	return article
}

// testFindNotFound 测试查找不存在的记录
func testFindNotFound(t *testing.T) {
	_, err := repo.FindByID(testCtx, 999999)
	if err == nil {
		t.Error("期望按不存在的 ID 查找时出错，但没有错误发生")
	} else if err != gorm.ErrRecordNotFound {
		t.Errorf("期望按不存在的 ID 查找时返回 gorm.ErrRecordNotFound，实际返回: %v", err)
	} else {
		t.Logf("成功捕获按不存在的 ID 查找的错误: %v", err)
	}
}
