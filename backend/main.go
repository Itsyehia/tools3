package main

import (
    "database/sql"
    "log"
    "net/http"
    "text/template"
    "encoding/json"

    _ "github.com/go-sql-driver/mysql"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Password string `json:"-"`
}

func dbConn() (db *sql.DB) {
    dbDriver := "mysql"
    dbUser := "root"
    dbPass := "root"
    dbName := "tools"
    db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
    if err != nil {
        panic(err.Error())
    }
    return db
}

var tmpl = template.Must(template.ParseGlob("form/*"))

func main() {
    log.Println("Server started on: http://localhost:4300")
    http.HandleFunc("/register", handleCORS(Register)) // Wrap Register with CORS
    http.HandleFunc("/login", handleCORS(Login))       // Wrap Login with CORS

    log.Fatal(http.ListenAndServe(":4300", nil))
}

// Middleware to handle CORS
func handleCORS(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Allow CORS for specific origin (you can change this to your frontend URL)
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == http.MethodOptions {
            // Handle preflight request
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r) // Call the next handler
    }
}

func Register(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        var user User
        err := json.NewDecoder(r.Body).Decode(&user)
        if err != nil {
            http.Error(w, "Invalid request payload", http.StatusBadRequest)
            return
        }

        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
        if err != nil {
            http.Error(w, "Error hashing password", http.StatusInternalServerError)
            return
        }

        db := dbConn()
        defer db.Close()

        insForm, err := db.Prepare("INSERT INTO users(username, password) VALUES(?,?)")
        if err != nil {
            panic(err.Error())
        }
        _, err = insForm.Exec(user.Username, hashedPassword)
        if err != nil {
            http.Error(w, "Error creating user", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
    } else {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func Login(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        username := r.FormValue("username")
        password := r.FormValue("password")

        db := dbConn()
        defer db.Close()

        var user User
        err := db.QueryRow("SELECT id, password FROM users WHERE username=?", username).Scan(&user.ID, &user.Password)
        if err != nil {
            http.Error(w, "User not found", http.StatusUnauthorized)
            return
        }

        err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
        if err != nil {
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        }

        http.Redirect(w, r, "/", http.StatusSeeOther)
    } else {
        tmpl.ExecuteTemplate(w, "Login", nil)
    }
}
