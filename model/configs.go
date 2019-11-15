package model

// Configs 全局参数配置
type Configs struct {
	ID        uint   `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Namespace string `gorm:"type:varchar(32)"`
	Field     string `gorm:"type:varchar(64)"`
	Type      string `gorm:"type:enum('string','int','float','json')"`
	Value     string `gorm:"type:varchar(512)"`
}
