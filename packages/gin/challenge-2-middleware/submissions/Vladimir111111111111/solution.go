package solution

import (
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// Article represents a blog article
type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// In-memory storage
var articles = []Article{
	{ID: 1, Title: "Getting Started with Go", Content: "Go is a programming language...", Author: "John Doe", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: 2, Title: "Web Development with Gin", Content: "Gin is a web framework...", Author: "Jane Smith", CreatedAt: time.Now(), UpdatedAt: time.Now()},
}
var nextID = 3

var totalRequests int
var startTime time.Time

func main() {
	// TODO: Create Gin router without default middleware
	// Use gin.New() instead of gin.Default()
	router := gin.New()
	// TODO: Setup custom middleware in correct order
	// 1. ErrorHandlerMiddleware (first to catch panics)
	router.Use(ErrorHandlerMiddleware())
	// 2. RequestIDMiddleware
	router.Use(RequestIDMiddleware())
	// 3. LoggingMiddleware
	router.Use(LoggingMiddleware())
	// 4. CORSMiddleware
	router.Use(CORSMiddleware())
	// 5. RateLimitMiddleware
	router.Use(RateLimitMiddleware())
	// 6. ContentTypeMiddleware
	router.Use(ContentTypeMiddleware())

	// TODO: Setup route groups
	// Public routes (no authentication required)
	public := router.Group("/")
	// Protected routes (require authentication)
	protected := router.Group("/")
	userOnly := protected.Group("/")
	adminOnly := protected.Group("/")
	// TODO: Define routes
	// Public: GET /ping, GET /articles, GET /articles/:id
	public.GET("/ping", ping)
	public.GET("/articles", getArticles)
	public.GET("/articles/:id", getArticle)
	// Protected: POST /articles, PUT /articles/:id, DELETE /articles/:id, GET /admin/stats
	protected.Use(AuthMiddleware())
	userOnly.Use(AuthMiddleware(), UserAuthorizationMiddleware())
	adminOnly.Use(AuthMiddleware(), AdminAuthorizationMiddleware())
	// Method
	userOnly.GET("/admin/stats", getStats)
	adminOnly.POST("/articles", createArticle)
	adminOnly.PUT("/articles/:id", updateArticle)
	adminOnly.DELETE("/articles/:id", deleteArticle)
	// TODO: Start server on port 8080
	startTime = time.Now()
	router.Run("localhost:8080")
}

// TODO: Implement middleware functions

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Generate UUID for request ID
		// Use github.com/google/uuid package
		request_id := uuid.New().String()
		// Store in context as "request_id"
		c.Set("request_id", request_id)
		// Add to response header as "X-Request-ID"
		c.Writer.Header().Set("X-Request-ID", request_id)

		c.Next()
	}
}

// LoggingMiddleware logs all requests with timing information
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Capture start time
		start := time.Now()
		totalRequests++

		c.Next()
		// TODO: Calculate duration and log request
		duration := time.Since(start)
		path := c.Request.URL.Path

		log.Printf("[%s] %s %s %d %v %s",
			c.GetString("request_id"),
			c.Request.Method,
			path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
		)
		// Format: [REQUEST_ID] METHOD PATH STATUS DURATION IP USER_AGENT
	}
}

// AuthMiddleware validates API keys for protected routes
func AuthMiddleware() gin.HandlerFunc {
	// TODO: Define valid API keys and their roles
	// "admin-key-123" -> "admin"
	// "user-key-456" -> "user"
	roles := map[string]string{
		"admin-key-123": "admin",
		"user-key-456":  "user",
	}

	return func(c *gin.Context) {
		// TODO: Get API key from X-API-Key header
		apiKey := c.GetHeader("X-API-Key")
		// TODO: Validate API key
		if apiKey != "admin-key-123" && apiKey != "user-key-456" {
			apiResponse := APIResponse{
				Error:     "Invalid apiKey",
				Data:      apiKey,
				RequestID: c.GetString("request_id"),
			}
			c.JSON(401, apiResponse)
			c.Abort()
			return
		}
		role := roles[apiKey]
		// TODO: Set user role in context
		c.Set("role", role)
		// TODO: Return 401 if invalid or missing
		c.Next()
	}
}
func AdminAuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") == "user" {
			apiResponse := APIResponse{Error: "Access denied"}
			c.JSON(401, apiResponse)
			c.Abort()
			return
		}
	}
}
func UserAuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") == "admin" {
			apiResponse := APIResponse{Error: "Access denied"}
			c.JSON(401, apiResponse)
			c.Abort()
			return
		}
	}
}

