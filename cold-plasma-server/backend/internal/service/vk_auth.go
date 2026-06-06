package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
	appjwt "cold-plasma-server/pkg/jwt"
)

type VKAuthService struct {
	users        repository.UserRepository
	clientID     string
	clientSecret string
	redirectURI  string
	jwtSecret    string
	jwtExpiry    int
	httpClient   *http.Client
}

func NewVKAuthService(users repository.UserRepository, clientID, clientSecret, redirectURI, jwtSecret string, jwtExpiry int) *VKAuthService {
	return &VKAuthService{
		users:        users,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		jwtSecret:    jwtSecret,
		jwtExpiry:    jwtExpiry,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

type VKExchangeParams struct {
	Code         string
	CodeVerifier string
	DeviceID     string
	RedirectURI  string
}

func (s *VKAuthService) Exchange(ctx context.Context, p VKExchangeParams) (models.User, string, error) {
	if s.clientID == "" {
		return models.User{}, "", se(ErrValidation, "VK_CLIENT_ID не настроен")
	}
	p.Code = strings.TrimSpace(p.Code)
	p.CodeVerifier = strings.TrimSpace(p.CodeVerifier)
	p.DeviceID = strings.TrimSpace(p.DeviceID)
	redirectURI := strings.TrimSpace(p.RedirectURI)
	if redirectURI == "" {
		redirectURI = s.redirectURI
	}
	if p.Code == "" || p.CodeVerifier == "" || p.DeviceID == "" || redirectURI == "" {
		return models.User{}, "", se(ErrValidation, "Не хватает параметров VK ID")
	}

	token, err := s.exchangeToken(ctx, p, redirectURI)
	if err != nil {
		return models.User{}, "", err
	}
	info, err := s.userInfo(ctx, token.AccessToken)
	if err != nil {
		return models.User{}, "", err
	}
	vkID := info.User.UserID
	if vkID == "" && token.UserID > 0 {
		vkID = fmt.Sprintf("%d", token.UserID)
	}
	if vkID == "" {
		return models.User{}, "", se(ErrValidation, "VK не вернул идентификатор пользователя")
	}

	// Телефон необязателен - пользователь может добавить позже
	phone := ""
	phoneVerified := false
	if info.User.Phone != "" {
		normalizedPhone, err := NormalizeRussianPhone(info.User.Phone)
		if err == nil {
			phone = normalizedPhone
			phoneVerified = true
		}
	}

	email := strings.TrimSpace(strings.ToLower(info.User.Email))
	name := strings.TrimSpace(strings.Join([]string{info.User.FirstName, info.User.LastName}, " "))
	if name == "" {
		name = "VK пользователь"
	}

	u, err := s.users.GetByVKID(ctx, vkID)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return models.User{}, "", err
	}

	if errors.Is(err, models.ErrNotFound) {
		// Если есть телефон, попробуем найти существующий аккаунт
		if phone != "" {
			if existing, findErr := s.users.GetByPhone(ctx, phone); findErr == nil {
				u = existing
				if u.VKID == "" {
					if err := s.users.LinkVK(ctx, u.ID, vkID, email, name); err != nil {
						return models.User{}, "", err
					}
					u, err = s.users.GetByID(ctx, u.ID)
					if err != nil {
						return models.User{}, "", err
					}
				}
			} else if !errors.Is(findErr, models.ErrNotFound) {
				return models.User{}, "", findErr
			}
		}

		// Создаём новый аккаунт, если не нашли существующий
		if u.ID == 0 {
			u, err = s.users.Create(ctx, repository.CreateUserParams{
				Email:         email,
				Name:          name,
				Phone:         phone,
				PasswordHash:  "",
				EmailVerified: email != "",
				PhoneVerified: phoneVerified,
				VKID:          vkID,
				AuthProvider:  "vk",
			})
			if err != nil {
				return models.User{}, "", err
			}
		}
	}

	if u.IsBlocked {
		return models.User{}, "", se(ErrForbidden, "Аккаунт заблокирован")
	}

	// Обновляем телефон, если VK вернул его и он отличается
	if phone != "" && (!u.PhoneVerified || u.Phone != phone) {
		if err := s.users.MarkPhoneVerified(ctx, u.ID, phone); err != nil {
			return models.User{}, "", err
		}
		u, err = s.users.GetByID(ctx, u.ID)
		if err != nil {
			return models.User{}, "", err
		}
	}

	jwtToken, err := appjwt.GenerateToken(int(u.ID), u.Email, u.Name, u.IsAdmin, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return models.User{}, "", fmt.Errorf("jwt: %w", err)
	}
	u.PasswordHash = ""
	return u, jwtToken, nil
}

type vkTokenResp struct {
	AccessToken string `json:"access_token"`
	UserID      int64  `json:"user_id"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func (s *VKAuthService) exchangeToken(ctx context.Context, p VKExchangeParams, redirectURI string) (vkTokenResp, error) {
	body := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     s.clientID,
		"code":          p.Code,
		"code_verifier": p.CodeVerifier,
		"device_id":     p.DeviceID,
		"redirect_uri":  redirectURI,
	}
	if s.clientSecret != "" {
		body["client_secret"] = s.clientSecret
	}
	var out vkTokenResp
	if err := s.postJSON(ctx, "https://id.vk.com/oauth2/auth", body, &out); err != nil {
		fmt.Printf("VK exchangeToken error: %v\n", err)
		return out, err
	}
	if out.Error != "" {
		fmt.Printf("VK API error: %s - %s\n", out.Error, out.ErrorDesc)
		return out, se(ErrValidation, "VK ID: "+firstNonEmptyString(out.ErrorDesc, out.Error))
	}
	if out.AccessToken == "" {
		fmt.Printf("VK returned empty access_token\n")
		return out, se(ErrValidation, "VK ID не вернул access_token")
	}
	return out, nil
}

type vkUserInfoResp struct {
	User struct {
		UserID    string `json:"user_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		Email     string `json:"email"`
	} `json:"user"`
	Error     string `json:"error"`
	ErrorDesc string `json:"error_description"`
}

func (s *VKAuthService) userInfo(ctx context.Context, accessToken string) (vkUserInfoResp, error) {
	body := map[string]string{
		"client_id":    s.clientID,
		"access_token": accessToken,
	}
	var out vkUserInfoResp
	if err := s.postJSON(ctx, "https://id.vk.com/oauth2/user_info", body, &out); err != nil {
		fmt.Printf("VK userInfo error: %v\n", err)
		return out, err
	}
	if out.Error != "" {
		fmt.Printf("VK userInfo API error: %s - %s\n", out.Error, out.ErrorDesc)
		return out, se(ErrValidation, "VK ID: "+firstNonEmptyString(out.ErrorDesc, out.Error))
	}
	fmt.Printf("VK userInfo success: user_id=%s, phone=%s, email=%s\n", out.User.UserID, out.User.Phone, out.User.Email)
	return out, nil
}

func (s *VKAuthService) postJSON(ctx context.Context, endpoint string, payload any, out any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal vk request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("new vk request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		fmt.Printf("VK HTTP request failed to %s: %v\n", endpoint, err)
		return fmt.Errorf("vk request: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Printf("VK response from %s: status=%d, body=%s\n", endpoint, resp.StatusCode, string(b))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return se(ErrValidation, fmt.Sprintf("VK ID вернул ошибку (HTTP %d): %s", resp.StatusCode, string(b)))
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("unmarshal vk response: %w", err)
	}
	return nil
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}
