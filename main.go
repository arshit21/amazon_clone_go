package main

import (
	"database/sql"
	"fmt"

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
	// Pass db as a parameter to getUsers
	router.GET("/customers", func(c *gin.Context) {
		getCustomers(c, db)
	})
	router.POST("/customers/register", func(c *gin.Context) {
		createCustomer(c, db)
	})
	router.POST("/vendors/register", func(c *gin.Context) {
		createVendors(c, db)
	})
	router.POST("/users/login", func(c *gin.Context) {
		loginHandler(c, db)
	})

	router.Run("localhost:8080")

}
