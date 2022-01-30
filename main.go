package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()

	r.POST("/url", postUrl)
	r.GET("/urls", getUrls)
	r.DELETE("/url", deleteUrl)
	// r.GET("/search", getSearch)
	r.Run()
}

func createClient(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := "precise-cabinet-280004"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done with
	// defer client.Close()
	return client
}

// PostUrlBody is the body of the requests
type PostUrlBody struct {
	UserId string `json:"user_id" binding:"required"`
	Url    string `json:"url" binding:"required"`
}

type Url struct {
	Url string `json:"url" binding:"required"`
	Id  string `json:"id" binding:"required"`
}

func deleteUrl(c *gin.Context) {
	// get the input
	id, ok := c.GetQuery("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is a required query param"})
		return
	}

	// get the client
	ctx := context.Background()
	client := createClient(ctx)
	defer client.Close()

	// delete the document
	_, err := client.Collection("urls").Doc(id).Delete(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func getUrls(c *gin.Context) {
	// get the input
	userId, ok := c.GetQuery("user_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is a required query param"})
		return
	}

	// get the client
	ctx := context.Background()
	client := createClient(ctx)
	defer client.Close()

	fmt.Println(userId)

	// get the urls
	urls := make([]Url, 0)
	iter := client.Collection("urls").Where("user_id", "==", userId).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			fmt.Println("finished iterating")
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data := doc.Data()
		urls = append(urls, Url{
			Url: data["user_id"].(string),
			Id:  data["id"].(string),
		})
	}
	c.JSON(http.StatusOK, gin.H{"urls": urls})
}

func postUrl(c *gin.Context) {
	// get the input
	var postUrlBody PostUrlBody
	if err := c.BindJSON(&postUrlBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get the client
	ctx := context.Background()
	client := createClient(ctx)
	defer client.Close()

	// todo: scrape the website
	// save the content
	id := uuid.New().String()
	_, err := client.Collection("urls").Doc(id).Set(ctx, map[string]interface{}{
		"id":      id,
		"user_id": postUrlBody.UserId,
		"url":     postUrlBody.Url,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
