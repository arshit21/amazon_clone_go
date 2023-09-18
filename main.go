package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// database details
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

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	//Create a new gin router
	router := gin.Default()

	//setup session management and session store using appropriate packages
	store := cookie.NewStore([]byte("c1d545ff2d75aae60f492d3e06ec0a10"))
	router.Use(sessions.Sessions("mysession", store))

	//define the api routes
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
	router.GET("vendors/my_products", authMiddleware(), func(c *gin.Context) {
		getMyProducts(c, db)
	})
	router.GET("customers/wallet", authMiddleware(), func(c *gin.Context) {
		getWalletDetails(c, db)
	})
	router.PUT("customers/wallet", authMiddleware(), func(c *gin.Context) {
		addMoneytoWallet(c, db)
	})
	router.GET("products/", func(c *gin.Context) {
		getAllProducts(c, db)
	})
	router.GET("products/:id", func(c *gin.Context) {
		getIndividualProduct(c, db)
	})
	router.PATCH("products/:id", authMiddleware(), func(c *gin.Context) {
		updateProductDetails(c, db)
	})
	router.POST("products/:id/buyNow", authMiddleware(), func(c *gin.Context) {
		buyNow(c, db)
	})
	router.POST("products/:id/addToCart", authMiddleware(), func(c *gin.Context) {
		addToCart(c, db)
	})
	router.GET("customers/previousOrders", authMiddleware(), func(c *gin.Context) {
		previousOrders(c, db)
	})
	router.GET("customers/cart", authMiddleware(), func(c *gin.Context) {
		viewcart(c, db)
	})
	router.POST("customers/cart/buy", authMiddleware(), func(c *gin.Context) {
		buycart(c, db)
	})
	router.GET("vendors/myOrders", authMiddleware(), func(c *gin.Context) {
		ordersForMe(c, db)
	})
	router.DELETE("customers/cart/:id", authMiddleware(), func(c *gin.Context) {
		removeFromCart(c, db)
	})
	router.Run("localhost:8080")

}
