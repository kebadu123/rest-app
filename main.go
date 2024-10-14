package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID    int    `json:"id"`
	Fname string `json:"first_name"`
	Lname string `json:"last_name"`
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/mydb")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	fmt.Println("Connected to the database successfully!")
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(TrustedProxyMiddleware())
	r.POST("/users", func(c *gin.Context) {
		createUser(c, db)
	})
	r.GET("/users", func(c *gin.Context) {
		getUsers(c, db)
	})
	r.PUT("/users/:id", func(c *gin.Context) {
		updateUser(c, db)
	})
	r.DELETE("/users/:id", func(c *gin.Context) {
		deleteUser(c, db)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

func createUser(c *gin.Context, db *sql.DB) {
	var user User
	// Bind the JSON data to the User struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert the user data into the database
	result, err := db.Exec("INSERT INTO person (id, first_name, last_name) VALUES (?, ?, ?)", user.ID, user.Fname, user.Lname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func getUsers(c *gin.Context, db *sql.DB) {
	rows, err := db.Query("SELECT id, first_name, last_name FROM person")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Fname, &user.Lname); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func updateUser(c *gin.Context, db *sql.DB) {
	id := c.Param("id")
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("UPDATE person SET first_name = ?, last_name = ? WHERE id = ?", user.Fname, user.Lname, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rows_affected": rowsAffected})
}

func deleteUser(c *gin.Context, db *sql.DB) {
	id := c.Param("id")
	result, err := db.Exec("DELETE FROM person WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rows_affected": rowsAffected})
}

// TrustedProxyMiddleware is a middleware that handles trusted proxies
func TrustedProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check the X-Forwarded-For header to determine the real client IP address
		clientIP := c.Request.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = c.ClientIP()
		}
		c.Set("X-Real-IP", clientIP)
		c.Next()
	}
}
