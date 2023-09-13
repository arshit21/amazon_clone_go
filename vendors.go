package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Product struct {
	Title       string `json:"title"`
	Brand       string `json:"brand"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Category    string `json:"category"`
	Units       int    `json:"units"`
}

func addProduct(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	session := sessions.Default(c)
	username := session.Get("username")

	details := db.QueryRow("SELECT first_name, last_name, username, email, is_customer, is_vendor from users WHERE username = $1", username)

	var user User
	err := details.Scan(&user.First_name, &user.Last_name, &user.Username, &user.Email, &user.Is_customer, &user.Is_vendor)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !user.Is_vendor {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "You are not a vendor"})
		return
	}

	var newProduct Product
	if err := c.BindJSON(&newProduct); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	var vendor_id int
	row := db.QueryRow("SELECT id FROM vendors WHERE username = $1", username)
	err_ := row.Scan(&vendor_id)

	if err_ != nil {
		// Handle the error, for example:
		fmt.Println("Error:", err)
		return
	}

	stmt, err := db.Prepare("INSERT INTO product (title, brand, price, description, category, units, vendor_id, image) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(newProduct.Title, newProduct.Brand, newProduct.Price, newProduct.Description, newProduct.Category, newProduct.Units, vendor_id, newProduct.Image); err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusCreated, newProduct)
}
