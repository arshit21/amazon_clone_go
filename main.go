package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

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
	store := cookie.NewStore([]byte("c1d545ff2d75aae60f492d3e06ec0a10"))
	router.Use(sessions.Sessions("mysession", store))
	router.GET("/users/my_profile", authMiddleware(), func(c *gin.Context) {
		getCurrentUser(c, db)
	})
	router.PATCH("users/my_profile", authMiddleware(), func(c *gin.Context) {
		UpdateMyDetails(c, db)
	})
	router.POST("/customers/register", func(c *gin.Context) {
		createCustomer(c, db)
	})
	router.POST("/vendors/register", func(c *gin.Context) {
		createVendors(c, db)
	})
	router.POST("/login", func(c *gin.Context) {
		loginHandler(c, db)
	})
	router.POST("/logout", handleLogout)
	router.POST("/vendors/add_product", authMiddleware(), func(c *gin.Context) {
		addProduct(c, db)
	})
	router.Run("localhost:8080")

}
