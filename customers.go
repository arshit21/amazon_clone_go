package main

import (
	"database/sql"
	"log"
	"math/big"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AddRequest struct {
	ToAdd int `json:"toAdd"`
}

type UnitsRequest struct {
	Units *big.Int `json:"units"`
}

type orderDetails struct {
	Product   string
	Units     int
	MoneyPaid int
	OrderDate time.Time
}

type orderDetails_2 struct {
	Product   int
	MoneyPaid int
	Units     int
	OrderDate time.Time
}

type shoppingCart struct {
	Product string
	Units   int
}

func getIndividualProduct(c *gin.Context, db *sql.DB) {
	// Get the product ID from the URL parameter.
	product_id := c.Param("id")

	// Query the database to get details of the product with the given ID.
	product_details := db.QueryRow("SELECT title, brand, price, description, image, category, units FROM product WHERE id = $1", product_id)

	// Scan the result into the 'product' variable.
	var product Product
	err := product_details.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product does not exist"})
		return
	}

	productDetails := make(map[string]interface{})
	productDetails["title"] = product.Title
	productDetails["brand"] = product.Brand
	productDetails["price"] = product.Price
	productDetails["description"] = product.Description
	productDetails["image"] = "http://" + c.Request.Host + "/images/" + filepath.Base(product.Image)
	productDetails["category"] = product.Category
	productDetails["units"] = product.Units
	// Respond with the product details in JSON format.
	c.IndentedJSON(http.StatusOK, productDetails)
}

func getAllProducts(c *gin.Context, db *sql.DB) {
	// Query the database to get details of all products.
	rows, err := db.Query("SELECT title, brand, price, description, image, category, units FROM product")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		// Scan each row of the result into a 'product' variable.
		var product Product
		err = rows.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Append the product details to the 'products' slice.
		products = append(products, product)

		// Update the image URLs to include the base URL of your server
		for i, product := range products {
			products[i].Image = "http://" + c.Request.Host + "/images/" + filepath.Base(product.Image)
		}
	}
	// Respond with the list of products in JSON format.
	c.IndentedJSON(http.StatusOK, products)
}

func getWalletDetails(c *gin.Context, db *sql.DB) {
	// Get the username from the session.
	session := sessions.Default(c)
	username := session.Get("username")

	// Query the database to get the customer ID associated with the username.
	var customer_id int
	err := db.QueryRow("SELECT id FROM customers WHERE username = $1", username).Scan(&customer_id)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not a customer"})
		return
	}

	// Query the database to get the wallet balance for the customer.
	var wallet_balance int
	err = db.QueryRow("SELECT balance FROM wallet WHERE customer_id = $1", customer_id).Scan(&wallet_balance)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with the wallet balance in JSON format.
	c.IndentedJSON(http.StatusOK, gin.H{"Wallet Balance": wallet_balance})
}

func addMoneytoWallet(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	session := sessions.Default(c)
	username := session.Get("username")

	// Query the database to get the customer ID associated with the username.
	var customer_id int
	err := db.QueryRow("SELECT id FROM customers WHERE username = $1", username).Scan(&customer_id)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not a customer"})
		return
	}

	// Query the database to get the current wallet balance for the customer.
	var wallet_balance int
	err = db.QueryRow("SELECT balance FROM wallet WHERE customer_id = $1", customer_id).Scan(&wallet_balance)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Bind the JSON request payload to the 'addRequest' struct.
	var addRequest AddRequest
	err = c.BindJSON(&addRequest)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate the new balance after adding the requested amount.
	toAdd := addRequest.ToAdd
	newBalance := wallet_balance + toAdd

	// Update the wallet balance in the database.
	_, err = db.Exec("UPDATE wallet SET balance = $1 WHERE customer_id = $2", newBalance, customer_id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with a success message and the updated wallet balance in JSON format.
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Money has been added", "Wallet Balance": newBalance})
}

