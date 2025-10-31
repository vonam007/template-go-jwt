package controllers

import (
	"encoding/json"
	"net/http"

	"template-go-jwt/internal/services"

	"template-go-jwt/internal/util/middleware"
	"template-go-jwt/internal/util/response"
)

type AuthController struct {
	svc *services.AuthService
}

func NewAuthController(svc *services.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

type loginRequest struct {
	UserID string `json:"user_id"`
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var lr loginRequest
	if err := json.NewDecoder(r.Body).Decode(&lr); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if lr.UserID == "" {
		response.Error(w, http.StatusBadRequest, "user_id required")
		return
	}
	token, err := c.svc.Login(lr.UserID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"token": token})
}

func (c *AuthController) Protected(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	response.JSON(w, http.StatusOK, map[string]string{"message": "protected", "user_id": userID})
}
