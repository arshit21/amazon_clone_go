package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getIndividualProduct(c *gin.Context, db *sql.DB) {
	product_id := c.Param("id")
	product_details := db.QueryRow("Select title, brand, price, description, image, category, units FROM product WHERE id = $1", product_id)
	var product Product

	err := product_details.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product does not exist"})
		return
	}
	c.IndentedJSON(http.StatusOK, product)
}

func getAllProducts(c *gin.Context, db *sql.DB) {
	rows, err := db.Query("SELECT title, brand, price, description, image, category, units FROM product")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err = rows.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, product)
	}
	c.IndentedJSON(http.StatusOK, products)
}
