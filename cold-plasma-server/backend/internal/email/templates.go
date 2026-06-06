package email

import (
	"fmt"
	"strings"
)

// EmailTemplate содержит текстовую и HTML версии письма
type EmailTemplate struct {
	Subject  string
	TextBody string
	HTMLBody string
}

// VerificationEmailTemplate создает шаблон письма подтверждения регистрации
func VerificationEmailTemplate(verifyURL, salonName string) EmailTemplate {
	if salonName == "" {
		salonName = "Холодная плазма"
	}

	subject := "Подтверждение регистрации"

	text := fmt.Sprintf(`Здравствуйте!

Спасибо за регистрацию в %s!

Для завершения регистрации подтвердите ваш email, перейдя по ссылке:
%s

Если это были не вы — просто игнорируйте это письмо.

С уважением,
Команда %s`, salonName, verifyURL, salonName)

	html := buildHTMLEmail(
		"Подтвердите регистрацию",
		fmt.Sprintf(`
			<p style="font-size: 16px; line-height: 1.6; color: #333333; margin: 0 0 20px 0;">
				Спасибо за регистрацию в <strong>%s</strong>!
			</p>
			<p style="font-size: 16px; line-height: 1.6; color: #333333; margin: 0 0 30px 0;">
				Для завершения регистрации подтвердите ваш email:
			</p>
			<table cellpadding="0" cellspacing="0" border="0" style="margin: 0 0 30px 0;">
				<tr>
					<td style="border-radius: 6px; background-color: #7C3AED;">
						<a href="%s" target="_blank" style="display: inline-block; padding: 16px 40px; font-size: 16px; font-weight: 600; color: #ffffff; text-decoration: none; border-radius: 6px;">
							Подтвердить email
						</a>
					</td>
				</tr>
			</table>
			<p style="font-size: 14px; line-height: 1.6; color: #666666; margin: 0;">
				Если кнопка не работает, скопируйте и вставьте эту ссылку в браузер:<br>
				<a href="%s" style="color: #7C3AED; word-break: break-all;">%s</a>
			</p>
			<p style="font-size: 14px; line-height: 1.6; color: #999999; margin: 30px 0 0 0;">
				Если это были не вы — просто игнорируйте это письмо.
			</p>
		`, salonName, verifyURL, verifyURL, verifyURL),
		salonName,
	)

	return EmailTemplate{
		Subject:  subject,
		TextBody: text,
		HTMLBody: html,
	}
}

// PasswordResetEmailTemplate создает шаблон письма восстановления пароля
func PasswordResetEmailTemplate(resetURL, salonName string) EmailTemplate {
	if salonName == "" {
		salonName = "Холодная плазма"
	}

	subject := "Восстановление пароля"

	text := fmt.Sprintf(`Здравствуйте!

Вы запросили восстановление пароля для вашего аккаунта в %s.

Для сброса пароля перейдите по ссылке:
%s

Ссылка действительна в течение 30 минут.

Если это были не вы — просто игнорируйте это письмо, ваш пароль останется без изменений.

С уважением,
Команда %s`, salonName, resetURL, salonName)

	html := buildHTMLEmail(
		"Восстановление пароля",
		fmt.Sprintf(`
			<p style="font-size: 16px; line-height: 1.6; color: #333333; margin: 0 0 20px 0;">
				Вы запросили восстановление пароля для вашего аккаунта.
			</p>
			<p style="font-size: 16px; line-height: 1.6; color: #333333; margin: 0 0 30px 0;">
				Для сброса пароля нажмите на кнопку ниже:
			</p>
			<table cellpadding="0" cellspacing="0" border="0" style="margin: 0 0 30px 0;">
				<tr>
					<td style="border-radius: 6px; background-color: #7C3AED;">
						<a href="%s" target="_blank" style="display: inline-block; padding: 16px 40px; font-size: 16px; font-weight: 600; color: #ffffff; text-decoration: none; border-radius: 6px;">
							Сбросить пароль
						</a>
					</td>
				</tr>
			</table>
			<p style="font-size: 14px; line-height: 1.6; color: #666666; margin: 0 0 20px 0;">
				Если кнопка не работает, скопируйте и вставьте эту ссылку в браузер:<br>
				<a href="%s" style="color: #7C3AED; word-break: break-all;">%s</a>
			</p>
			<p style="font-size: 14px; line-height: 1.6; color: #DC2626; margin: 0 0 20px 0;">
				⏱️ Ссылка действительна в течение 30 минут.
			</p>
			<p style="font-size: 14px; line-height: 1.6; color: #999999; margin: 0;">
				Если это были не вы — просто игнорируйте это письмо, ваш пароль останется без изменений.
			</p>
		`, resetURL, resetURL, resetURL),
		salonName,
	)

	return EmailTemplate{
		Subject:  subject,
		TextBody: text,
		HTMLBody: html,
	}
}

// buildHTMLEmail создает полный HTML email с общим layout
func buildHTMLEmail(title, content, salonName string) string {
	if salonName == "" {
		salonName = "Холодная плазма"
	}

	// Экранируем HTML в названии салона
	salonName = strings.ReplaceAll(salonName, "<", "&lt;")
	salonName = strings.ReplaceAll(salonName, ">", "&gt;")

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<title>%s</title>
	<!--[if mso]>
	<style type="text/css">
		body, table, td {font-family: Arial, Helvetica, sans-serif !important;}
	</style>
	<![endif]-->
</head>
<body style="margin: 0; padding: 0; background-color: #f5f5f5; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
	<table cellpadding="0" cellspacing="0" border="0" width="100%%" style="background-color: #f5f5f5; padding: 40px 20px;">
		<tr>
			<td align="center">
				<!-- Основной контейнер -->
				<table cellpadding="0" cellspacing="0" border="0" width="600" style="max-width: 600px; background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
					<!-- Шапка -->
					<tr>
						<td style="background: linear-gradient(135deg, #7C3AED 0%%, #A78BFA 100%%); padding: 40px 40px 30px 40px; border-radius: 12px 12px 0 0; text-align: center;">
							<h1 style="margin: 0; font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px;">
								%s
							</h1>
						</td>
					</tr>

					<!-- Контент -->
					<tr>
						<td style="padding: 40px;">
							<h2 style="margin: 0 0 20px 0; font-size: 22px; font-weight: 600; color: #1f2937;">
								%s
							</h2>
							%s
						</td>
					</tr>

					<!-- Футер -->
					<tr>
						<td style="padding: 30px 40px; background-color: #f9fafb; border-radius: 0 0 12px 12px; border-top: 1px solid #e5e7eb;">
							<p style="margin: 0 0 10px 0; font-size: 14px; line-height: 1.6; color: #6b7280; text-align: center;">
								С уважением,<br>
								<strong style="color: #374151;">%s</strong>
							</p>
							<p style="margin: 10px 0 0 0; font-size: 12px; line-height: 1.5; color: #9ca3af; text-align: center;">
								Это автоматическое письмо, пожалуйста, не отвечайте на него.
							</p>
						</td>
					</tr>
				</table>

				<!-- Дополнительная информация -->
				<table cellpadding="0" cellspacing="0" border="0" width="600" style="max-width: 600px; margin-top: 20px;">
					<tr>
						<td style="padding: 0 20px; text-align: center;">
							<p style="margin: 0; font-size: 12px; line-height: 1.5; color: #9ca3af;">
								© 2026 %s. Все права защищены.
							</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>`, title, salonName, title, content, salonName, salonName)
}
