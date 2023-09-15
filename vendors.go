package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Product struct for storing the product information
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

	//get current user details using gin sessions
	session := sessions.Default(c)
	username := session.Get("username")

	//get the user details using username stored in the session
	details := db.QueryRow("SELECT first_name, last_name, username, email, is_customer, is_vendor from users WHERE username = $1", username)

	//create new user variable of type User and scan the current vendor details in the user struct
	var user User
	err := details.Scan(&user.First_name, &user.Last_name, &user.Username, &user.Email, &user.Is_customer, &user.Is_vendor)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//check if the current user is a vendor
	if !user.Is_vendor {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "You are not a vendor"})
		return
	}

	//create newProdut of type Product and bind the request json data to that newProduct
	var newProduct Product
	if err := c.BindJSON(&newProduct); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	//get vendor_id of the current vendor
	var vendor_id int
	row := db.QueryRow("SELECT id FROM vendors WHERE username = $1", username)
	err_ := row.Scan(&vendor_id)
	if err_ != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	//prepare a SQL statement for inserting data into product table and handling any potential errors
	stmt, err := db.Prepare("INSERT INTO product (title, brand, price, description, category, units, vendor_id, image) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	//Execute the SQL statement to populate the table with the appropriate data
	if _, err := stmt.Exec(newProduct.Title, newProduct.Brand, newProduct.Price, newProduct.Description, newProduct.Category, newProduct.Units, vendor_id, newProduct.Image); err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusCreated, newProduct)
}

func getMyProducts(c *gin.Context, db *sql.DB) {
	//get username for the session and check if the user exists and is the user a vendor
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

	//get the vendor id of the current vendor
	var vendor_id int
	row := db.QueryRow("SELECT id FROM vendors WHERE username = $1", username)
	err_ := row.Scan(&vendor_id)
	if err_ != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Query products with specific vendor ID from the database.
	rows, err := db.Query("SELECT title, brand, price, description, image, category, units FROM product WHERE vendor_id = $1", vendor_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Close the rows when done to prevent memory leaks.
	defer rows.Close()

	// Create a slice to hold product data.
	var products []Product
	for rows.Next() {
		var newProduct Product
		// Scan the retrieved row's columns into the 'newProduct' struct fields.
		err := rows.Scan(&newProduct.Title, &newProduct.Brand, &newProduct.Price, &newProduct.Description, &newProduct.Image, &newProduct.Category, &newProduct.Units)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Append the product data to the 'products' slice.
		products = append(products, newProduct)
	}
	c.JSON(http.StatusOK, products)
}

func updateProductDetails(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	session := sessions.Default(c)
	username := session.Get("username")

	// Query the database for the vendor ID associated with the username.
	details := db.QueryRow("SELECT id from vendors WHERE username = $1", username)

	var vendor_id int
	err := details.Scan(&vendor_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "You are not a vendor"})
		return
	}
	// Retrieve the product ID from the URL
	product_id := c.Param("id")

	// Query the database to get the vendor ID associated with the product.
	prod_details := db.QueryRow("Select vendor_id FROM product WHERE id = $1", product_id)

	//Check if the product's vendor ID matches the vendor's ID.
	var product_vendor_id int
	err_ := prod_details.Scan(&product_vendor_id)
	if err_ != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product does not exist"})
		return
	}

	if product_vendor_id != vendor_id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you can only edit your own products"})
		return
	}

	// Query the database for the product details.
	product_details := db.QueryRow("Select title, brand, price, description, image, category, units FROM product WHERE id = $1", product_id)

	var product Product
	_err_ := product_details.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)
	if _err_ != nil {
		log.Fatal(_err_)
	}

	// Bind the updated product details from the request JSON.
	var updatedProduct Product
	if err := c.BindJSON(&updatedProduct); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Apply default values if fields are empty in the update request.
	if len(updatedProduct.Title) == 0 {
		updatedProduct.Title = product.Title
	}

	if len(updatedProduct.Brand) == 0 {
		updatedProduct.Brand = product.Brand
	}

	if updatedProduct.Price == 0 {
		updatedProduct.Price = product.Price
	}

	if len(updatedProduct.Description) == 0 {
		updatedProduct.Description = product.Description
	}

	if len(updatedProduct.Image) == 0 {
		updatedProduct.Image = product.Image
	}

	if len(updatedProduct.Category) == 0 {
		updatedProduct.Category = product.Category
	}

	if updatedProduct.Units == 0 {
		updatedProduct.Units = product.Units
	}

	// Update the product details in the database
	_, err__ := db.Exec("UPDATE product SET title = $1, brand = $2, price = $3, description = $4, image = $5, category = $6, units = $7 WHERE id = $8", updatedProduct.Title,
		updatedProduct.Brand, updatedProduct.Price, updatedProduct.Description, updatedProduct.Image, updatedProduct.Category, updatedProduct.Units, product_id)
	if err__ != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err_.Error()})
		return
	}
	// Respond with a success message and the updated product details.
	c.JSON(http.StatusOK, gin.H{"message": "Product details updated succesfully", "Updated Details": updatedProduct})

}

func ordersForMe(c *gin.Context, db *sql.DB) {

	session := sessions.Default(c)
	username := session.Get("username")

	// Query the database to get the vendor ID associated with the username.
	var vendorId int
	err := db.QueryRow("SELECT id from vendors where username=$1", username).Scan(&vendorId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a vendor"})
		return
	}

	rows, err := db.Query("SELECT product_id, money_paid, units, date_created FROM orders WHERE vendor_id = $1", vendorId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var orders []orderDetails
	for rows.Next() {
		var order_1 orderDetails_2
		err := rows.Scan(&order_1.Product, &order_1.MoneyPaid, &order_1.Units, &order_1.OrderDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Query the database for the product title associated with the order.
		product_title := db.QueryRow("SELECT title FROM product WHERE id = $1", order_1.Product)
		var title string
		err = product_title.Scan(&title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var final_order orderDetails
		final_order.Product = title
		final_order.MoneyPaid = order_1.MoneyPaid
		final_order.Units = order_1.Units
		final_order.OrderDate = order_1.OrderDate
		orders = append(orders, final_order)
	}
	// Respond with a JSON containing the orders associated with the vendor.
	c.IndentedJSON(http.StatusOK, gin.H{"Here are all your Orders": orders})
}
