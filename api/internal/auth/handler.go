package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/platform/config"
)

const refreshCookieName = "refresh_token"

type Handler struct {
	repo   *Repository
	tokens *TokenManager
	cfg    config.Config
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type membershipResponse struct {
	TenantID string `json:"tenant_id"`
	BranchID string `json:"branch_id"`
	Role     string `json:"role"`
}

type authResponse struct {
	AccessToken string               `json:"access_token"`
	TokenType   string               `json:"token_type"`
	ExpiresIn   int64                `json:"expires_in_seconds"`
	User        userResponse         `json:"user"`
	Memberships []membershipResponse `json:"memberships"`
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewHandler(repo *Repository, tokens *TokenManager, cfg config.Config) *Handler {
	return &Handler{repo: repo, tokens: tokens, cfg: cfg}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	auth := group.Group("/auth")
	auth.POST("/login", h.login)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request payload.",
			"details":    err.Error(),
			"request_id": c.Writer.Header().Get("X-Request-ID"),
		})
		return
	}

	emailNormalized := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := h.repo.FindUserByEmail(c.Request.Context(), emailNormalized)
	if err != nil || !user.IsActive {
		h.unauthorized(c)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.unauthorized(c)
		return
	}

	memberships, err := h.repo.ListMembershipsByUserID(c.Request.Context(), user.ID)
	if err != nil {
		h.internalError(c)
		return
	}

	accessToken, _, err := h.tokens.NewAccessToken(user.ID, user.Email)
	if err != nil {
		h.internalError(c)
		return
	}

	rawRefresh, refreshHash, refreshExpiresAt, err := h.tokens.NewRefreshToken()
	if err != nil {
		h.internalError(c)
		return
	}

	err = h.repo.CreateRefreshToken(c.Request.Context(), RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: refreshExpiresAt,
	}, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		h.internalError(c)
		return
	}

	h.setRefreshCookie(c, rawRefresh, refreshExpiresAt)
	c.JSON(http.StatusOK, h.toAuthResponse(user, memberships, accessToken))
}

func (h *Handler) refresh(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshCookieName)
	if err != nil || strings.TrimSpace(rawRefresh) == "" {
		h.unauthorized(c)
		return
	}

	refreshHash := h.tokens.HashRefreshToken(rawRefresh)
	oldToken, user, err := h.repo.FindActiveRefreshToken(c.Request.Context(), refreshHash)
	if err != nil {
		h.unauthorized(c)
		return
	}

	rawReplacement, replacementHash, replacementExpiresAt, err := h.tokens.NewRefreshToken()
	if err != nil {
		h.internalError(c)
		return
	}

	replacement := RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: replacementHash,
		ExpiresAt: replacementExpiresAt,
	}

	if err := h.repo.RotateRefreshToken(c.Request.Context(), oldToken.ID, replacement, c.Request.UserAgent(), c.ClientIP()); err != nil {
		h.internalError(c)
		return
	}

	memberships, err := h.repo.ListMembershipsByUserID(c.Request.Context(), user.ID)
	if err != nil {
		h.internalError(c)
		return
	}

	accessToken, _, err := h.tokens.NewAccessToken(user.ID, user.Email)
	if err != nil {
		h.internalError(c)
		return
	}

	h.setRefreshCookie(c, rawReplacement, replacementExpiresAt)
	c.JSON(http.StatusOK, h.toAuthResponse(user, memberships, accessToken))
}

func (h *Handler) logout(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshCookieName)
	if err == nil && strings.TrimSpace(rawRefresh) != "" {
		_ = h.repo.RevokeByTokenHash(c.Request.Context(), h.tokens.HashRefreshToken(rawRefresh))
	}

	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) setRefreshCookie(c *gin.Context, value string, expiresAt time.Time) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		refreshCookieName,
		value,
		int(time.Until(expiresAt).Seconds()),
		"/",
		"",
		h.cfg.AppEnv == "prod",
		true,
	)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, "", -1, "/", "", h.cfg.AppEnv == "prod", true)
}

func (h *Handler) toAuthResponse(user User, memberships []Membership, accessToken string) authResponse {
	out := make([]membershipResponse, 0, len(memberships))
	for _, m := range memberships {
		out = append(out, membershipResponse{
			TenantID: m.TenantID.String(),
			BranchID: m.BranchID.String(),
			Role:     m.Role,
		})
	}

	return authResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   h.tokens.AccessTTLSeconds(),
		User: userResponse{
			ID:    user.ID.String(),
			Email: user.Email,
		},
		Memberships: out,
	}
}

func (h *Handler) unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":       "unauthorized",
		"message":    "Invalid credentials or session.",
		"request_id": c.Writer.Header().Get("X-Request-ID"),
	})
}

func (h *Handler) internalError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"code":       "internal_error",
		"message":    "Something went wrong.",
		"request_id": c.Writer.Header().Get("X-Request-ID"),
	})
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
