package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/platform/config"
)

const refreshCookieName = "refresh_token"
const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

type Handler struct {
	repo   *Repository
	tokens *TokenManager
	cfg    config.Config
}

type loginRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	MembershipID string `json:"membership_id"`
}

type switchMembershipRequest struct {
	MembershipID string `json:"membership_id" binding:"required"`
}

type membershipResponse struct {
	MembershipID string `json:"membership_id"`
	TenantID     string `json:"tenant_id"`
	BranchID     string `json:"branch_id"`
	Role         string `json:"role"`
}

type authResponse struct {
	AccessToken          string               `json:"access_token"`
	TokenType            string               `json:"token_type"`
	ExpiresIn            int64                `json:"expires_in_seconds"`
	User                 userResponse         `json:"user"`
	ActiveMembership     membershipResponse   `json:"active_membership"`
	AvailableMemberships []membershipResponse `json:"available_memberships"`
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
	auth.POST("/switch-membership", h.switchMembership)
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

	activeMembership, status, code, message := selectLoginMembership(memberships, req.MembershipID)
	if status != 0 {
		c.AbortWithStatusJSON(status, gin.H{
			"code":       code,
			"message":    message,
			"request_id": c.Writer.Header().Get("X-Request-ID"),
		})
		return
	}

	accessToken, _, err := h.tokens.NewAccessToken(user.ID, user.Email, ScopeClaims{
		MembershipID: activeMembership.ID.String(),
		TenantID:     activeMembership.TenantID.String(),
		BranchID:     activeMembership.BranchID.String(),
		Role:         activeMembership.Role,
	})
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
		ID:           newUUID(),
		UserID:       user.ID,
		MembershipID: activeMembership.ID,
		TokenHash:    refreshHash,
		ExpiresAt:    refreshExpiresAt,
	}, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		h.internalError(c)
		return
	}

	h.setRefreshCookie(c, rawRefresh, refreshExpiresAt)
	h.setCSRFCookie(c, newCSRFToken(), refreshExpiresAt)
	c.JSON(http.StatusOK, h.toAuthResponse(user, memberships, activeMembership, accessToken))
}

func (h *Handler) refresh(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshCookieName)
	if err != nil || strings.TrimSpace(rawRefresh) == "" {
		h.unauthorized(c)
		return
	}

	if !h.validateCSRFFromSessionAction(c) {
		return
	}

	refreshHash := h.tokens.HashRefreshToken(rawRefresh)
	oldToken, user, activeMembership, err := h.repo.FindActiveRefreshToken(c.Request.Context(), refreshHash)
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
		ID:           newUUID(),
		UserID:       user.ID,
		MembershipID: oldToken.MembershipID,
		TokenHash:    replacementHash,
		ExpiresAt:    replacementExpiresAt,
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

	if !containsMembership(memberships, activeMembership.ID) {
		h.unauthorized(c)
		return
	}

	accessToken, _, err := h.tokens.NewAccessToken(user.ID, user.Email, ScopeClaims{
		MembershipID: activeMembership.ID.String(),
		TenantID:     activeMembership.TenantID.String(),
		BranchID:     activeMembership.BranchID.String(),
		Role:         activeMembership.Role,
	})
	if err != nil {
		h.internalError(c)
		return
	}

	h.setRefreshCookie(c, rawReplacement, replacementExpiresAt)
	h.setCSRFCookie(c, newCSRFToken(), replacementExpiresAt)
	c.JSON(http.StatusOK, h.toAuthResponse(user, memberships, activeMembership, accessToken))
}

func (h *Handler) logout(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshCookieName)
	if err == nil && strings.TrimSpace(rawRefresh) != "" {
		if !h.validateCSRFFromSessionAction(c) {
			return
		}

		_ = h.repo.RevokeByTokenHash(c.Request.Context(), h.tokens.HashRefreshToken(rawRefresh))
	}

	h.clearRefreshCookie(c)
	h.clearCSRFCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) switchMembership(c *gin.Context) {
	var req switchMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request payload.",
			"details":    err.Error(),
			"request_id": c.Writer.Header().Get("X-Request-ID"),
		})
		return
	}

	rawRefresh, err := c.Cookie(refreshCookieName)
	if err != nil || strings.TrimSpace(rawRefresh) == "" {
		h.unauthorized(c)
		return
	}

	if !h.validateCSRFFromSessionAction(c) {
		return
	}

	refreshHash := h.tokens.HashRefreshToken(rawRefresh)
	oldToken, user, oldMembership, err := h.repo.FindActiveRefreshToken(c.Request.Context(), refreshHash)
	if err != nil {
		h.unauthorized(c)
		return
	}

	memberships, err := h.repo.ListMembershipsByUserID(c.Request.Context(), user.ID)
	if err != nil {
		h.internalError(c)
		return
	}

	selectedMembership, status, code, message := selectExplicitMembership(memberships, req.MembershipID)
	if status != 0 {
		c.AbortWithStatusJSON(status, gin.H{
			"code":       code,
			"message":    message,
			"request_id": c.Writer.Header().Get("X-Request-ID"),
		})
		return
	}

	rawReplacement, replacementHash, replacementExpiresAt, err := h.tokens.NewRefreshToken()
	if err != nil {
		h.internalError(c)
		return
	}

	replacement := RefreshToken{
		ID:           newUUID(),
		UserID:       user.ID,
		MembershipID: selectedMembership.ID,
		TokenHash:    replacementHash,
		ExpiresAt:    replacementExpiresAt,
	}

	if err := h.repo.RotateRefreshToken(c.Request.Context(), oldToken.ID, replacement, c.Request.UserAgent(), c.ClientIP()); err != nil {
		h.internalError(c)
		return
	}

	if selectedMembership.ID != oldMembership.ID {
		if err := h.repo.CreateScopeSwitchAuditLog(c.Request.Context(), user.ID, oldMembership, selectedMembership, c.Writer.Header().Get("X-Request-ID")); err != nil {
			h.internalError(c)
			return
		}
	}

	accessToken, _, err := h.tokens.NewAccessToken(user.ID, user.Email, ScopeClaims{
		MembershipID: selectedMembership.ID.String(),
		TenantID:     selectedMembership.TenantID.String(),
		BranchID:     selectedMembership.BranchID.String(),
		Role:         selectedMembership.Role,
	})
	if err != nil {
		h.internalError(c)
		return
	}

	h.setRefreshCookie(c, rawReplacement, replacementExpiresAt)
	h.setCSRFCookie(c, newCSRFToken(), replacementExpiresAt)
	c.JSON(http.StatusOK, h.toAuthResponse(user, memberships, selectedMembership, accessToken))
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

