package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var db *gorm.DB
var authClient *auth.Client

// User モデル
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name"`
	Email    string `json:"email" gorm:"unique"`
}

// データベース初期化
func initDB() {
	dsn := "username:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("データベース接続失敗")
	}
	db.AutoMigrate(&User{})
}

// Firebase初期化
func initFirebase() {
	opt := option.WithCredentialsFile("path/to/firebase-service-account.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic("Firebase 初期化失敗")
	}
	authClient, err = app.Auth(context.Background())
	if err != nil {
		panic("Firebase 認証クライアント取得失敗")
	}
}

func main() {
	initDB()
	initFirebase()
	r := gin.Default()

	r.POST("/register", registerUser)
	r.POST("/login", loginUser)
	r.GET("/profile", authMiddleware(), getProfile)

	r.Run(":8080")
}

// ユーザー登録処理
func registerUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエスト"})
		return
	}
	db.Create(&user)
	c.JSON(http.StatusCreated, user)
}

// ログイン処理
func loginUser(c *gin.Context) {
	idToken := c.PostForm("idToken")
	if idToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "トークンが必要です"})
		return
	}

	token, err := authClient.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ログイン成功", "uid": token.UID})
}

// 認証ミドルウェア
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		idToken := strings.TrimSpace(strings.Replace(c.GetHeader("Authorization"), "Bearer", "", 1))
		if idToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "認証トークンが必要です"})
			c.Abort()
			return
		}
		_, err := authClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "認証失敗"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ユーザープロフィール取得
func getProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "認証済みユーザーです"})
}

// 実行コマンド一覧
// go mod init gin_gorm_firebase
// go mod tidy
// go run main.go
