package main

import (
    "log"
    "net/http"
    "os"

    "ecommerce-backend/internal/database"
    "ecommerce-backend/internal/routes"

    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    // Get database URL
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        log.Fatal("DATABASE_URL is missing")
    }

    // Connect to PostgreSQL
    db, err := database.Connect(databaseURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Test database query
    var databaseName string

    err = db.QueryRow("SELECT current_database()").Scan(&databaseName)
    if err != nil {
        log.Fatal("Failed to execute database query:", err)
    }

    log.Println("Successfully connected to PostgreSQL!")
    log.Println("Connected database:", databaseName)

    // Register API routes
  	mux := routes.RegisterRoutes(db)

    // Get application port
    port := os.Getenv("APP_PORT")
    if port == "" {
        port = "8080"
    }

    // Start HTTP server
    log.Println("Server is running on port", port)

    err = http.ListenAndServe(":"+port, mux)
    if err != nil {
        log.Fatal("Failed to start server:", err)
    }
}