func (h *Handler) setCSRFCookie(c *gin.Context, value string, expiresAt time.Time) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		csrfCookieName,
		value,
		int(time.Until(expiresAt).Seconds()),
		"/",
		"",
		h.cfg.AppEnv == "prod",
		false,
	)
}

func (h *Handler) clearCSRFCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(csrfCookieName, "", -1, "/", "", h.cfg.AppEnv == "prod", false)
}

func (h *Handler) toAuthResponse(user User, memberships []Membership, activeMembership Membership, accessToken string) authResponse {
	out := make([]membershipResponse, 0, len(memberships))
	for _, m := range memberships {
		out = append(out, membershipResponse{
			MembershipID: m.ID.String(),
			TenantID:     m.TenantID.String(),
			BranchID:     m.BranchID.String(),
			Role:         m.Role,
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
		ActiveMembership: membershipResponse{
			MembershipID: activeMembership.ID.String(),
			TenantID:     activeMembership.TenantID.String(),
			BranchID:     activeMembership.BranchID.String(),
			Role:         activeMembership.Role,
		},
		AvailableMemberships: out,
	}
}

func (h *Handler) validateCSRFFromSessionAction(c *gin.Context) bool {
	csrfCookie, err := c.Cookie(csrfCookieName)
	if err != nil || strings.TrimSpace(csrfCookie) == "" {
		h.forbiddenScopeSelection(c, "Invalid session action.")
		return false
	}

	csrfHeader := strings.TrimSpace(c.GetHeader(csrfHeaderName))
	if csrfHeader == "" || csrfHeader != csrfCookie {
		h.forbiddenScopeSelection(c, "Invalid session action.")
		return false
	}

	if !isTrustedOriginOrReferer(c) {
		h.forbiddenScopeSelection(c, "Invalid session action.")
		return false
	}

	return true
}

func isTrustedOriginOrReferer(c *gin.Context) bool {
	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		return false
	}

	if origin := strings.TrimSpace(c.GetHeader("Origin")); origin != "" {
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, host)
	}

	if referer := strings.TrimSpace(c.GetHeader("Referer")); referer != "" {
		u, err := url.Parse(referer)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, host)
	}

	return false
}

func selectLoginMembership(memberships []Membership, selectedMembershipID string) (Membership, int, string, string) {
	if len(memberships) == 0 {
		return Membership{}, http.StatusForbidden, "forbidden_scope_selection", "Invalid membership selection."
	}

	selectedMembershipID = strings.TrimSpace(selectedMembershipID)
	if len(memberships) == 1 {
		only := memberships[0]
		if selectedMembershipID == "" {
			return only, 0, "", ""
		}
		selectedID, err := uuid.Parse(selectedMembershipID)
		if err != nil {
			return Membership{}, http.StatusBadRequest, "validation_error", "Invalid membership selection payload."
		}
		if selectedID != only.ID {
			return Membership{}, http.StatusForbidden, "forbidden_scope_selection", "Invalid membership selection."
		}
		return only, 0, "", ""
	}

	if selectedMembershipID == "" {
		return Membership{}, http.StatusBadRequest, "validation_error", "Membership selection is required."
	}

	return selectExplicitMembership(memberships, selectedMembershipID)
}

func selectExplicitMembership(memberships []Membership, selectedMembershipID string) (Membership, int, string, string) {
	selectedID, err := uuid.Parse(strings.TrimSpace(selectedMembershipID))
	if err != nil {
		return Membership{}, http.StatusBadRequest, "validation_error", "Invalid membership selection payload."
	}

	for _, m := range memberships {
		if m.ID == selectedID {
			return m, 0, "", ""
		}
	}

	return Membership{}, http.StatusForbidden, "forbidden_scope_selection", "Invalid membership selection."
}

func containsMembership(memberships []Membership, membershipID uuid.UUID) bool {
	for _, m := range memberships {
		if m.ID == membershipID {
			return true
		}
	}
	return false
}

func newCSRFToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return uuid.NewString()
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

func newUUID() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
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

func (h *Handler) forbiddenScopeSelection(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"code":       "forbidden_scope_selection",
		"message":    message,
		"request_id": c.Writer.Header().Get("X-Request-ID"),
	})
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
