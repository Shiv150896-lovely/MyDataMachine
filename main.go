package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var (
	db      *sql.DB
	dbMutex sync.Mutex // Mutex for protecting access to the database
)

type Employee struct {
	ID       int     `json:"Id"`
	Name     string  `json:"Name"`
	Position string  `json:"Position"`
	Salary   float64 `json:"Salary"`
}

func InitDB() {
	var err error
	db, err = sql.Open("mysql", "root:Shivuma@tcp(localhost:3306)/MyData_Database")
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS employees (
		  Id INT AUTO_INCREMENT PRIMARY KEY,
		  Name VARCHAR(255) NOT NULL,
		  Position VARCHAR(255) NOT NULL,
		  Salary FLOAT
	  )`)
	if err != nil {
		panic("Failed to create employees table: " + err.Error())
	}

	fmt.Println("Employee table created successfully")
}

func main() {
	r := gin.Default()
	InitDB()

	// CRUD endpoints
	r.POST("/employees", func(c *gin.Context) {
		var emp Employee
		if err := c.BindJSON(&emp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		dbMutex.Lock()
		defer dbMutex.Unlock()

		_, err := db.Exec("INSERT INTO employees (Name, Position, Salary) VALUES (?, ?, ?)", emp.Name, emp.Position, emp.Salary)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Employee created successfully"})
	})

	r.GET("/employees/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
			return
		}

		dbMutex.Lock()
		defer dbMutex.Unlock()

		var emp Employee
		err = db.QueryRow("SELECT Id, Name, Position, Salary FROM employees WHERE Id = ?", id).Scan(&emp.ID, &emp.Name, &emp.Position, &emp.Salary)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}

		c.JSON(http.StatusOK, emp)
	})

	r.PUT("/employees/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
			return
		}

		var emp Employee
		if err := c.BindJSON(&emp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		dbMutex.Lock()
		defer dbMutex.Unlock()

		_, err = db.Exec("UPDATE employees SET Name = ?, Position = ?, Salary = ? WHERE Id = ?", emp.Name, emp.Position, emp.Salary, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Employee updated successfully"})
	})

	r.DELETE("/employees/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
			return
		}

		dbMutex.Lock()
		defer dbMutex.Unlock()

		_, err = db.Exec("DELETE FROM employees WHERE Id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete employee"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Employee deleted successfully"})
	})

	// Start the Gin server
	if err := r.Run(":8080"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
