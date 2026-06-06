package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env                  string
	Port                 string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMODE            string
	JWTSecret            string
	JWTExpiry            int
	ChatLogEncryptionKey string
	PhoneCodeTTLSeconds  int

	CORSAllowedOrigins string

	// AI
	AIProvider           string // gemini|openai
	OpenAIAPIKey         string
	OpenAIModel          string
	OpenAIBaseURL        string
	GeminiAPIKey         string
	GeminiModel          string
	GeminiFallbackModels string

	// Email / верификация
	PublicBaseURL        string
	EmailVerifySecret    string
	EmailVerifyExpirySec int

	SMTPHost                  string
	SMTPPort                  string
	SMTPUser                  string
	SMTPPassword              string
	SMTPFrom                  string
	SMTPFromName              string
	SMTPImplicitTLS           bool
	SMTPInsecureSkipVerifyTLS bool

	SMSProvider  string
	SMSHTTPURL   string
	SMSHTTPToken string
	SMSFrom      string

	VKClientID     string
	VKClientSecret string
	VKRedirectURI  string

	TelegramBotToken       string
	TelegramAdminChatID    string
	TelegramReminderThread string
	TelegramRevenueChatID  string
	TelegramRevenueThread  string
	TelegramBotUsername    string
	TelegramWebhookSecret  string
	UploadDir              string

	// Локальный контекст салона (используется в системном промпте)
	SalonName       string
	SalonAddress    string
	SalonWorkHours  string
	SalonDirections string
	SalonParking    string
	SalonPopularFAQ string

	PDFFontPath string
}

func Load() (*Config, error) {
	jwtExpiry, err := strconv.Atoi(getenv("JWT_EXPIRY", "86400"))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_EXPIRY: %w", err)
	}

	emailVerifyExpiry, err := strconv.Atoi(getenv("EMAIL_VERIFY_EXPIRY", "86400"))
	if err != nil {
		return nil, fmt.Errorf("parse EMAIL_VERIFY_EXPIRY: %w", err)
	}
	phoneCodeTTL, err := strconv.Atoi(getenv("PHONE_CODE_TTL_SECONDS", "300"))
	if err != nil {
		return nil, fmt.Errorf("parse PHONE_CODE_TTL_SECONDS: %w", err)
	}

	smtpImplicitTLS, err := strconv.ParseBool(getenv("SMTP_IMPLICIT_TLS", "false"))
	if err != nil {
		return nil, fmt.Errorf("parse SMTP_IMPLICIT_TLS: %w", err)
	}
	smtpInsecure, err := strconv.ParseBool(getenv("SMTP_TLS_INSECURE", "false"))
	if err != nil {
		return nil, fmt.Errorf("parse SMTP_TLS_INSECURE: %w", err)
	}

	return &Config{
		Env:                  getenv("APP_ENV", "development"),
		Port:                 getenv("PORT", "8080"),
		DBHost:               getenv("DB_HOST", "db"),
		DBPort:               getenv("DB_PORT", "5432"),
		DBUser:               getenv("DB_USER", "postgres"),
		DBPassword:           getenv("DB_PASSWORD", "postgres"),
		DBName:               getenv("DB_NAME", "cold_plasma"),
		DBSSLMODE:            getenv("DB_SSLMODE", "disable"),
		JWTSecret:            getenv("JWT_SECRET", ""),
		JWTExpiry:            jwtExpiry,
		ChatLogEncryptionKey: firstNonEmpty(getenv("CHAT_LOG_ENCRYPTION_KEY", ""), getenv("JWT_SECRET", "")),
		PhoneCodeTTLSeconds:  phoneCodeTTL,
		CORSAllowedOrigins:   getenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:8081"),

		AIProvider:           getenv("AI_PROVIDER", "gemini"),
		OpenAIAPIKey:         getenv("OPENAI_API_KEY", ""),
		OpenAIModel:          getenv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBaseURL:        getenv("OPENAI_BASE_URL", "https://api.openai.com"),
		GeminiAPIKey:         getenv("GEMINI_API_KEY", ""),
		GeminiModel:          getenv("GEMINI_MODEL", "gemini-3-flash-preview"),
		GeminiFallbackModels: getenv("GEMINI_FALLBACK_MODELS", "gemini-3-flash-preview,gemini-2.5-flash,gemini-3.1-flash-lite"),

		PublicBaseURL:        getenv("PUBLIC_BASE_URL", "http://localhost:8081"),
		EmailVerifySecret:    firstNonEmpty(getenv("EMAIL_VERIFY_SECRET", ""), getenv("JWT_SECRET", "")),
		EmailVerifyExpirySec: emailVerifyExpiry,

		SMTPHost:                  getenv("SMTP_HOST", ""),
		SMTPPort:                  getenv("SMTP_PORT", "587"),
		SMTPUser:                  getenv("SMTP_USER", ""),
		SMTPPassword:              getenv("SMTP_PASSWORD", ""),
		SMTPFrom:                  getenv("SMTP_FROM", ""),
		SMTPFromName:              getenv("SMTP_FROM_NAME", ""),
		SMTPImplicitTLS:           smtpImplicitTLS,
		SMTPInsecureSkipVerifyTLS: smtpInsecure,

		SMSProvider:  getenv("SMS_PROVIDER", "log"),
		SMSHTTPURL:   getenv("SMS_HTTP_URL", ""),
		SMSHTTPToken: getenv("SMS_HTTP_TOKEN", ""),
		SMSFrom:      getenv("SMS_FROM", "cold_plasma"),

		VKClientID:     getenv("VK_CLIENT_ID", ""),
		VKClientSecret: getenv("VK_CLIENT_SECRET", ""),
		VKRedirectURI:  getenv("VK_REDIRECT_URI", "http://localhost:5173/account"),

		TelegramBotToken:       getenv("TELEGRAM_BOT_TOKEN", ""),
		TelegramAdminChatID:    getenv("TELEGRAM_ADMIN_CHAT_ID", ""),
		TelegramReminderThread: getenv("TELEGRAM_ADMIN_REMINDER_MESSAGE_THREAD_ID", ""),
		TelegramRevenueChatID:  getenv("TELEGRAM_REVENUE_CHAT_ID", ""),
		TelegramRevenueThread:  getenv("TELEGRAM_REVENUE_MESSAGE_THREAD_ID", ""),
		TelegramBotUsername:    getenv("TELEGRAM_BOT_USERNAME", ""),
		TelegramWebhookSecret:  getenv("TELEGRAM_WEBHOOK_SECRET", ""),
		UploadDir:              getenv("UPLOAD_DIR", "./uploads"),

		SalonName:       getenv("SALON_NAME", "Холодная плазма"),
		SalonAddress:    getenv("SALON_ADDRESS", "г. Северодвинск, (укажите точный адрес в .env)"),
		SalonWorkHours:  getenv("SALON_WORK_HOURS", "Ежедневно 10:00–20:00"),
		SalonDirections: getenv("SALON_DIRECTIONS", "Укажите, как добраться (в .env)."),
		SalonParking:    getenv("SALON_PARKING", "Укажите, где припарковаться (в .env)."),
		SalonPopularFAQ: getenv("SALON_POPULAR_FAQ", "Вопросы жителей района: (укажите в .env, например: «Больно ли?», «Сколько держится эффект?», «Можно ли летом?», «Есть ли противопоказания?»)."),

		PDFFontPath: getenv("PDF_FONT_PATH", "./assets/fonts/Roboto-Regular.ttf"),
	}, nil
}

