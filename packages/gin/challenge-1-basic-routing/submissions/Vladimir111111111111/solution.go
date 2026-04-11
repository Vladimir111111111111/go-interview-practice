package solution

import (
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// User represents a user in our system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// In-memory storage
var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 30},
	{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 25},
	{ID: 3, Name: "Bob Wilson", Email: "bob@example.com", Age: 35},
}
var nextID = 4

func main() {
	// TODO: Create Gin router
	router := gin.Default()
	// TODO: Setup routes
	// GET /users - Get all users
	router.GET("/users", getAllUsers)
	// GET /users/:id - Get user by ID
	router.GET("/users/:id", getUserByID)
	// POST /users - Create new user
	router.POST("/users", createUser)
	// PUT /users/:id - Update user
	router.PUT("/users/:id", updateUser)
	// DELETE /users/:id - Delete user
	router.DELETE("/users/:id", deleteUser)
	// GET /users/search - Search users by name
	router.GET("/users/search", searchUsers)

	// TODO: Start server on port 8080
	router.Run("localhost:8080")
}

// TODO: Implement handler functions

// getAllUsers handles GET /users
func getAllUsers(c *gin.Context) {
	// TODO: Return all users
	var response Response
	response.Success = true
	response.Message = "Here are all users"
	response.Data = users
	response.Code = 200
	c.JSON(http.StatusOK, response)
}

// getUserByID handles GET /users/:id
func getUserByID(c *gin.Context) {
	// TODO: Get user by ID
	var response Response

	idStr := c.Param("id")
	// Handle invalid ID format
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error = "Invalid ID"
		response.Code = 400
		c.IndentedJSON(http.StatusBadRequest, response)
		return
	}
	// Return 404 if user not found
	for _, a := range users {
		if a.ID == id {
			response.Success = true
			response.Data = a
			response.Message = "User found"
			response.Code = 200
			c.IndentedJSON(http.StatusOK, response)
			return
		}
	}
	response.Error = "User not found"
	response.Code = 404
	c.IndentedJSON(http.StatusNotFound, response)
}

// createUser handles POST /users
func createUser(c *gin.Context) {
	// TODO: Parse JSON request body
	var response Response

	var newUser User
	err := c.BindJSON(&newUser)
	if err != nil {
		response.Error = "Unvalid request body"
		c.IndentedJSON(http.StatusBadRequest, response)
		return
	}

	err = validateUser(newUser)
	if err != nil {
		response.Error = err.Error()
		c.IndentedJSON(http.StatusBadRequest, response)
		return
	}
	for _, value := range users {
		if value.Name == newUser.Name || value.Email == newUser.Email {
			response.Error = "User with this name or email already exists"
			c.IndentedJSON(http.StatusBadRequest, response)
			return
		}
	}
	// Validate required fields idk
	newUser.ID = nextID
	nextID += 1
	// Add user to storage
	users = append(users, newUser)
	// Return created user
	response.Success = true
	response.Message = "User created"
	response.Data = newUser
	response.Code = 201
	c.IndentedJSON(http.StatusCreated, response)
}

// updateUser handles PUT /users/:id
func updateUser(c *gin.Context) {
	// TODO: Get user ID from path
	var response Response

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error = "Failed to parse the id parameter"
		response.Code = 400
		c.IndentedJSON(http.StatusBadRequest, response)
		return
	}
	// Find and update user
	userPtr, index := findUserByID(id)
	if userPtr == nil || index == -1 {
		response.Error = "User not found"
		response.Code = 404
		c.IndentedJSON(http.StatusNotFound, response)
		return
	}
	// Parse JSON request body
	var newUser User
	err = c.BindJSON(&newUser)
	if err != nil {
		response.Error = "Invalid response body"
		response.Code = 400
		c.IndentedJSON(http.StatusBadRequest, response)
		return
	}
	err = validateUser(newUser)
	if err != nil {
		response.Error = err.Error()
		c.IndentedJSON(http.StatusBadRequest, response)
		return
	}
	newUser.ID = id
	*userPtr = newUser
	// Return updated user
	response.Success = true
	response.Data = *userPtr
	response.Message = "User data updated"
	c.IndentedJSON(http.StatusOK, response)
}

// deleteUser handles DELETE /users/:id
func deleteUser(c *gin.Context) {
	// TODO: Get user ID from path
	var response Response

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error = "Failed to parse the request"
		response.Code = 400
		c.IndentedJSON(http.StatusBadRequest, response)
		return
	}
	// Find and remove user
	_, idx := findUserByID(id)
	if idx == -1 {
		response.Error = "User not found"
		c.IndentedJSON(http.StatusNotFound, response)
		return
	}
	users = slices.Delete(users, idx, idx+1)
	// Return success message
	response.Success = true
	response.Message = "User deleted"
	response.Code = 200
	c.IndentedJSON(http.StatusOK, response)
}

// searchUsers handles GET /users/search?name=value
func searchUsers(c *gin.Context) {
	// TODO: Get name query parameter
	var response Response
	name := c.Query("name")
	if name == "" {
		response.Error = "No parameter"
		c.IndentedJSON(http.StatusBadRequest, response)
		return
	}
	var foundUsers = []User{}
	// Filter users by name (case-insensitive)
	for _, user := range users {
		result := strings.Split(user.Name, " ")
		if strings.EqualFold(name, result[0]) {
			foundUsers = append(foundUsers, user)
		}
	}
	if len(foundUsers) == 0 {
		response.Message = "No users found"
		response.Success = true
		response.Data = foundUsers
		c.IndentedJSON(http.StatusOK, response)
		return
	}
	// Return matching users
	response.Success = true
	response.Data = foundUsers
	c.IndentedJSON(http.StatusOK, response)
}

// Helper function to find user by ID
func findUserByID(id int) (*User, int) {
	// TODO: Implement user lookup
	for i := range users {
		if users[i].ID == id {
			return &users[i], i
		}
	}
	// Return user pointer and index, or nil and -1 if not found
	return nil, -1
}

// Helper function to validate user data
var nameRegex = regexp.MustCompile(`^[\p{L}']+[\s][\p{L}']+$`)
var emailRegex = regexp.MustCompile(`^[\p{L}_\d\.]+@[\p{L}]+\.[\p{L}]{2,}$`)

func validateUser(user User) error {
	// TODO: Implement validation
	// Check required fields: Name, Email
	// Validate email format (basic check)
	if !nameRegex.MatchString(user.Name) {
		return errors.New("Invalid Name")
	}
	if !emailRegex.MatchString(user.Email) {
		return errors.New("Invalid Email")
	}

	return nil
}
