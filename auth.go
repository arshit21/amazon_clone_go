package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
)

// defined a User struct for storing user data
type User struct {
	First_name  string `json:"first_name"`
	Last_name   string `json:"last_name"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Is_customer bool   `json:"is_customer"`
	Is_vendor   bool   `json:"is_vendor"`
	ID          int    `json:"id"`
}

// Function to get current user details
func getCurrentUser(c *gin.Context, db *sql.DB) {
	//get current user details using gin sessions
	session := sessions.Default(c)
	username := session.Get("username")

	//get current user details from database using username stored in session and scan it to a user variable of User struct type
	details := db.QueryRow("SELECT id, first_name, last_name, username, email, is_customer, is_vendor from users WHERE username = $1", username)

	var user User
	err := details.Scan(&user.ID, &user.First_name, &user.Last_name, &user.Username, &user.Email, &user.Is_customer, &user.Is_vendor)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}
	//output the user details
	c.IndentedJSON(http.StatusOK, user)
}

func UpdateMyDetails(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")
	//get current user details stored in a cookie
	session := sessions.Default(c)
	username := session.Get("username")

	//get current user details from database using username stored in session and scan it to a user variable of User struct type
	details := db.QueryRow("SELECT id, first_name, last_name, username, email, is_customer, is_vendor from users WHERE username = $1", username)

	var user User
	err_ := details.Scan(&user.ID, &user.First_name, &user.Last_name, &user.Username, &user.Email, &user.Is_customer, &user.Is_vendor)
	if err_ != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err_.Error()})
		return
	}

	//create updated user of type User
	var updatedUser User

	//bind json data from request to the updateUser struct
	if err := c.BindJSON(&updatedUser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	//check if the following fields were present in the request data, if not, set it to its original value
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

	//update the users table with the updatedUser struct
	_, err := db.Exec("UPDATE users SET first_name = $1, last_name = $2, username = $3, email = $4 WHERE username = $5", updatedUser.First_name, updatedUser.Last_name, updatedUser.Username, updatedUser.Email, username)

	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	//check for whether the user is a vendor or customer, and update the respective table with the updated values
	if user.Is_customer {
		_, err_ := db.Exec("UPDATE customers SET first_name = $1, last_name = $2, username = $3, email = $4 WHERE username = $5", updatedUser.First_name, updatedUser.Last_name, updatedUser.Username, updatedUser.Email, username)
		if err_ != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err_.Error()})
			return
		}
	}
	if user.Is_vendor {
		_, _err_ := db.Exec("UPDATE vendors SET first_name = $1, last_name = $2, username = $3, email = $4 WHERE username = $5", updatedUser.First_name, updatedUser.Last_name, updatedUser.Username, updatedUser.Email, username)
		if _err_ != nil {
			log.Fatal(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
	}
	//return success message
	c.JSON(http.StatusOK, gin.H{"message": "User details updated successfully", "Updated User": updatedUser})
}

func createCustomer(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")

	//create a newuser of type User and bind the json data from the request to this newuser
	var newuser User
	if err := c.BindJSON(&newuser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	//hash the password to securely store it in thte database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newuser.Password), bcrypt.DefaultCost)

	newuser.Is_customer = true
	if err != nil {
		log.Fatal(err)
	}

	//prepare a SQL statement for inserting data into customers table and handling any potential errors
	stmt, err := db.Prepare("INSERT INTO customers (first_name, last_name, username, email, password) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	//Execute the sql statement and populate the table values with the approriate details stored in the newuser struct
	if _, err := stmt.Exec(newuser.First_name, newuser.Last_name, newuser.Username, newuser.Email, hashedPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
		return
	}

	//prepare a SQL statement for inserting data into users table and handling any potential errors
	stmt_2, err := db.Prepare("INSERT INTO users (first_name, last_name, username, email, password, is_customer) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt_2.Close()

	//Execute the sql statement and populate the table values with the approriate details stored in the newuser struct
	if _, err := stmt_2.Exec(newuser.First_name, newuser.Last_name, newuser.Username, newuser.Email, hashedPassword, newuser.Is_customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
		return
	}

	//get customer_id of the newly registered customer from the database
	var customer_id int
	err = db.QueryRow("SELECT id FROM customers WHERE username = $1", newuser.Username).Scan(&customer_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//prepare a SQL statement for inserting data into wallet table and handling any potential errors
	stmt_3, err := db.Prepare("INSERT INTO wallet (balance, customer_id) VALUES ($1, $2)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt_3.Close()

	//Execute the SQL statement and set the customer wallet to 0
	if _, err := stmt_3.Exec(0, customer_id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//prepare a SQL statement for inserting data into cart table and handling any potential errors
	stmt_4, err := db.Prepare("INSERT INTO cart (cost, customer_id) VALUES ($1, $2)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt_4.Close()

	//Execute the SQL statement and populate the table with appropriate data
	if _, err := stmt_4.Exec(0, customer_id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Here are your details", "details": newuser})
}

func createVendors(c *gin.Context, db *sql.DB) {
	c.Header("Content-Type", "application/json")

	//create a newuser of type User and bind the json data from the request to this newuser
	var newuser User
	if err := c.BindJSON(&newuser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	//hash the password to securely store it in thte database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newuser.Password), bcrypt.DefaultCost)
	newuser.Is_vendor = true
	if err != nil {
		log.Fatal(err)
	}

	//prepare a SQL statement for inserting data into vendors table and handling any potential errors
	stmt, err := db.Prepare("INSERT INTO vendors (first_name, last_name, username, email, password) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	//Execute the sql statement and populate the table values with the approriate details stored in the newuser struct
	if _, err := stmt.Exec(newuser.First_name, newuser.Last_name, newuser.Username, newuser.Email, hashedPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
		return
	}

	//prepare a SQL statement for inserting data into users table and handling any potential errors
	stmt_2, err := db.Prepare("INSERT INTO users (first_name, last_name, username, email, password, is_vendor) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt_2.Close()

	//Execute the sql statement and populate the table values with the approriate details stored in the newuser struct
	if _, err := stmt_2.Exec(newuser.First_name, newuser.Last_name, newuser.Username, newuser.Email, hashedPassword, newuser.Is_vendor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or email already exists"})
		return
	}

	c.IndentedJSON(http.StatusCreated, newuser)
}

func loginHandler(c *gin.Context, db *sql.DB) {
	//create a loginRequest Struct to store the request data
	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	//bind request json data to loginRequest struct
	if err := c.BindJSON(&loginRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	//get user details using the username provided in the request data with appropriate error handling
	var user User
	err := db.QueryRow("SELECT * FROM users WHERE username = $1", loginRequest.Username).
		Scan(&user.First_name, &user.Last_name, &user.Username, &user.Email, &user.Password, &user.Is_customer, &user.Is_vendor, &user.ID)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username"})
		return
	}

	//compare the hashed password stored in the database to the one provided in the request
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	//set session values to store the login information
	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("authenticated", true)
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": user})

}

func handleLogout(c *gin.Context) {
	//clear the sessions to logout
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// function to check if the user is logged in using sessions
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		auth := session.Get("authenticated")
		if auth != nil && auth.(bool) {
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not logged In"})
			c.Abort()
		}
	}
}
