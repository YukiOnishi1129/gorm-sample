package my

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// Migrate program
func Migrate() {
	// gorm.Open: DB接続 インスタンス生成
	db, er := gorm.Open("sqlite3", "data.sqlite3")
	if er != nil {
		fmt.Println(er)
		return
	}

	defer db.Close()

	// Migration実施
	db.AutoMigrate(&User{}, &Group{}, &Post{}, &Comment{})
}