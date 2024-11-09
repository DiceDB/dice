package main

import (
	"context"
	"encoding/json"
	
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/dicedb/dicedb-go" // DiceDB Go SDK
)

type URL struct {
	ID       string `json:"id"`
	LongURL  string `json:"long_url"`
	ShortURL string `json:"short_url"`
}

var db *dicedb.Client

// Initialize DiceDB connection
func init() {
	db = dicedb.NewClient(&dicedb.Options{
		Addr: "localhost:7379",
	})
}

// Creates a short URL from a given long URL
func CreateShortURL(c *gin.Context) {
	var requestBody URL
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Generate unique short ID and construct the short URL
	shortID := uuid.New().String()[:8]
	requestBody.ID = shortID
	requestBody.ShortURL = "http://localhost:8080/" + shortID

	// Serialize URL struct to JSON and store it in DiceDB
	urlData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
		return
	}

	if err := db.Set(context.Background(), shortID, urlData, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"short_url": requestBody.ShortURL})
}

// Redirects to the original URL based on the short URL ID
func RedirectURL(c *gin.Context) {
	id := c.Param("id")

	// Retrieve stored URL data from DiceDB
	urlData, err := db.Get(context.Background(), id).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// Deserialize JSON data back into URL struct
	var url URL
	if err := json.Unmarshal([]byte(urlData), &url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode URL data"})
		return
	}

	// Redirect user to the original long URL
	c.Redirect(http.StatusFound, url.LongURL)
}

func main() {
	router := gin.Default()

	// Define endpoints for creating short URLs and redirecting
	router.POST("/shorten", CreateShortURL)
	router.GET("/:id", RedirectURL)

	// Start the server on port 8080
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
