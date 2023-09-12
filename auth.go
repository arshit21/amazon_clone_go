package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	// ID         int    `json:"id"`
	First_name  string `json:"first_name"`
	Last_name   string `json:"last_name"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Is_customer bool   `json:"is_customer"`
	Is_vendor   bool   `json:"is_vendor"`
}

func getCurrentUser(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	session := sessions.Default(c)
	username := session.Get("username")

	details := db.QueryRow("SELECT first_name, last_name, username, email, is_customer, is_vendor from users WHERE username = $1", username)

	var user User
	err := details.Scan(&user.First_name, &user.Last_name, &user.Username, &user.Email, &user.Is_customer, &user.Is_vendor)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, user)
}

func UpdateMyDetails(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	session := sessions.Default(c)
	username := session.Get("username")

	details := db.QueryRow("SELECT first_name, last_name, username, email, is_customer, is_vendor from users WHERE username = $1", username)

	var user User
	error := details.Scan(&user.First_name, &user.Last_name, &user.Username, &user.Email, &user.Is_customer, &user.Is_vendor)
	if error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	var updatedUser User
	if err := c.BindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if len(updatedUser.Username) == 0 {
		updatedUser.Username = user.Username
	}
	if len(updatedUser.First_name) == 0 {
		updatedUser.First_name = user.First_name
	}
	if len(updatedUser.Last_name) == 0 {
		updatedUser.Last_name = user.Last_name
	}
	if len(updatedUser.Email) == 0 {
		updatedUser.Email = user.Email
	}

	_, err := db.Exec("UPDATE users SET first_name = $1, last_name = $2, username = $3, email = $4 WHERE username = $5", updatedUser.First_name, updatedUser.Last_name, updatedUser.Username, updatedUser.Email, username)

	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User details updated successfully"})
}

func createCustomer(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	var newuser User
	if err := c.BindJSON(&newuser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newuser.Password), bcrypt.DefaultCost)
	newuser.Is_customer = true
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := db.Prepare("INSERT INTO customers (first_name, last_name, username, email, password) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(newuser.First_name, newuser.Last_name, newuser.Username, newuser.Email, hashedPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
		return
	}
	stmt_2, err := db.Prepare("INSERT INTO users (first_name, last_name, username, email, password, is_customer) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt_2.Close()

	if _, err := stmt_2.Exec(newuser.First_name, newuser.Last_name, newuser.Username, newuser.Email, hashedPassword, newuser.Is_customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Here are your details", "details": user})
}

func createVendors(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	var newuser User
	if err := c.BindJSON(&newuser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newuser.Password), bcrypt.DefaultCost)
	newuser.Is_vendor = true
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := db.Prepare("INSERT INTO vendors (first_name, last_name, username, email, password) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(newuser.First_name, newuser.Last_name, newuser.Username, newuser.Email, hashedPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
		return
	}
	stmt_2, err := db.Prepare("INSERT INTO users (first_name, last_name, username, email, password, is_vendor) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt_2.Close()

	if _, err := stmt_2.Exec(newuser.First_name, newuser.Last_name, newuser.Username, newuser.Email, hashedPassword, newuser.Is_vendor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
		return
	}

	c.IndentedJSON(http.StatusCreated, newuser)
}

func loginHandler(c *gin.Context, db *sql.DB) {
	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&loginRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var user User
	err := db.QueryRow("SELECT * FROM users WHERE username = $1", loginRequest.Username).
		Scan(&user.First_name, &user.Last_name, &user.Username, &user.Email, &user.Password, &user.Is_customer, &user.Is_vendor)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("authenticated", true)
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": user})

}

func handleLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		auth := session.Get("authenticated")
		if auth != nil && auth.(bool) {
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
		}
	}
}
