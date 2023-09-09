package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID         int    `json:"id"`
	First_name string `json:"first_name"`
	Last_name  string `json:"last_name"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

// sets the constants for connecting to our database
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "testing321"
	dbname   = "amazon_clone"
)

func main() {
	// string containing all the information required to connect to our database
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// The sql.Open() function takes two arguments - a driver name, and a string that tells that driver how to connect to our database
	// and then returns a pointer to a sql.DB and an error.
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	// Pass db as a parameter to getUsers
	router.GET("/users", func(c *gin.Context) {
		getUsers(c, db)
	})
	router.POST("/users/register", func(c *gin.Context) {
		createUser(c, db)
	})
	router.POST("/users/login", func(c *gin.Context) {
		loginHandler(c, db)
	})

	router.Run("localhost:8080")

}

// Pass db as a parameter to getUsers
func getUsers(c *gin.Context, db *sql.DB) {

	c.Header("Content-Type", "application/json")
	rows, err := db.Query("SELECT id, first_name, last_name, username, email FROM users")
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var a User
		err := rows.Scan(&a.ID, &a.First_name, &a.Last_name, &a.Username, &a.Email)
		if err != nil {
			log.Fatal(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		users = append(users, a)
	}
	c.IndentedJSON(http.StatusOK, users)
}

func createUser(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	var newuser User
	if err := c.BindJSON(&newuser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newuser.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := db.Prepare("INSERT INTO users (id, first_name, last_name, username, email, password) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(newuser.ID, newuser.First_name, newuser.Last_name, newuser.Username, newuser.Email, hashedPassword); err != nil {
		log.Fatal(err)
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
		Scan(&user.ID, &user.First_name, &user.Last_name, &user.Username, &user.Email, &user.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": user})
}
