package handler

import (
	"net/http"
	"strings"

	"cold-plasma-server/internal/service"
	"cold-plasma-server/internal/transport/middleware"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth          *service.AuthService
	verify        *service.EmailVerificationService
	phoneVerify   *service.PhoneVerificationService
	passwordReset *service.PasswordResetService
	vk            *service.VKAuthService
}

func NewAuthHandler(auth *service.AuthService, verify *service.EmailVerificationService, phoneVerify *service.PhoneVerificationService, passwordReset *service.PasswordResetService, vk *service.VKAuthService) *AuthHandler {
	return &AuthHandler{auth: auth, verify: verify, phoneVerify: phoneVerify, passwordReset: passwordReset, vk: vk}
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	LastName string `json:"last_name"`
	Surname  string `json:"surname"`
	Phone    string `json:"phone"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	firstName := strings.TrimSpace(req.Name)
	lastName := strings.TrimSpace(req.LastName)
	if lastName == "" {
		lastName = strings.TrimSpace(req.Surname)
	}
	if firstName == "" || lastName == "" {
		fail(c, http.StatusBadRequest, "Имя и фамилия обязательны")
		return
	}
	fullName := strings.Join([]string{firstName, lastName}, " ")
	u, sent, err := h.auth.Register(c.Request.Context(), req.Email, req.Password, fullName, req.Phone)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	created(c, gin.H{"user": u, "verification_required": true, "verification_sent": sent})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	u, token, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"user": u, "token": token})
}

func (h *AuthHandler) Me(c *gin.Context) {
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}

	u, err := h.auth.Me(c.Request.Context(), uCtx.UserID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, u)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	if h.verify == nil {
		fail(c, http.StatusBadRequest, "Верификация email не настроена")
		return
	}
	token := c.Query("token")
	if err := h.verify.VerifyToken(c.Request.Context(), token); err != nil {
		// Для ссылки из письма удобнее просто редиректнуть с флагом ошибки
		c.Redirect(http.StatusFound, "/account?verified=0")
		return
	}
	c.Redirect(http.StatusFound, "/account?verified=1")
}

type resendReq struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	if h.verify == nil {
		fail(c, http.StatusBadRequest, "Верификация email не настроена")
		return
	}
	var req resendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	_, _ = h.verify.SendByEmail(c.Request.Context(), req.Email)
	ok(c, gin.H{"ok": true})
}

type forgotPasswordReq struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	if h.passwordReset == nil {
		fail(c, http.StatusBadRequest, "Восстановление пароля не настроено")
		return
	}
	var req forgotPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	if err := h.passwordReset.Request(c.Request.Context(), req.Email); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

type resetPasswordReq struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	if h.passwordReset == nil {
		fail(c, http.StatusBadRequest, "Восстановление пароля не настроено")
		return
	}
	var req resetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	if err := h.passwordReset.Reset(c.Request.Context(), req.Email, req.Token, req.Password); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

func (h *AuthHandler) SendPhoneCode(c *gin.Context) {
	if h.phoneVerify == nil {
		fail(c, http.StatusBadRequest, "Подтверждение телефона не настроено")
		return
	}
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}
	phone, err := h.phoneVerify.SendCode(c.Request.Context(), uCtx.UserID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"ok": true, "phone": phone})
}

type updatePhoneReq struct {
	Phone string `json:"phone"`
}

func (h *AuthHandler) UpdatePhone(c *gin.Context) {
	if h.phoneVerify == nil {
		fail(c, http.StatusBadRequest, "Подтверждение телефона не настроено")
		return
	}
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}
	var req updatePhoneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	u, err := h.phoneVerify.UpdatePhone(c.Request.Context(), uCtx.UserID, req.Phone)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"user": u})
}

type verifyPhoneReq struct {
	Code string `json:"code"`
}

func (h *AuthHandler) VerifyPhone(c *gin.Context) {
	if h.phoneVerify == nil {
		fail(c, http.StatusBadRequest, "Подтверждение телефона не настроено")
		return
	}
	uCtx, okUser := middleware.GetUser(c)
	if !okUser {
		fail(c, http.StatusUnauthorized, "Нужна авторизация")
		return
	}
	var req verifyPhoneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	u, err := h.phoneVerify.Verify(c.Request.Context(), uCtx.UserID, req.Code)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"user": u})
}

type vkExchangeReq struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	DeviceID     string `json:"device_id"`
	RedirectURI  string `json:"redirect_uri"`
}

func (h *AuthHandler) VKExchange(c *gin.Context) {
	if h.vk == nil {
		fail(c, http.StatusBadRequest, "VK ID не настроен")
		return
	}
	var req vkExchangeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	u, token, err := h.vk.Exchange(c.Request.Context(), service.VKExchangeParams{
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		DeviceID:     req.DeviceID,
		RedirectURI:  req.RedirectURI,
	})
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"user": u, "token": token})
}
