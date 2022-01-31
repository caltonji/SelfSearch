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
	"time"
	"sort"
	"strings"
)


var cache map[string][]FullUrl
const ExamplePadding = 100

func main() {
	cache = make(map[string][]FullUrl)
	r := gin.Default()


	r.POST("/url", postUrl)
	r.GET("/urls", getUrls)
	r.DELETE("/url", deleteUrl)
	r.GET("/search", getSearch)
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
	CreatedAt time.Time `json:"created_at" binding:"required"`
	Title string `json:"title" binding:"required"`
	Image string `json:"image" binding:"required"`
	Id  string `json:"id" binding:"required"`
}

type FullUrl struct {
	Id  string `json:"id" binding:"required"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
	Url string `json:"url" binding:"required"`
	Title string `json:"title" binding:"required"`
	Image string `json:"image" binding:"required"`
	UserId string  `json:"user_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type SearchResult struct {
	Url string `json:"url" binding:"required"`
	Title string `json:"title" binding:"required"`
	Image string `json:"image" binding:"required"`
	Occurrences int `json:"occurrences" binding:"required"`
	ExampleText string `json:"example_text" binding:"required"`
}


func getSearch(c *gin.Context) {
	q, ok := c.GetQuery("q")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is a required query param"})
		return
	}
	q = strings.ToLower(q)
	userId, ok := c.GetQuery("user_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is a required query param"})
		return
	}
	urls, ok := cache[userId]
	if !ok {
		fmt.Printf("Cache miss. Populating %s.\n", userId)
		// get the client
		ctx := context.Background()
		client := createClient(ctx)
		defer client.Close()

		urls = make([]FullUrl, 0)
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
			urls = append(urls, FullUrl{
				Id:  data["id"].(string),
				CreatedAt: data["created_at"].(time.Time),
				Url: data["url"].(string),
				Title: data["title"].(string),
				Image:  data["image"].(string),
				UserId:  data["user_id"].(string),
				Content:  data["content"].(string),
			})
		}
		cache[userId] = urls
	}
	fmt.Printf("Searching %d urls for query '%s'.\n", len(urls), q)

	results := make([]SearchResult, 0)
	for _, fullUrl := range urls {
		lowerContent := strings.ToLower(fullUrl.Content)
		lowerTitle := strings.ToLower(fullUrl.Title)
		titleOccurrences := strings.Count(lowerTitle, q)
		contentOccurrences := strings.Count(lowerContent, q)
		exampleText := ""
		if contentOccurrences > 0 {
			startIndex := strings.Index(lowerContent, q)
			startExampleIndex := startIndex - ExamplePadding
			if startExampleIndex < 0 {
				startExampleIndex = 0
			}
			endExampleIndex := startIndex + len(q) + ExamplePadding
			if endExampleIndex > len(fullUrl.Content) {
				endExampleIndex = len(fullUrl.Content)
			}
			exampleText = fullUrl.Content[startExampleIndex:endExampleIndex]
		} else if titleOccurrences > 0 {
			exampleText = fullUrl.Title
		} else {
			continue
		}
		results = append(results, SearchResult{
			Url: fullUrl.Url,
			Title: fullUrl.Title,
			Image: fullUrl.Image,
			Occurrences: titleOccurrences + contentOccurrences,
			ExampleText: exampleText,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Occurrences > results[j].Occurrences
	})
	c.JSON(http.StatusOK, results)
	fmt.Printf("Found %d search results.\n", len(results))
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

	// get the user_id
	doc, err := client.Collection("urls").Doc(id).Get(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	data := doc.Data()
	userId := data["user_id"].(string)

	// delete the document
	_, err = client.Collection("urls").Doc(id).Delete(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	// invalidate the cache
	delete(cache, userId)
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
			Url: data["url"].(string),
			CreatedAt: data["created_at"].(time.Time),
			Title: data["title"].(string),
			Image:  data["image"].(string),
			Id:  data["id"].(string),
		})
	}
	sort.Slice(urls, func(i, j int) bool {
		return urls[i].CreatedAt.After(urls[j].CreatedAt)
	})
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

	// scrape the website
	result, err := scrape(postUrlBody.Url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// save the content
	id := uuid.New().String()
	_, err = client.Collection("urls").Doc(id).Set(ctx, map[string]interface{}{
		"id":      id,
		"created_at": time.Now(),
		"user_id": postUrlBody.UserId,
		"url":     postUrlBody.Url,
		"title":   result.Title,
		"image":   result.Image,
		"content": result.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// invalidate the cache
	delete(cache, postUrlBody.UserId)
}
