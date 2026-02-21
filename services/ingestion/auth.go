package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("SUPER_SECRET_PHISHVAULT_KEY") // In production, load from env

// User struct
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"` // Only for input
	CreatedAt time.Time `json:"created_at"`
}

// Credentials for Login/Register
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// JWT Claims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// RegisterHandler creates a new user
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	if creds.Username == "" || creds.Password == "" {
		http.Error(w, "Username and Password are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// Insert into DB
	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", creds.Username, string(hashedPassword))
	if err != nil {
		// Check for unique constraint violation (simplified check)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// LoginHandler authenticates a user and returns a token
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", creds.Username).Scan(&storedHash)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Token Generation Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	json.NewEncoder(w).Encode(map[string]string{
		"token":    tokenString,
		"username": creds.Username,
	})
}

// AuthMiddleware validates JWT token
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("AuthMiddleware: Request %s %s", r.Method, r.URL.Path)
		// handle CORS preflight for protected routes
		if r.Method == "OPTIONS" {
			log.Printf("AuthMiddleware: Handling OPTIONS preflight")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			return
		}

		tokenString := ""

		// 1. Try Header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. Try Cookie (optional, but good for browser)
		if tokenString == "" {
			c, err := r.Cookie("token")
			if err == nil {
				tokenString = c.Value
			}
		}

		if tokenString == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "Unauthorized: Invalid signature", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Unauthorized: Token invalid", http.StatusUnauthorized)
			return
		}

		// Proceed
		next(w, r)
	}
}
