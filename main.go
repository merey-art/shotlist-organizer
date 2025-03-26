package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Shot struct {
	ID          int    `json:"id"`
	Scene       string `json:"scene"`
	ShotType    string `json:"shot_type"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/shotlist?parseTime=true")
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("MySQL unreachable: %v", err)
	}

	log.Println("✅ Connected to MySQL!")
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()

	r.LoadHTMLGlob("public/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.Static("/uploads", "./uploads")

	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload image"})
			return
		}
		filePath := "uploads/" + file.Filename
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}
		imageURL := "http://localhost:8000/" + filePath
		c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
	})

	// Create a new shot
	r.POST("/shots", func(c *gin.Context) {
		scene := c.PostForm("scene")
		shotType := c.PostForm("shot_type")
		description := c.PostForm("description")

		file, err := c.FormFile("image")
		imageURL := ""
		if err == nil {
			filePath := "uploads/" + file.Filename
			if err := c.SaveUploadedFile(file, filePath); err == nil {
				imageURL = "http://localhost:8000/" + filePath
			}
		}

		_, err = db.Exec("INSERT INTO shots (scene, shot_type, description, image_url) VALUES (?, ?, ?, ?)",
			scene, shotType, description, imageURL)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save shot to database"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Shot created successfully"})
	})

	// Get all shots
	r.GET("/shots", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, scene, shot_type, description, image_url FROM shots")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve shots"})
			return
		}
		defer rows.Close()

		var shots []Shot
		for rows.Next() {
			var shot Shot
			if err := rows.Scan(&shot.ID, &shot.Scene, &shot.ShotType, &shot.Description, &shot.ImageURL); err != nil {
				continue
			}
			shots = append(shots, shot)
		}

		c.JSON(http.StatusOK, shots)
	})

	// Delete a shot
	r.DELETE("/shots/:id", func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM shots WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete shot"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Shot deleted"})
	})

	// Optional: API status
	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Shot List Organizer API is running"})
	})

	r.Run(":8000")
}
