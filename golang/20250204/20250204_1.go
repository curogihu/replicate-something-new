packge main 

import (
    "context"
    "fmt"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/s3/types"
    "mime/multipart"
    "net/http"
    "os"
)

// File メタデータ管理用モデル
type File struct {
    ID          uint    `json:"id" gorm:"primaryKey"`
    Filename    string  `json:"filename"`
    URL         string  `json:"url"`
}

var db *gorm.DB
var s3Client *s3.Client
const bucketName = "your-s3-bucket-name"

// データベースの初期化
func initDB() {
    var err error
    db, err = gorm.Open(sqlite.Open("flies.db"), &gorm.Config{})
    if err != nil {
        panic("database connection failure")
    }
    db.AutoMigrate(&File)
}

// initialize s3 client
func initS3() {
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        panic("failed to load aws configuration")
    }
    s3Client = s3.NewFromConfig(cfg)
}

func main() {
    initDB()
    initS3()
    r := gin.Default()

    // root end point
    r.POST("/upload", uploadFile)
    r.GET("/files", getFiles)
    r.GET("files/:id", getFileByID)

    r.Run(":8080")
}

// file upload handler
func uploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "it could not fetch file"})
        return
    }

    // upload file to s3
    s3URL, err := uploadToS3(file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed upload file to s3"})
        return
    }

    // save file to database
    savedFile := File{Filename: file.Filename, URL: s3URL}
    db.Create(&savedFile)

    c.JSON(http.StatusCreated, savedFile)
}

// upload file to s3
func uploadToS3(file *multipart.FileHeader) (string, error) {
    openedFile, err := file.Open()
    if err != nil {
        return "", err
    }
    defer openedFile.Close()

    _, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(file.Filename),
        Body:   openedFile,
        ACL:    types.ObjectCannedACLPublicRead,
    })
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, file.Filename), nil
}

// get all files
func getFiles(c *gin.Context) {
    var files []File
    db.Find(&files)
    c.JSON(http.StatusOK, files)
}

// get file by id
func getFileByID(c *gin.Context) {
    id := c.Param("id")
    var file File
    result := db.First(&file, id)
    if result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
        return
    }
    c.JSON(http.StatusOK, file)
}
