package login

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"meditrack-backend/internal/models"
	"meditrack-backend/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

const maxRequestBodySize = 1 << 20 // 1 MB

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type LoginUserResponse struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	Status    int       `json:"status"`
	Role      int       `json:"role"`
}

func respondJSON(w http.ResponseWriter, status int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func LoginHandler(db *sql.DB, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Validate HTTP method.
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			respondJSON(w, http.StatusMethodNotAllowed, APIResponse{
				Success: false,
				Message: "Method not allowed.",
			})
			return
		}

		// 2. Check required dependencies.
		if db == nil {
			log.Println("Login failed: database connection is nil")
			respondJSON(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "An internal server error occurred.",
			})
			return
		}

		// 3. Limit and close the request body.
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
		defer r.Body.Close()

		// 4. Decode JSON.
		var req models.LoginRequest

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			var maxBytesErr *http.MaxBytesError

			if errors.As(err, &maxBytesErr) {
				respondJSON(w, http.StatusRequestEntityTooLarge, APIResponse{
					Success: false,
					Message: "Request body is too large.",
				})
				return
			}

			respondJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid JSON request body.",
			})
			return
		}

		// Reject multiple JSON values or trailing data.
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			respondJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Request body must contain a single JSON object.",
			})
			return
		}

		// 5. Validate credentials.
		email := strings.TrimSpace(req.Data.Email)
		password := req.Data.Password

		if email == "" || strings.TrimSpace(password) == "" {
			respondJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Email and password are required.",
			})
			return
		}

		if strings.TrimSpace(jwtSecret) == "" {
			log.Println("Login failed: JWT secret is not configured")
			respondJSON(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "An internal server error occurred.",
			})
			return
		}

		// 6. Find user.
		var user models.User
		var hashedPassword string

		err := db.QueryRowContext(
			r.Context(),
			`SELECT id, firstName, lastName, phone, email,
			        password, role, status, created_at
			 FROM users
			 WHERE email = $1`,
			email,
		).Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Phone,
			&user.Email,
			&hashedPassword,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
		)

		if errors.Is(err, sql.ErrNoRows) {
			respondJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "User does not exist.",
			})
			return
		}

		if err != nil {
			log.Printf("Login database query failed: %v", err)
			respondJSON(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "An internal server error occurred.",
			})
			return
		}

		// 7. Verify password.
		if err := bcrypt.CompareHashAndPassword(
			[]byte(hashedPassword),
			[]byte(password),
		); err != nil {
			respondJSON(w, http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Invalid email or password.",
			})
			return
		}

		// 8. Reject inactive users.
		if user.Status != 1 {
			respondJSON(w, http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Invalid email or password.",
			})
			return
		}

		// 9. Generate JWT using the configured secret.
		token, err := utils.GenerateJWT(user.Email, jwtSecret)
		if err != nil {
			log.Printf("Login JWT generation failed: %v", err)
			respondJSON(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Unable to complete login.",
			})
			return
		}

		// 10. Return only public user information.
		userResponse := LoginUserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt,
			Status:    user.Status,
			Role:      user.Role,
		}

		respondJSON(w, http.StatusOK, APIResponse{
			Success: true,
			Message: "Login successful.",
			Data: map[string]interface{}{
				"user":  userResponse,
				"token": token,
			},
		})
	}
}
