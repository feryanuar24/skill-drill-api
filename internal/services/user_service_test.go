package services_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"skill-drill-api/internal/api"
	"skill-drill-api/internal/services"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func boolPtr(b bool) *bool {
	return &b
}

func TestCreateUserRequest_Validation(t *testing.T) {
	r := gin.New()
	r.POST("/users", func(c *gin.Context) {
		var req services.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errors": api.FormatValidationError(err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"role":        req.Role,
			"is_verified": req.IsVerified,
		})
	})

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
		checkResp  func(t *testing.T, resp map[string]interface{})
	}{
		{
			name: "success with role admin and is_verified true",
			body: map[string]interface{}{
				"username":    "adminuser",
				"email":       "admin@example.com",
				"password":    "SecurePass123!",
				"role":        "admin",
				"is_verified": true,
			},
			wantStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp map[string]interface{}) {
				if resp["role"] != "admin" {
					t.Errorf("expected role 'admin', got %v", resp["role"])
				}
				if resp["is_verified"] != true {
					t.Errorf("expected is_verified true, got %v", resp["is_verified"])
				}
			},
		},
		{
			name: "success with role student and is_verified false",
			body: map[string]interface{}{
				"username":    "studentuser",
				"email":       "student@example.com",
				"password":    "SecurePass123!",
				"role":        "student",
				"is_verified": false,
			},
			wantStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp map[string]interface{}) {
				if resp["role"] != "student" {
					t.Errorf("expected role 'student', got %v", resp["role"])
				}
				if resp["is_verified"] != false {
					t.Errorf("expected is_verified false, got %v", resp["is_verified"])
				}
			},
		},
		{
			name: "success with omitted role and is_verified",
			body: map[string]interface{}{
				"username": "basicuser",
				"email":    "basic@example.com",
				"password": "SecurePass123!",
			},
			wantStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp map[string]interface{}) {
				if resp["is_verified"] != nil {
					t.Errorf("expected is_verified nil, got %v", resp["is_verified"])
				}
			},
		},
		{
			name: "invalid role",
			body: map[string]interface{}{
				"username": "invalidrole",
				"email":    "role@example.com",
				"password": "SecurePass123!",
				"role":     "superadmin",
			},
			wantStatus: http.StatusBadRequest,
			checkResp:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.wantStatus, w.Code, w.Body.String())
			}

			if tc.checkResp != nil {
				var resp map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestUpdateUserRequest_Validation(t *testing.T) {
	r := gin.New()
	r.PUT("/users/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
		if id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		var req services.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errors": api.FormatValidationError(err)})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
		checkErr   func(t *testing.T, body string)
	}{
		{
			name: "valid with only is_verified",
			body: map[string]interface{}{
				"is_verified": true,
			},
			wantStatus: http.StatusOK,
			checkErr:   nil,
		},
		{
			name: "valid with only username",
			body: map[string]interface{}{
				"username": "updatedname",
			},
			wantStatus: http.StatusOK,
			checkErr:   nil,
		},
		{
			name: "valid with only email",
			body: map[string]interface{}{
				"email": "updated@example.com",
			},
			wantStatus: http.StatusOK,
			checkErr:   nil,
		},
		{
			name: "valid with only password",
			body: map[string]interface{}{
				"password": "NewValidPass123!",
			},
			wantStatus: http.StatusOK,
			checkErr:   nil,
		},
		{
			name: "valid with only role",
			body: map[string]interface{}{
				"role": "admin",
			},
			wantStatus: http.StatusOK,
			checkErr:   nil,
		},
		{
			name: "invalid role",
			body: map[string]interface{}{
				"role": "superadmin",
			},
			wantStatus: http.StatusBadRequest,
			checkErr: func(t *testing.T, body string) {
				var resp map[string]map[string][]string
				_ = json.Unmarshal([]byte(body), &resp)
				if _, ok := resp["errors"]["role"]; !ok {
					t.Errorf("expected validation error on role, got %s", body)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest(http.MethodPut, "/users/10", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.wantStatus, w.Code, w.Body.String())
			}

			if tc.checkErr != nil {
				tc.checkErr(t, w.Body.String())
			}
		})
	}
}
