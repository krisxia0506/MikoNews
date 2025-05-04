package model

import (
	"time"
)

// Article 代表存储的文章投稿信息 (与 init.sql 同步)
type Article struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title      string    `gorm:"column:title;type:varchar(255);not null;default:''" json:"title"`                         // 文章标题
	Content    string    `gorm:"column:content;type:text;not null;" json:"content"`                                       // 文章内容 (非指针，匹配 NOT NULL)
	AuthorID   string    `gorm:"column:author_id;type:varchar(64);not null;default:'';index:idx_author" json:"author_id"` // 作者飞书OpenID
	AuthorName string    `gorm:"column:author_name;type:varchar(64);not null;default:'匿名用户'" json:"author_name"`          // 作者名字
	CreatedAt  time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定 GORM 使用的表名
func (Article) TableName() string {
	return "articles"
}

/* // 旧的构造函数，暂时移除，Service层将处理创建逻辑
// NewArticle 创建新的文章对象
func NewArticle(title, content, authorID, authorName string) *Article {
	now := time.Now()
	return &Article{
		Title:      title,
		Content:    content, // 这里需要处理 string -> *string
		AuthorID:   authorID,
		AuthorName: authorName,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
*/