// CORSMiddleware handles cross-origin requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Set CORS headers
		origin := c.GetHeader("Origin")
		// Allow origins: http://localhost:3000, https://myblog.com
		allowedOrigins := map[string]bool{
			"http://localhost:3000": true,
			"https://myblog.com":    true,
		}
		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Origin", origin)
		// Allow methods: GET, POST, PUT, DELETE, OPTIONS
		// Allow headers: Content-Type, X-API-Key, X-Request-ID
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		// TODO: Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.Status(204)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting per IP
type RateLimiter struct {
	visitors    map[string]*visitor
	mu          sync.RWMutex
	requests    int
	duration    time.Duration
	nextCleanup time.Time
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(requests int, duration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors:    make(map[string]*visitor),
		requests:    requests,
		duration:    duration,
		nextCleanup: time.Now().Add(time.Minute),
	}

	// Clean up old visitors every minute
	go rl.cleanupVisitors()

	return rl
}
func (rl *RateLimiter) cleanupVisitors() {
	timer := time.NewTimer(time.Minute)
	defer timer.Stop()

	for {
		<-timer.C
		log.Print("Visitors cleanup")
		rl.mu.Lock()
		rl.visitors = make(map[string]*visitor)
		rl.mu.Unlock()
		timer.Reset(time.Minute)
		rl.nextCleanup = time.Now().Add(time.Minute)
	}
}
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Every(rl.duration), rl.requests)
		rl.visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

var rl *RateLimiter
var once sync.Once

func RateLimitMiddleware() gin.HandlerFunc {
	// TODO: Implement rate limiting
	// Limit: 100 requests per IP per minute
	// Use golang.org/x/time/rate package
	// Set headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
	// Return 429 if rate limit exceeded
	rl := NewRateLimiter(100, time.Minute)
	return func(c *gin.Context) {
		limiter := rl.getVisitor(c.ClientIP())
		c.Header("X-RateLimit-Limit", "100")
		c.Header("X-RateLimit-Reset", strconv.FormatInt(rl.nextCleanup.Unix(), 10))
		remaining := int(math.Round(limiter.Tokens())) - 1
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if !limiter.Allow() {
			c.JSON(429, APIResponse{Error: "Rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ContentTypeMiddleware validates content type for POST/PUT requests
func ContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Check content type for POST/PUT requests
		// Must be application/json
		// Return 415 if invalid content type
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				apiResponse := APIResponse{
					Error: "Content-Type must be application/json",
				}
				c.JSON(415, apiResponse)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// ErrorHandlerMiddleware handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// TODO: Handle panics gracefully
		// Return consistent error response format
		// Include request ID in response
		var errMsg string
		switch v := recovered.(type) {
		case error:
			errMsg = v.Error()
		case string:
			errMsg = v
		default:
			errMsg = fmt.Sprintf("%v", v)
		}

		apiResponse := APIResponse{
			Success:   false,
			Error:     "Internal server error",
			Message:   errMsg,
			RequestID: c.GetString("request_id"),
		}
		c.JSON(500, apiResponse)
		c.Abort()
	})
}

// TODO: Implement route handlers

// ping handles GET /ping - health check endpoint
func ping(c *gin.Context) {
	// TODO: Return simple pong response with request ID
	apiResponse := APIResponse{
		Success:   true,
		RequestID: c.GetString("request_id"),
		Message:   "pong",
	}
	c.JSON(200, apiResponse)
}

// getArticles handles GET /articles - get all articles with pagination
func getArticles(c *gin.Context) {
	// TODO: Implement pagination (optional)
	// TODO: Return articles in standard format
	apiResponse := APIResponse{
		Success:   true,
		Data:      articles,
		Message:   "Here are all the articles",
		RequestID: c.GetString("request_id"),
	}
	c.JSON(200, apiResponse)
}

// getArticle handles GET /articles/:id - get article by ID
func getArticle(c *gin.Context) {
	// TODO: Get article ID from URL paramete

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apiResponse := APIResponse{Error: "invalid id"}
		c.JSON(400, apiResponse)
		return
	}
	// TODO: Find article by ID
	for _, value := range articles {
		if value.ID == id {
			apiResponse := APIResponse{
				Success: true,
				Message: "Here is this article",
				Data:    value,
			}
			c.JSON(200, apiResponse)
			return
		}
	}
	// TODO: Return 404 if not found
	apiResponse := APIResponse{Error: "Article not found"}
	c.JSON(404, apiResponse)
}

// createArticle handles POST /articles - create new article (protected)
func createArticle(c *gin.Context) {
	// TODO: Parse JSON request body
	var article Article
	err := c.BindJSON(&article)
	if err != nil {
		apiResponse := APIResponse{
			Error: "Invalid JSON body",
			Data:  article,
		}
		c.JSON(400, apiResponse)
		c.Abort()
		return
	}
	article.ID = nextID
	nextID++
	// TODO: Validate required fields
	err = validateArticle(article)
	if err != nil {
		apiResponse := APIResponse{
			Error:   err.Error(),
			Message: c.GetString("role"),
			Data:    article,
		}
		c.JSON(400, apiResponse)
		return
	}
	// TODO: Add article to storage
	articles = append(articles, article)
	// TODO: Return created article
	apiResponse := APIResponse{
		Success: true,
		Data:    article,
		Message: "Article created",
	}
	c.JSON(201, apiResponse)
}

// updateArticle handles PUT /articles/:id - update article (protected)
func updateArticle(c *gin.Context) {
	// TODO: Get article ID from URL parameter
	var apiResponse APIResponse

	if c.GetString("role") == "user" {
		apiResponse := APIResponse{Error: "Access denied"}
		c.JSON(403, apiResponse)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apiResponse := APIResponse{Error: "Invalid ID"}
		c.JSON(400, apiResponse)
		return
	}
	articlePtr, idx := findArticleByID(id)
	if articlePtr == nil || idx < 0 {
		apiResponse := APIResponse{Error: "Article not found"}
		c.JSON(404, apiResponse)
		return
	}
	// TODO: Parse JSON request body
	err = validateArticle(*articlePtr)
	if err != nil {
		apiResponse := APIResponse{
			Error: err.Error(),
			Data:  articlePtr,
		}
		c.JSON(422, apiResponse)
		return
	}
	var updatedArtice Article
	err = c.BindJSON(&updatedArtice)
	if err != nil {
		apiResponse := APIResponse{
			Error: "Invalid JSON body",
			Data:  updatedArtice,
		}
		c.JSON(400, apiResponse)
		return
	}
	// TODO: Find and update article
	updatedArtice.ID = id
	*articlePtr = updatedArtice
	// TODO: Return updated article
	apiResponse = APIResponse{
		Message: "Article found and updated",
		Data:    articlePtr,
		Success: true,
	}
	c.JSON(200, apiResponse)
}

// deleteArticle handles DELETE /articles/:id - delete article (protected)
func deleteArticle(c *gin.Context) {
	// TODO: Get article ID from URL parameter
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apiRespone := APIResponse{Error: "Invalid ID"}
		c.JSON(422, apiRespone)
		return
	}
	// TODO: Find and remove article
	_, idx := findArticleByID(id)
	if idx == -1 {
		apiResponse := APIResponse{
			Error: "Article not found",
			Data:  idx,
		}
		c.JSON(404, apiResponse)
	}
	articles = slices.Delete(articles, idx, idx+1)
	// TODO: Return success message
	apiResponse := APIResponse{
		Message: "Article deleted",
		Data:    articles,
		Success: true,
	}
	c.JSON(200, apiResponse)
}

