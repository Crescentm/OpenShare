package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Folder 文件夹表
type Folder struct {
	ID        uuid.UUID      `gorm:"type:uuid;primarykey" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null;index:idx_folders_parent_name,unique,priority:2" json:"name"` // 文件夹名，同级目录下唯一
	ParentID  *uuid.UUID     `gorm:"type:uuid;index:idx_folders_parent_name,unique,priority:1" json:"parent_id,omitempty"`   // 父文件夹ID，与 Name 组成联合唯一索引
	Path      string         `gorm:"type:varchar(1000);not null;uniqueIndex" json:"path"`                                    // 完整路径，全局唯一
	Status    string         `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Parent   *Folder  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Folder `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Files    []File   `gorm:"foreignKey:FolderID" json:"files,omitempty"`
	Tags     []Tag    `gorm:"many2many:folder_tags" json:"tags,omitempty"`
}

// BeforeCreate 创建前自动生成 UUID
func (f *Folder) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// TableName 指定表名
func (Folder) TableName() string {
	return "folders"
}
