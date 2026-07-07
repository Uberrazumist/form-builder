package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    _ "github.com/lib/pq"
)

type RegisterInput struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    FullName string `json:"full_name"`
}

func main() {
    // Переменные окружения
    host := os.Getenv("DB_HOST")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    port := os.Getenv("DB_PORT")
    if host == "" { host = "localhost" }
    if user == "" { user = "formuser" }
    if password == "" { password = "formpass" }
    if dbname == "" { dbname = "formsdb" }
    if port == "" { port = "5432" }

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        host, user, password, dbname, port)

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Failed to connect: ", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal("Cannot ping DB: ", err)
    }
    log.Println("Connected to DB")

    // Создаём таблицу users (если нет)
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            email TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL,
            full_name TEXT,
            role TEXT DEFAULT 'viewer',
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
    if err != nil {
        log.Fatal("Failed to create table: ", err)
    }
    log.Println("Table users ensured")

    // Обработчик регистрации
    http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var input RegisterInput
        if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }
        if input.Email == "" || input.Password == "" {
            http.Error(w, "Email and password required", http.StatusBadRequest)
            return
        }
        // (хеширование опускаем, добавим позже)
        _, err := db.Exec("INSERT INTO users (email, password, full_name) VALUES ($1, $2, $3)",
            input.Email, input.Password, input.FullName)
        if err != nil {
            http.Error(w, "User already exists or DB error", http.StatusConflict)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{
            "message": "User registered successfully",
        })
    })

    // Тестовый пинг
    http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("pong"))
    })

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
