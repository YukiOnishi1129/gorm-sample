package main

import (
	"Go/go-gorm/my"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	my.Migrate()
}