package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
)

// User モデルの定義
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"-"` // パスワードはレスポンスに含めない
}

var db *gorm.DB

// データベースの初期化
func initDB() {
	dsn := "username:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("データベース接続失敗")
	}
	// テーブル作成
	db.AutoMigrate(&User{})
}

func main() {
	initDB()
	e := echo.New()

	// ミドルウェアの設定
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ルートエンドポイント
	e.GET("/users", getUsers)
	e.POST("/users", createUser)
	e.GET("/users/:id", getUserByID)
	e.DELETE("/users/:id", deleteUser)

	e.Start(":8080") // ポート8080で起動
}

// ユーザー一覧取得
func getUsers(c echo.Context) error {
	var users []User
	db.Find(&users)
	return c.JSON(http.StatusOK, users)
}

// ユーザー作成
func createUser(c echo.Context) error {
	user := new(User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "リクエストのバインドに失敗しました"})
	}
	if err := c.Validate(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "バリデーションエラー"})
	}

	// 【セキュリティ】パスワードのハッシュ化を行うべき（本番環境では必須）
	db.Create(&user)
	return c.JSON(http.StatusCreated, user)
}

// IDからユーザー取得
func getUserByID(c echo.Context) error {
	id := c.Param("id")
	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "ユーザーが見つかりません"})
	}
	return c.JSON(http.StatusOK, user)
}

// ユーザー削除
func deleteUser(c echo.Context) error {
	id := c.Param("id")
	db.Delete(&User{}, id)
	return c.JSON(http.StatusOK, echo.Map{"message": "ユーザーを削除しました"})
}

// 実行コマンド一覧
// go mod init echo_gorm_mysql
// go get -u github.com/labstack/echo/v4
// go get -u gorm.io/driver/mysql
// go get -u gorm.io/gorm
// go run main.go
