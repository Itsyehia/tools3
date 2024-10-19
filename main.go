package main

import (
    "database/sql"
    "log"
    "net/http"
    "text/template"

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
    dbName := "tools" // Change to your database name
    db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
    if err != nil {
        panic(err.Error())
    }
    return db
}

var tmpl = template.Must(template.ParseGlob("form/*"))

func main() {
    log.Println("Server started on: http://localhost:4200")
    http.HandleFunc("/register", Register)
    http.HandleFunc("/login", Login)
    http.ListenAndServe(":4200", nil)
}

func Register(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        username := r.FormValue("username")
        password := r.FormValue("password")

        // Hash the password
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            http.Error(w, "Error hashing password", http.StatusInternalServerError)
            return
        }

        db := dbConn()
        defer db.Close()

        // Insert new user into the database
        insForm, err := db.Prepare("INSERT INTO users(username, password) VALUES(?,?)")
        if err != nil {
            panic(err.Error())
        }
        _, err = insForm.Exec(username, hashedPassword)
        if err != nil {
            http.Error(w, "Error creating user", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/login", http.StatusSeeOther)
    } else {
        tmpl.ExecuteTemplate(w, "Register", nil)
    }
}

func Login(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        username := r.FormValue("username")
        password := r.FormValue("password")

        db := dbConn()
        defer db.Close()

        // Get the user from the database
        var user User
        err := db.QueryRow("SELECT id, password FROM users WHERE username=?", username).Scan(&user.ID, &user.Password)
        if err != nil {
            http.Error(w, "User not found", http.StatusUnauthorized)
            return
        }

        // Compare the hashed password with the provided password
        err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
        if err != nil {
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        }

        // User logged in successfully, handle session or redirect
        http.Redirect(w, r, "/", http.StatusSeeOther)
    } else {
        tmpl.ExecuteTemplate(w, "Login", nil)
    }
}
