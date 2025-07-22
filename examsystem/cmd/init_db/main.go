package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// 清空数据库所有数据（保留表结构）
func clearAllData(db *gorm.DB) error {
	var tables []struct{ Name string }
	if err := db.Raw("SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE'sqlite_%'").Scan(&tables).Error; err != nil {
		return fmt.Errorf("获取表名失败: %v", err)
	}

	if err := db.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		return fmt.Errorf("禁用外键约束失败: %v", err)
	}
	defer db.Exec("PRAGMA foreign_keys = ON")

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", table.Name)).Error; err != nil {
			return fmt.Errorf("清空 %s 表失败: %v", table.Name, err)
		}
		if err := db.Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s'", table.Name)).Error; err != nil {
			log.Printf("重置 %s 表自增ID失败: %v", table.Name, err)
		}
	}
	return nil
}

// 在事务中执行SQL脚本
func executeSQLScriptInTransaction(db *gorm.DB, sqlScript string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	statements := splitSQLStatements(sqlScript)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err = tx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("执行SQL语句失败: %v\n语句: %s", err, stmt)
		}
	}
	return nil
}

// 分割SQL语句
func splitSQLStatements(sqlScript string) []string {
	var statements []string
	var currentStmt strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range sqlScript {
		if r == '\'' || r == '"' || r == '`' {
			if !inQuote {
				inQuote = true
				quoteChar = r
			} else if r == quoteChar {
				prevChar := currentStmt.Len() - 1
				if prevChar >= 0 && rune(currentStmt.String()[prevChar]) == '\\' {
				} else {
					inQuote = false
				}
			}
		}

		currentStmt.WriteRune(r)

		if !inQuote && r == ';' {
			statements = append(statements, currentStmt.String())
			currentStmt.Reset()
		}
	}

	if currentStmt.Len() > 0 {
		statements = append(statements, currentStmt.String())
	}
	return statements
}

func main() {
	dbPath := "../../examsystem.db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("打开数据库失败:", err)
	}

	sqlFilePath := filepath.Join("../../migrations", "init.sql")
	sqlContent, err := ioutil.ReadFile(sqlFilePath)
	if err != nil {
		log.Fatalf("读取init.sql失败: %v", err)
	}

	if err := executeSQLScriptInTransaction(db, string(sqlContent)); err != nil {
		log.Fatalf("执行init.sql脚本失败: %v", err)
	}

	if err := clearAllData(db); err != nil {
		log.Fatalf("清空数据失败: %v", err)
	}

	var count int64
	db.Table("users").Where("username =?", "admin").Count(&count)
	if count == 0 {
		hash := sha256.Sum256([]byte("123456"))
		hashedPassword := hex.EncodeToString(hash[:])

		adminUser := map[string]interface{}{
			"username":      "admin",
			"password_hash": hashedPassword,
			"role":          "admin",
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		}

		if err := db.Table("users").Create(adminUser).Error; err != nil {
			log.Fatalf("创建默认用户失败: %v", err)
		}
	}
}
