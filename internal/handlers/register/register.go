package register

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"meditrack-backend/internal/models"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const maxRequestBody = 1 << 20 // 1 MB

type registerResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, payload registerResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			respondJSON(w, http.StatusMethodNotAllowed, registerResponse{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		if db == nil {
			respondJSON(w, http.StatusInternalServerError, registerResponse{
				Success: false,
				Message: "Server error",
			})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
		defer r.Body.Close()

		var req models.RegisterRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, registerResponse{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		// Reject trailing JSON or any additional JSON value.
		var extra interface{}
		if err := decoder.Decode(&extra); !errors.Is(err, nil) {
			if !errors.Is(err, sql.ErrNoRows) && err.Error() != "EOF" {
				respondJSON(w, http.StatusBadRequest, registerResponse{
					Success: false,
					Message: "Invalid request body",
				})
				return
			}
		} else {
			respondJSON(w, http.StatusBadRequest, registerResponse{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		data := req.Data

		data.FirstName = strings.TrimSpace(data.FirstName)
		data.LastName = strings.TrimSpace(data.LastName)
		data.Email = strings.ToLower(strings.TrimSpace(data.Email))
		data.Phone = strings.TrimSpace(data.Phone)

		if data.FirstName == "" ||
			data.LastName == "" ||
			data.Email == "" ||
			data.Phone == "" ||
			strings.TrimSpace(data.Password) == "" {
			respondJSON(w, http.StatusBadRequest, registerResponse{
				Success: false,
				Message: "All fields are required",
			})
			return
		}

		if _, err := mail.ParseAddress(data.Email); err != nil {
			respondJSON(w, http.StatusBadRequest, registerResponse{
				Success: false,
				Message: "Invalid email address",
			})
			return
		}

		if !utf8.ValidString(data.Password) || len([]byte(data.Password)) > 72 {
			respondJSON(w, http.StatusBadRequest, registerResponse{
				Success: false,
				Message: "Password must be valid UTF-8 and no longer than 72 bytes",
			})
			return
		}

		if utf8.RuneCountInString(data.Password) < 8 {
			respondJSON(w, http.StatusBadRequest, registerResponse{
				Success: false,
				Message: "Password must be at least 8 characters",
			})
			return
		}

		// This registration flow only accepts role ID 1.
		if data.Role != 1 {
			respondJSON(w, http.StatusBadRequest, registerResponse{
				Success: false,
				Message: "Invalid role",
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(data.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			log.Printf("Password hashing failed: %v", err)
			respondJSON(w, http.StatusInternalServerError, registerResponse{
				Success: false,
				Message: "Server error",
			})
			return
		}

		query := `
			INSERT INTO users (
				firstname,
				lastname,
				phone,
				email,
				password,
				role
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT DO NOTHING
			RETURNING id, status, created_at
		`

		var (
			id        int
			status    int
			createdAt sql.NullTime
		)

		err = db.QueryRowContext(
			r.Context(),
			query,
			data.FirstName,
			data.LastName,
			data.Phone,
			data.Email,
			string(hashedPassword),
			data.Role,
		).Scan(&id, &status, &createdAt)

		if errors.Is(err, sql.ErrNoRows) {
			respondJSON(w, http.StatusConflict, registerResponse{
				Success: false,
				Message: "Email or phone already exists",
			})
			return
		}

		if err != nil {
			log.Printf("User registration failed: %v", err)
			respondJSON(w, http.StatusInternalServerError, registerResponse{
				Success: false,
				Message: "Server error",
			})
			return
		}

		var createdAtValue interface{}
		if createdAt.Valid {
			createdAtValue = createdAt.Time
		}

		respondJSON(w, http.StatusCreated, registerResponse{
			Success: true,
			Message: "User registered successfully",
			Data: map[string]interface{}{
				"id":         id,
				"firstName":  data.FirstName,
				"lastName":   data.LastName,
				"phone":      data.Phone,
				"email":      data.Email,
				"role":       data.Role,
				"status":     status,
				"created_at": createdAtValue,
			},
		})
	}
}