// IsProduction сообщает, запущено ли приложение в боевом окружении.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(strings.TrimSpace(c.Env), "production") ||
		strings.EqualFold(strings.TrimSpace(c.Env), "prod")
}

// Validate проверяет критичные для безопасности настройки.
// В production пустые/дефолтные секреты приводят к ошибке (fail-fast),
// в dev — к предупреждению с подстановкой небезопасных значений-заглушек.
func (c *Config) Validate() error {
	weakJWT := func(s string) bool {
		s = strings.TrimSpace(s)
		return s == "" || s == "supersecret" || s == "change_me" || len(s) < 16
	}
	weakDBPass := func(s string) bool {
		s = strings.TrimSpace(s)
		return s == "" || s == "postgres"
	}

	if c.IsProduction() {
		if weakJWT(c.JWTSecret) {
			return fmt.Errorf("JWT_SECRET не задан или слишком слабый: задайте случайную строку ≥16 символов")
		}
		if weakDBPass(c.DBPassword) {
			return fmt.Errorf("DB_PASSWORD не задан или дефолтный: задайте надёжный пароль")
		}
		if c.ChatLogEncryptionKey == c.JWTSecret {
			return fmt.Errorf("CHAT_LOG_ENCRYPTION_KEY должен отличаться от JWT_SECRET")
		}
		if strings.TrimSpace(c.TelegramBotToken) != "" && strings.TrimSpace(c.TelegramWebhookSecret) == "" {
			return fmt.Errorf("TELEGRAM_WEBHOOK_SECRET обязателен, когда задан TELEGRAM_BOT_TOKEN")
		}
		return nil
	}

	// dev: не падаем, но громко предупреждаем и подставляем заглушки.
	if weakJWT(c.JWTSecret) {
		log.Printf("[WARN] config: JWT_SECRET слабый/не задан — допустимо только в dev (APP_ENV=%q)", c.Env)
		if strings.TrimSpace(c.JWTSecret) == "" {
			c.JWTSecret = "dev-insecure-secret-change-me"
		}
	}
	if strings.TrimSpace(c.ChatLogEncryptionKey) == "" {
		c.ChatLogEncryptionKey = c.JWTSecret
	}
	if strings.TrimSpace(c.EmailVerifySecret) == "" {
		c.EmailVerifySecret = c.JWTSecret
	}
	return nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMODE)
}

func getenv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}
