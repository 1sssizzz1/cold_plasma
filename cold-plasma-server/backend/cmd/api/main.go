package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cold-plasma-server/config"
	"cold-plasma-server/internal/ai"
	"cold-plasma-server/internal/email"
	"cold-plasma-server/internal/repository/postgres"
	"cold-plasma-server/internal/security"
	"cold-plasma-server/internal/service"
	"cold-plasma-server/internal/sms"
	tgservice "cold-plasma-server/internal/telegram/service"
	tgtransport "cold-plasma-server/internal/telegram/transport"
	"cold-plasma-server/internal/transport/handler"
	"cold-plasma-server/internal/transport/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config validation: %v", err)
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	// Репозитории
	userRepo := postgres.NewUserRepo(pool)
	procedureRepo := postgres.NewProcedureRepo(pool)
	procedureReviewRepo := postgres.NewProcedureReviewRepo(pool)
	beforeAfterRepo := postgres.NewBeforeAfterRepo(pool)
	bookingRepo := postgres.NewBookingRepo(pool)
	adminNoteRepo := postgres.NewAdminNoteRepo(pool)
	bonusRepo := postgres.NewBonusRepo(pool)
	chatRepo := postgres.NewChatRepo(pool)
	settingsRepo := postgres.NewSettingsRepo(pool)
	txManager := postgres.NewTxManager(pool)

	// Email
	var mailer email.Sender
	if cfg.SMTPHost != "" {
		fromName := cfg.SMTPFromName
		if strings.TrimSpace(fromName) == "" {
			fromName = cfg.SalonName
		}
		mailer = email.NewSMTPSender(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUser,
			cfg.SMTPPassword,
			cfg.SMTPFrom,
			fromName,
			cfg.SMTPImplicitTLS,
			cfg.SMTPInsecureSkipVerifyTLS,
		)
	}
	verifier := service.NewEmailVerificationService(userRepo, mailer, cfg.PublicBaseURL, cfg.EmailVerifySecret, cfg.EmailVerifyExpirySec)
	smsSender := sms.Sender(sms.NewLogSender())
	if strings.EqualFold(cfg.SMSProvider, "http") {
		smsSender = sms.NewHTTPSender(cfg.SMSHTTPURL, cfg.SMSHTTPToken, cfg.SMSFrom)
	}
	chatLogCrypto, err := security.NewTextCipher(cfg.ChatLogEncryptionKey)
	if err != nil {
		log.Fatalf("chat log encryption: %v", err)
	}

	// Сервисы
	authSvc := service.NewAuthService(userRepo, verifier, cfg.JWTSecret, cfg.JWTExpiry)
	phoneVerifySvc := service.NewPhoneVerificationService(userRepo, smsSender, cfg.PhoneCodeTTLSeconds)
	passwordResetSvc := service.NewPasswordResetService(userRepo, mailer, cfg.PublicBaseURL)
	vkAuthSvc := service.NewVKAuthService(userRepo, cfg.VKClientID, cfg.VKClientSecret, cfg.VKRedirectURI, cfg.JWTSecret, cfg.JWTExpiry)
	telegramSvc := tgservice.New(
		bookingRepo,
		userRepo,
		procedureRepo,
		settingsRepo,
		smsSender,
		tgtransport.NewBotClient(cfg.TelegramBotToken, cfg.TelegramAdminChatID),
		tgtransport.NewPollingClient(cfg.TelegramBotToken),
		cfg.TelegramBotUsername,
		cfg.TelegramReminderThread,
		cfg.TelegramRevenueChatID,
		cfg.TelegramRevenueThread,
	)
	procedureSvc := service.NewProcedureService(procedureRepo, procedureReviewRepo)
	beforeAfterSvc := service.NewBeforeAfterService(beforeAfterRepo)
	bookingSvc := service.NewBookingService(procedureRepo, bookingRepo, adminNoteRepo, txManager, telegramSvc)
	bonusSvc := service.NewBonusService(userRepo, bonusRepo, txManager)
	chatSvc := service.NewChatService(cfg, ai.NewProvider(cfg), chatRepo, userRepo, chatLogCrypto)
	pdfSvc := service.NewPDFService(cfg)
	uploadSvc := service.NewUploadService(cfg.UploadDir)
	adminSvc := service.NewAdminService(chatRepo, bookingRepo, adminNoteRepo, settingsRepo, telegramSvc, chatLogCrypto)

	// Handlers
	r := router.New(cfg, router.Handlers{
		Auth:        handler.NewAuthHandler(authSvc, verifier, phoneVerifySvc, passwordResetSvc, vkAuthSvc),
		Procedures:  handler.NewProcedureHandler(procedureSvc),
		BeforeAfter: handler.NewBeforeAfterHandler(beforeAfterSvc),
		Bookings:    handler.NewBookingHandler(bookingSvc),
		Bonus:       handler.NewBonusHandler(bonusSvc),
		Chat:        handler.NewChatHandler(chatSvc),
		PDF:         handler.NewPDFHandler(pdfSvc),
		Admin:       handler.NewAdminHandler(adminSvc, bonusSvc),
		Telegram:    handler.NewTelegramHandler(telegramSvc, cfg.TelegramWebhookSecret),
		Uploads:     handler.NewUploadHandler(uploadSvc),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      25 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("API listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Запускаем Telegram polling (альтернатива webhook для обхода блокировок)
	go func() {
		telegramSvc.StartPolling(ctx)
	}()

	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				telegramSvc.SendDueReminders(ctx, time.Now())
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		_ = adminSvc.CleanupOldChatSessions(ctx, time.Now())
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := adminSvc.CleanupOldChatSessions(ctx, time.Now()); err != nil {
					log.Printf("cleanup old ai chat sessions: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Printf("shutdown complete")
}