func buyNow(c *gin.Context, db *sql.DB) {
	// Extract product ID from URL parameter.
	product_id := c.Param("id")

	// Query the database to get product details.
	product_details := db.QueryRow("SELECT title, brand, price, description, image, category, units FROM product WHERE id = $1", product_id)
	var product Product

	// Scan the result into the 'product' struct.
	err := product_details.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product does not exist"})
		return
	}

	// Extract username from session.
	c.Header("Content-Type", "application/json")
	session := sessions.Default(c)
	username := session.Get("username")
	var customerId int

	// Query the database to get customer ID.
	err = db.QueryRow("SELECT id from customers where username=$1", username).Scan(&customerId)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}

	// Bind JSON request payload to 'units' variable.
	var units UnitsRequest
	err = c.BindJSON(&units)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Payload"})
		return
	}
	productUnits := big.NewInt(int64(product.Units))

	// Compare productUnits with units.Units
	if units.Units.Cmp(productUnits) == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Out of Stock"})
		return
	}
	inStockUnits := int(units.Units.Int64())
	//check if wallet has enough balance
	moneyPayable := product.Price * int(inStockUnits)

	var walletBalance int
	err = db.QueryRow("SELECT balance FROM wallet WHERE customer_id = $1", customerId).Scan(&walletBalance)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if moneyPayable > walletBalance {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient Wallet balance"})
		return
	}

	var product_vendor_id int
	err = db.QueryRow("SELECT vendor_id FROM product WHERE id = $1", product_id).Scan(&product_vendor_id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Prepare SQL statement for inserting order details.
	stmt, err := db.Prepare("INSERT into orders (product_id, customer_id, money_paid, units, vendor_id) VALUES ($1, $2, $3, $4, $5)")

	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Execute the SQL statement to insert order details.
	if _, err := stmt.Exec(product_id, customerId, moneyPayable, inStockUnits, product_vendor_id); err != nil {
		log.Fatal(err)
	}

	//update product units and customer wallet
	product.Units = product.Units - int(inStockUnits)
	walletBalance = walletBalance - moneyPayable

	_, err = db.Exec("UPDATE wallet SET balance = $1 WHERE customer_id = $2", walletBalance, customerId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = db.Exec("UPDATE product SET units = $1 WHERE id = $2", product.Units, product_id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//return the order details
	var Order orderDetails
	Order.Product = product.Title
	Order.MoneyPaid = moneyPayable
	Order.Units = inStockUnits

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Product ordered", "Order details": Order})
}

func previousOrders(c *gin.Context, db *sql.DB) {
	// Get username from session.
	session := sessions.Default(c)
	username := session.Get("username")

	var customerId int

	// Query the database to get the customer ID associated with the username.
	err := db.QueryRow("SELECT id from customers where username=$1", username).Scan(&customerId)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}

	// Query the database to get previous orders.
	rows, err := db.Query("SELECT product_id, money_paid, units, date_created FROM orders WHERE customer_id = $1", customerId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var orders []orderDetails

	// Iterate through the rows of the query result.
	for rows.Next() {
		var order_1 orderDetails_2
		err := rows.Scan(&order_1.Product, &order_1.MoneyPaid, &order_1.Units, &order_1.OrderDate)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Query the database to get the product title associated with the product ID.
		product_title := db.QueryRow("SELECT title FROM product WHERE id = $1", order_1.Product)
		var title string
		err = product_title.Scan(&title)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create a final orderDetails struct with product details and add it to the orders slice.
		var final_order orderDetails
		final_order.Product = title
		final_order.MoneyPaid = order_1.MoneyPaid
		final_order.Units = order_1.Units
		final_order.OrderDate = order_1.OrderDate
		orders = append(orders, final_order)
	}

	// Send a JSON response containing the list of orders.
	c.IndentedJSON(http.StatusOK, gin.H{"Here are all your Orders": orders})
}

func addToCart(c *gin.Context, db *sql.DB) {
	// Get the product ID from the request parameters.
	product_id := c.Param("id")

	// Query the database to get product details.
	product_details := db.QueryRow("SELECT title, brand, price, description, image, category, units FROM product WHERE id = $1", product_id)
	var product Product

	// Scan the product details into the product struct.
	err := product_details.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product does not exist"})
		return
	}

	// Set response content type to JSON.
	c.Header("Content-Type", "application/json")

	// Get the username from the session.
	session := sessions.Default(c)
	username := session.Get("username")

	var customerId int

	// Query the database to get the customer ID associated with the username.
	err = db.QueryRow("SELECT id from customers WHERE username=$1", username).Scan(&customerId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}

	var units UnitsRequest

	// Bind the request JSON data to the units variable.
	err = c.BindJSON(&units)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Payload"})
		return
	}
	productUnits := big.NewInt(int64(product.Units))

	// Compare productUnits with units.Units
	if units.Units.Cmp(productUnits) == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Out of Stock"})
		return
	}
	inStockUnits := int(units.Units.Int64())

	var cartId int

	// Query the database to get the customer's cart ID.
	err = db.QueryRow("SELECT id FROM cart WHERE customer_id = $1", customerId).Scan(&cartId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	moneyPayable := product.Price * int(inStockUnits)

	// Prepare an SQL statement to insert the product into the cart.
	stmt, err := db.Prepare("INSERT INTO cart_object (product_id, customer_id, cart_id, units, money) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Execute the SQL statement to insert the product into the cart.
	if _, err := stmt.Exec(product_id, customerId, cartId, inStockUnits, moneyPayable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cartCost int

	// Query the database to get the current cart cost.
	err = db.QueryRow("SELECT cost FROM cart WHERE customer_id = $1", customerId).Scan(&cartCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update the cart cost with the new addition.
	cartCost = cartCost + moneyPayable

	// Update the cart cost in the database.
	_, err = db.Exec("UPDATE cart SET cost = $1 WHERE customer_id = $2", cartCost, customerId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return a JSON response indicating that the product has been added to the cart.
	c.IndentedJSON(http.StatusOK, gin.H{"message": "added to cart"})
}

func viewcart(c *gin.Context, db *sql.DB) {
	// Get the user's session to retrieve their username.
	session := sessions.Default(c)
	username := session.Get("username")

	var customerId int
	var cartId int

	// Query the database to get the customer's ID using their username.
	err := db.QueryRow("SELECT id from customers where username=$1", username).Scan(&customerId)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}

	// Query the database to retrieve the customer's cart ID.
	err = db.QueryRow("SELECT id FROM cart WHERE customer_id =$1", customerId).Scan(&cartId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Query the database to get the list of products and their quantities in the cart.
	rows, err := db.Query("SELECT product_id, units FROM cart_object WHERE cart_id = $1", cartId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Initialize a slice to store the cart items.
	var cart_objects []shoppingCart

	for rows.Next() {
		var product_id int
		var product_units int

		// Scan the current row to extract product ID and units.
		err = rows.Scan(&product_id, &product_units)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var product_title string

		// Query the database to get the product title based on the product ID.
		err = db.QueryRow("SELECT title FROM product WHERE id = $1", product_id).Scan(&product_title)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create a new cart object to represent a product in the cart.
		var cart_object shoppingCart

		// Set the product title and units in the cart object.
		cart_object.Product = product_title
		cart_object.Units = product_units

		// Append the cart object to the list of cart items.
		cart_objects = append(cart_objects, cart_object)
	}

	// Query the database to get the total cost of the items in the cart.
	var cartCost int
	err = db.QueryRow("SELECT cost FROM cart WHERE customer_id=$1", customerId).Scan(&cartCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the cart details in the JSON response.
	c.IndentedJSON(http.StatusOK, gin.H{"Cart total": cartCost, "Items": cart_objects})
}

func buycart(c *gin.Context, db *sql.DB) {
	// Get the user's session to retrieve their username.
	session := sessions.Default(c)
	username := session.Get("username")

	// Initialize variables to store the customer's ID, cart ID, cart cost, and wallet balance.
	var customerId int
	var cartId int
	var cartCost int
	var walletBalance int

	// Query the database to get the customer's ID using their username.
	err := db.QueryRow("SELECT id from customers WHERE username=$1", username).Scan(&customerId)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}

	// Query the database to retrieve the customer's cart ID.
	err = db.QueryRow("SELECT id FROM cart WHERE customer_id =$1 ", customerId).Scan(&cartId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Query the database to get the total cost of the items in the cart.
	err = db.QueryRow("SELECT cost FROM cart WHERE customer_id=$1", customerId).Scan(&cartCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Query the database to get the customer's wallet balance.
	err = db.QueryRow("SELECT balance FROM wallet WHERE customer_id = $1", customerId).Scan(&walletBalance)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the cart cost exceeds the wallet balance.
	if cartCost > walletBalance {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance in wallet"})
		return
	}

	// Initialize a slice to store the details of all ordered products.
	var allOrders []orderDetails

	// Query the database to get the list of products and their quantities in the cart.
	rows, err := db.Query("SELECT product_id, units FROM cart_object WHERE cart_id = $1", cartId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize variables to store product details.
		var product_id int
		var product_units int

		// Scan the current row to extract product ID and units.
		err = rows.Scan(&product_id, &product_units)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Query the database to get the details of the ordered product.
		product_details := db.QueryRow("SELECT title, brand, price, description, image, category, units FROM product WHERE id = $1", product_id)
		var product Product

		err := product_details.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Product does not exist"})
			return
		}

		// Calculate the total cost for the ordered product.
		money := product.Price * product_units

		// Query the database to get the vendor ID of the product.
		var product_vendor_id int
		err = db.QueryRow("SELECT vendor_id FROM product WHERE id = $1", product_id).Scan(&product_vendor_id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Prepare and execute a SQL statement to record the order details.
		stmt, err := db.Prepare("INSERT into orders (product_id, customer_id, money_paid, units, vendor_id) VALUES ($1, $2, $3, $4, $5)")

		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		if _, err := stmt.Exec(product_id, customerId, money, product_units, product_vendor_id); err != nil {
			log.Fatal(err)
		}

		// Update product units, wallet balance, and cart items after the purchase.
		product.Units = product.Units - product_units
		walletBalance = walletBalance - money

		_, err = db.Exec("UPDATE wallet SET balance = $1 WHERE customer_id = $2", walletBalance, customerId)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec("UPDATE product SET units = $1 WHERE id = $2", product.Units, product_id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create an orderDetails object to store details of the ordered product.
		var Order orderDetails
		Order.Product = product.Title
		Order.MoneyPaid = money
		Order.Units = product_units

		// Add the orderDetails object to the list of all orders.
		allOrders = append(allOrders, Order)

		// Remove the purchased items from the cart.
		_, err = db.Exec("DELETE FROM cart_object WHERE cart_id = $1", cartId)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Update the cart cost to 0 after the purchase.
	_, err = db.Exec("UPDATE cart SET cost = $1 WHERE customer_id = $2", 0, customerId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the details of all ordered products in the JSON response.
	c.JSON(http.StatusOK, allOrders)
}

func removeFromCart(c *gin.Context, db *sql.DB) {
	// Get the user's session to retrieve their username.
	session := sessions.Default(c)
	username := session.Get("username")

	// Initialize variables to store the customer's ID, cart ID and cart cos\
	var customerId int
	var cartId int
	var cartCost int

	// Query the database to get the customer's ID using their username.
	err := db.QueryRow("SELECT id from customers WHERE username=$1", username).Scan(&customerId)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}

	// Query the database to retrieve the customer's cart ID.
	err = db.QueryRow("SELECT id FROM cart WHERE customer_id =$1 ", customerId).Scan(&cartId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Query the database to get the total cost of the items in the cart.
	err = db.QueryRow("SELECT cost FROM cart WHERE customer_id=$1", customerId).Scan(&cartCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cartObjectId := c.Param("id")
	var cartIdCheck int

	err = db.QueryRow("SELECT cart_id FROM cart_object WHERE id=$1", cartObjectId).Scan(&cartIdCheck)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Item not found"})
		return
	}

	if cartIdCheck != cartId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Item not in your cart"})
		return
	}

	var product_id int
	var quantity int

	err = db.QueryRow("SELECT product_id, units FROM cart_object WHERE id = $1", cartObjectId).Scan(&product_id, &quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var product_price int
	err = db.QueryRow("SELECT price FROM product WHERE id = $1", product_id).Scan(&product_price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	money := product_price * quantity
	cartCost = cartCost - money

	_, err = db.Exec("UPDATE cart SET cost = $1 WHERE customer_id = $2", cartCost, customerId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = db.Exec("DELETE FROM cart_object WHERE id = $1", cartObjectId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Message": "Item removed from your cart successfully"})
}
