package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AddRequest struct {
	ToAdd int `json:"toAdd"`
}

type UnitsRequest struct {
	Units int `json:"units"`
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
	product_id := c.Param("id")
	product_details := db.QueryRow("SELECT title, brand, price, description, image, category, units FROM product WHERE id = $1", product_id)
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

func getWalletDetails(c *gin.Context, db *sql.DB) {
	session := sessions.Default(c)
	username := session.Get("username")

	var customer_id int
	err := db.QueryRow("SELECT id FROM customers WHERE username = $1", username).Scan(&customer_id)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not a customer"})
		return
	}
	var wallet_balance int
	err = db.QueryRow("SELECT balance FROM wallet WHERE customer_id = $1", customer_id).Scan(&wallet_balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"Wallet Balance": wallet_balance})
}

func addMoneytoWallet(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	session := sessions.Default(c)
	username := session.Get("username")

	var customer_id int
	err := db.QueryRow("SELECT id FROM customers WHERE username = $1", username).Scan(&customer_id)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not a customer"})
		return
	}
	var wallet_balance int
	err = db.QueryRow("SELECT balance FROM wallet WHERE customer_id = $1", customer_id).Scan(&wallet_balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var addRequest AddRequest
	err = c.BindJSON(&addRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	toAdd := addRequest.ToAdd
	newBalance := wallet_balance + toAdd

	_, err = db.Exec("UPDATE wallet SET balance = $1 WHERE customer_id = $2", newBalance, customer_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Money has been added", "Wallet Balance": newBalance})
}

func buyNow(c *gin.Context, db *sql.DB) {
	product_id := c.Param("id")
	product_details := db.QueryRow("SELECT title, brand, price, description, image, category, units FROM product WHERE id = $1", product_id)
	var product Product

	err := product_details.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product does not exist"})
		return
	}

	c.Header("Content-Type", "application/json")
	session := sessions.Default(c)
	username := session.Get("username")
	var customerId int
	err = db.QueryRow("SELECT id from customers where username=$1", username).Scan(&customerId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}
	var units UnitsRequest
	err = c.BindJSON(&units)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Payload"})
		return
	}
	if units.Units > product.Units {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Out of Stock"})
		return
	}
	moneyPayable := product.Price * units.Units
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

	stmt, err := db.Prepare("INSERT into orders (product_id, customer_id, money_paid, units, vendor_id) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(product_id, customerId, moneyPayable, units.Units, product_vendor_id); err != nil {
		log.Fatal(err)
	}
	product.Units = product.Units - units.Units
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

	var Order orderDetails
	Order.Product = product.Title
	Order.MoneyPaid = moneyPayable
	Order.Units = units.Units
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Product ordered", "Order details": Order})
}

func previousOrders(c *gin.Context, db *sql.DB) {
	session := sessions.Default(c)
	username := session.Get("username")
	var customerId int
	err := db.QueryRow("SELECT id from customers where username=$1", username).Scan(&customerId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}
	rows, err := db.Query("SELECT product_id, money_paid, units, date_created FROM orders WHERE customer_id = $1", customerId)
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
	c.IndentedJSON(http.StatusOK, gin.H{"Here are all your Orders": orders})
}

func addToCart(c *gin.Context, db *sql.DB) {
	product_id := c.Param("id")
	product_details := db.QueryRow("SELECT title, brand, price, description, image, category, units FROM product WHERE id = $1", product_id)
	var product Product

	err := product_details.Scan(&product.Title, &product.Brand, &product.Price, &product.Description, &product.Image, &product.Category, &product.Units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product does not exist"})
		return
	}

	c.Header("Content-Type", "application/json")
	session := sessions.Default(c)
	username := session.Get("username")
	var customerId int
	err = db.QueryRow("SELECT id from customers WHERE username=$1", username).Scan(&customerId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}

	var units UnitsRequest
	err = c.BindJSON(&units)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Payload"})
		return
	}
	if units.Units > product.Units {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Out of Stock"})
		return
	}

	var cartId int
	err = db.QueryRow("SELECT id FROM cart WHERE customer_id = $1", customerId).Scan(&cartId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	moneyPayable := product.Price * units.Units
	stmt, err := db.Prepare("INSERT INTO cart_object (product_id, customer_id, cart_id, units, money) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(product_id, customerId, cartId, units.Units, moneyPayable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var cartCost int
	err = db.QueryRow("SELECT cost FROM cart WHERE customer_id = $1", customerId).Scan(&cartCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cartCost = cartCost + moneyPayable

	_, err = db.Exec("UPDATE cart SET cost = $1 WHERE customer_id = $2", cartCost, customerId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "added to cart"})
}

func viewcart(c *gin.Context, db *sql.DB) {
	session := sessions.Default(c)
	username := session.Get("username")
	var customerId int
	err := db.QueryRow("SELECT id from customers where username=$1", username).Scan(&customerId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not a customer"})
		return
	}
	var cartId int
	err = db.QueryRow("SELECT id FROM cart WHERE customer_id =$1 ", customerId).Scan(&cartId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := db.Query("SELECT product_id, units FROM cart_object WHERE cart_id = $1", cartId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var cart_objects []shoppingCart
	for rows.Next() {
		var product_id int
		var product_units int
		err = rows.Scan(&product_id, &product_units)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var product_title string
		err = db.QueryRow("SELECT title FROM product WHERE id = $1", product_id).Scan(&product_title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var cart_object shoppingCart
		cart_object.Product = product_title
		cart_object.Units = product_units
		cart_objects = append(cart_objects, cart_object)
	}
	var cartCost int
	err = db.QueryRow("SELECT cost FROM cart WHERE customer_id=$1", customerId).Scan(&cartCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"Cart total": cartCost, "Items": cart_objects})
}
