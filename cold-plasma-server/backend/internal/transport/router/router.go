package router

import (
	"strings"

	"cold-plasma-server/config"
	"cold-plasma-server/internal/transport/handler"
	"cold-plasma-server/internal/transport/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth        *handler.AuthHandler
	Procedures  *handler.ProcedureHandler
	BeforeAfter *handler.BeforeAfterHandler
	Bookings    *handler.BookingHandler
	Bonus       *handler.BonusHandler
	Chat        *handler.ChatHandler
	PDF         *handler.PDFHandler
	Admin       *handler.AdminHandler
	Telegram    *handler.TelegramHandler
	Uploads     *handler.UploadHandler
}

func New(cfg *config.Config, h Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	allowedOrigins := splitCSV(cfg.CORSAllowedOrigins)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	// Лимитеры по IP для чувствительных к перебору/спаму эндпоинтов.
	// authLimit делится между всеми auth-маршрутами (общий бюджет на IP),
	// smsLimit жёстче — отправка SMS стоит денег.
	// ВНИМАНИЕ: за обратным прокси (nginx) корректный IP зависит от
	// доверенных прокси gin — задайте r.SetTrustedProxies / TrustedPlatform в проде.
	authLimit := middleware.RateLimit(0.5, 15)
	smsLimit := middleware.RateLimit(1.0/60.0, 3)

	// Публичные
	api.GET("/procedures", h.Procedures.List)
	api.GET("/procedures/:id", h.Procedures.Get)
	api.GET("/procedures/:id/reviews", h.Procedures.Reviews)
	api.GET("/before-after-results", h.BeforeAfter.List)
	api.GET("/master-profile", h.Admin.MasterProfile)
	api.POST("/chat", middleware.AuthOptional(cfg.JWTSecret), h.Chat.Ask)
	api.POST("/chat/session", middleware.AuthOptional(cfg.JWTSecret), h.Chat.OpenSession)
	api.POST("/chat/session/close", middleware.AuthOptional(cfg.JWTSecret), h.Chat.CloseSession)
	api.GET("/pdf/care-memo", middleware.AuthOptional(cfg.JWTSecret), h.PDF.CareMemo)
	api.POST("/telegram/webhook", h.Telegram.Webhook)
	api.HEAD("/telegram/webhook", func(c *gin.Context) {
		c.Status(200)
	})

	// Auth
	api.POST("/auth/register", authLimit, h.Auth.Register)
	api.POST("/auth/login", authLimit, h.Auth.Login)
	api.POST("/auth/forgot-password", authLimit, h.Auth.ForgotPassword)
	api.POST("/auth/reset-password", authLimit, h.Auth.ResetPassword)
	api.POST("/auth/vk/exchange", authLimit, h.Auth.VKExchange)
	api.GET("/auth/verify", h.Auth.VerifyEmail)
	api.POST("/auth/resend-verification", authLimit, h.Auth.ResendVerification)

	// Закрытые
	auth := api.Group("/")
	auth.Use(middleware.AuthRequired(cfg.JWTSecret))
	auth.GET("/me", h.Auth.Me)
	auth.POST("/auth/phone/update", h.Auth.UpdatePhone)
	auth.POST("/auth/phone/send-code", smsLimit, h.Auth.SendPhoneCode)
	auth.POST("/auth/phone/verify", h.Auth.VerifyPhone)

	auth.POST("/bookings", h.Bookings.Create)
	auth.GET("/bookings", h.Bookings.ListMy)

	auth.GET("/bonus/balance", h.Bonus.Balance)
	auth.GET("/bonus/logs", h.Bonus.Logs)
	auth.POST("/bonus/spend", h.Bonus.Spend)
	auth.GET("/telegram/status", h.Telegram.Status)
	auth.POST("/telegram/link", h.Telegram.CreateLink)
	auth.POST("/reviews", h.Procedures.CreateReview)

	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.AdminRequired())
	admin.GET("/chat-logs", h.Admin.ChatLogs)
	admin.GET("/chat-sessions", h.Admin.ChatSessions)
	admin.DELETE("/chat-sessions/:id", h.Admin.DeleteChatSession)
	admin.GET("/notification-settings", h.Admin.NotificationSettings)
	admin.POST("/notification-settings", h.Admin.UpdateNotificationSettings)
	admin.GET("/master-profile", h.Admin.MasterProfile)
	admin.POST("/master-profile", h.Admin.UpdateMasterProfile)
	admin.GET("/bonus/user", h.Admin.BonusUserByPhone)
	admin.POST("/bonus/spend", h.Admin.SpendBonus)
	admin.POST("/bonus/award", h.Admin.AwardBonus)
	admin.GET("/procedures", h.Procedures.AdminList)
	admin.POST("/procedures", h.Procedures.AdminCreate)
	admin.PUT("/procedures/:id", h.Procedures.AdminUpdate)
	admin.DELETE("/procedures/:id", h.Procedures.AdminDelete)
	admin.POST("/uploads/procedure-media", h.Uploads.ProcedureMedia)
	admin.GET("/reviews", h.Procedures.AdminReviews)
	admin.DELETE("/reviews/:id", h.Procedures.AdminDeleteReview)
	admin.GET("/before-after-results", h.BeforeAfter.AdminList)
	admin.POST("/before-after-results", h.BeforeAfter.AdminCreate)
	admin.PUT("/before-after-results/:id", h.BeforeAfter.AdminUpdate)
	admin.DELETE("/before-after-results/:id", h.BeforeAfter.AdminDelete)
	admin.GET("/booking-requests", h.Admin.BookingRequests)
	admin.GET("/active-bookings", h.Admin.ActiveBookings)
	admin.GET("/completed-bookings", h.Admin.CompletedBookings)
	admin.POST("/bookings/:id/confirm", h.Admin.ConfirmBooking)
	admin.POST("/bookings/:id/complete", h.Admin.CompleteBooking)
	admin.POST("/bookings/:id/reschedule", h.Admin.RescheduleBooking)
	admin.DELETE("/bookings/:id", h.Admin.DeleteBooking)

	return r
}

func splitCSV(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		out = append(out, "http://localhost:5173")
	}
	return out
}