// getStats handles GET /admin/stats - get API usage statistics (admin only)
func getStats(c *gin.Context) {
	// TODO: Check if user role is "admin"
	// TODO: Return mock statistics
	uptime := time.Since(startTime)

	stats := map[string]interface{}{
		"total_articles": len(articles),
		"total_requests": totalRequests, // Could track this in middleware
		"uptime":         uptime.String(),
	}

	// TODO: Return stats in standard format
	apiResponse := APIResponse{
		Success: true,
		Data:    stats,
		Message: "Statistics",
	}
	c.JSON(200, apiResponse)
}

// Helper functions

// findArticleByID finds an article by ID
func findArticleByID(id int) (*Article, int) {
	// TODO: Implement article lookup
	for i, value := range articles {
		if value.ID == id {
			return &articles[i], i
		}
	}
	// Return article pointer and index, or nil and -1 if not found
	return nil, -1
}

var titleRegEx = regexp.MustCompile(`^[\p{L}\s']+$`)
var contentRegEx = regexp.MustCompile(`^[\p{L}\s\.']+$`)
var authorRegEx = regexp.MustCompile(`^[\p{L}']+[\s][\p{L}']+$`)

// validateArticle validates article data
func validateArticle(article Article) error {
	// TODO: Implement validation
	// Check required fields: Title, Content, Author

	if !titleRegEx.MatchString(article.Title) {
		return errors.New("Title validation failed")
	}
	if !contentRegEx.MatchString(article.Content) {
		return errors.New("Content validation failed")
	}
	if !authorRegEx.MatchString(article.Author) {
		return errors.New("Author validation failed")
	}

	return nil
}
