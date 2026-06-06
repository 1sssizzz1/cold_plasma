# Холодная плазма — Северодвинск

SPA-сайт салона холодной плазмы с личным кабинетом, онлайн-записью, бонусной системой, email-подтверждением регистрации, PDF-памяткой и AI-консультантом.

## Текущий этап

Проект уже собран как рабочий fullstack-прототип:

- frontend: React SPA с главной страницей, каталогом процедур, записью, страницей "До/После", кабинетом, помощью, чат-виджетом и админ-страницей;
- backend: Go API на Gin со слоями `transport / service / repository`;
- база: PostgreSQL, миграции и стартовые процедуры;
- авторизация: регистрация, вход, JWT, `me`, email-верификация через SMTP/MailHog;
- записи: создание и просмотр своих записей;
- бонусы: баланс, история, списание бонусов;
- AI-чат: Gemini по умолчанию, OpenAI опционально, сохранение логов с шифрованием;
- админка: просмотр логов AI-диалогов для пользователей с `is_admin=true`;
- PDF: генерация памятки по уходу после процедуры;
- инфраструктура: Docker Compose, Makefile, nginx для SPA.

## Стек

- Backend: Go 1.25, Gin, pgx, PostgreSQL, JWT, gofpdf.
- Frontend: React 18, TypeScript, Vite, Tailwind CSS, Zustand, Framer Motion.
- Infra: Docker, docker-compose, Makefile, nginx, MailHog.
- AI: Gemini (`AI_PROVIDER=gemini`) или OpenAI (`AI_PROVIDER=openai`).

## Структура

```text
cold-plasma-server/
  backend/
    cmd/api/                  # точка входа API
    config/                   # env-конфигурация
    internal/
      ai/                     # провайдеры Gemini/OpenAI
      email/                  # SMTP-отправка
      repository/postgres/    # работа с PostgreSQL
      security/               # шифрование текстов чата
      service/                # бизнес-логика
      transport/              # handlers, middleware, router
    migrations/               # SQL-миграции
  frontend/
    src/
      components/             # общие UI-компоненты
      pages/                  # страницы SPA
      store/                  # Zustand stores
      utils/                  # API-клиент
  docker-compose.yml
  Makefile
  .env.example
```

## Быстрый старт через Docker

Скопируйте пример окружения:

```bash
cp .env.example .env
```

Заполните минимум в `.env`:

```env
JWT_SECRET=change_me
CHAT_LOG_ENCRYPTION_KEY=change_me_for_chat_logs

SALON_ADDRESS=г. Северодвинск, точный адрес
SALON_DIRECTIONS=Как добраться: ...
SALON_PARKING=Парковка: ...

AI_PROVIDER=gemini
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-3-flash-preview
```

Поднимите проект:

```bash
make up
```

Адреса после запуска:

- сайт: `http://localhost:8081`;
- API: `http://localhost:8080`;
- healthcheck API: `http://localhost:8080/healthz`;
- MailHog для dev-писем: `http://localhost:8025`;
- админ-страница SPA: `http://localhost:8081/admin`.

Полезные команды:

```bash
make ps        # статус контейнеров
make logs      # логи всех сервисов
make down      # остановить проект
make migrate   # применить миграции отдельно
make build     # пересобрать docker-образы
```

## Локальная разработка

Базу можно оставить в Docker:

```bash
docker compose up -d db mailhog
make migrate
```

API:

```bash
cd backend
go run ./cmd/api
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

Локальный frontend будет доступен на `http://localhost:5173`. Vite проксирует запросы на API через `VITE_API_URL=/api/v1`.

## Переменные окружения

Основные настройки лежат в `.env.example`.

Критичные переменные:

- `JWT_SECRET` - секрет JWT-токенов;
- `CHAT_LOG_ENCRYPTION_KEY` - ключ шифрования входящих и исходящих сообщений AI-чата;
- `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PUBLIC_PORT` - настройки PostgreSQL;
- `CORS_ALLOWED_ORIGINS` - origins для локального frontend и docker-сборки;
- `PUBLIC_BASE_URL` - публичный адрес сайта для ссылок email-подтверждения;
- `EMAIL_VERIFY_SECRET` - секрет токенов email-верификации, если пустой, используется `JWT_SECRET`;
- `SMTP_*` - SMTP-настройки; в Docker по умолчанию используется MailHog;
- `SMS_PROVIDER`, `SMS_HTTP_URL`, `SMS_HTTP_TOKEN`, `SMS_FROM`, `PHONE_CODE_TTL_SECONDS` - подтверждение российского телефона;
- `VK_CLIENT_ID`, `VK_CLIENT_SECRET`, `VK_REDIRECT_URI`, `VITE_VK_CLIENT_ID`, `VITE_VK_REDIRECT_URI` - вход через VK ID;
- `TELEGRAM_BOT_TOKEN`, `TELEGRAM_ADMIN_CHAT_ID` - Telegram-уведомления о новых записях и напоминания за день;
- `TELEGRAM_ADMIN_REMINDER_MESSAGE_THREAD_ID` - тема админ-группы для напоминаний о принятых заявках на сегодня и завтра;
- `TELEGRAM_REVENUE_CHAT_ID`, `TELEGRAM_REVENUE_MESSAGE_THREAD_ID` - отдельный Telegram-чат или тема для отчёта по выручке за месяц;
- `SALON_*` - локальный контекст салона для AI-консультанта и PDF;
- `VITE_API_URL`, `VITE_YM_ID`, `VITE_GA_ID` - параметры сборки frontend.

AI-настройки:

```env
AI_PROVIDER=gemini
GEMINI_API_KEY=
GEMINI_MODEL=gemini-3-flash-preview
GEMINI_FALLBACK_MODELS=gemini-3-flash-preview,gemini-2.5-flash,gemini-3.1-flash-lite

# или OpenAI
AI_PROVIDER=openai
OPENAI_API_KEY=
OPENAI_MODEL=gpt-4o-mini
OPENAI_BASE_URL=https://api.openai.com
```

## Frontend-маршруты

- `/` - главная страница;
- `/procedures` - каталог процедур;
- `/booking` - онлайн-запись;
- `/before-after` - страница "До/После";
- `/account` - личный кабинет, вход и регистрация;
- `/help` - помощь и справочная информация;
- `/admin` - админ-страница логов AI-чата, доступна только администраторам;
- `*` - страница 404.

## API

Базовый префикс: `/api/v1`.

Публичные маршруты:

- `GET /healthz`
- `GET /api/v1/procedures`
- `GET /api/v1/procedures/:id`
- `POST /api/v1/chat`
- `GET /api/v1/pdf/care-memo`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/vk/exchange`
- `GET /api/v1/auth/verify?token=...`
- `POST /api/v1/auth/resend-verification`

Маршруты с JWT:

- `GET /api/v1/me`
- `POST /api/v1/auth/phone/send-code`
- `POST /api/v1/auth/phone/update`
- `POST /api/v1/auth/phone/verify`
- `POST /api/v1/bookings`
- `GET /api/v1/bookings`
- `GET /api/v1/bonus/balance`
- `GET /api/v1/bonus/logs`
- `POST /api/v1/bonus/spend`

Админ-маршруты:

- `GET /api/v1/admin/chat-logs`

Примеры тел запросов:

```json
{
  "email": "client@example.com",
  "password": "secret123",
  "name": "Анна",
  "phone": "+79001234567"
}
```

```json
{
  "procedure_id": 1,
  "datetimes": ["2026-05-14T15:30:00+03:00", "2026-05-15T12:00:00+03:00"],
  "comment": "Хочу консультацию перед процедурой",
  "bonus_used": 0,
  "notify_sms": true,
  "notify_telegram": true
}
```

```json
{
  "message": "Подойдет ли холодная плазма летом?",
  "user_name": "Анна",
  "user_email": "client@example.com",
  "user_phone": "+7...",
  "history": []
}
```

## Админ-доступ

Админка открывается только для авторизованного пользователя с `is_admin=true` в таблице `users`.

Пример SQL для локальной разработки:

```sql
UPDATE users
SET is_admin = true
WHERE email = 'client@example.com';
```

## Проверка

Backend:

```bash
cd backend
go test ./...
```

Frontend:

```bash
cd frontend
npm run build
```

## Заметки по состоянию

- В миграциях есть демо-процедуры для первичного наполнения каталога.
- Создание записи требует подтверждённый российский телефон пользователя.
- Для production нужно заменить `JWT_SECRET`, `CHAT_LOG_ENCRYPTION_KEY`, SMTP-настройки и заполнить точные данные салона.
- Если `CHAT_LOG_ENCRYPTION_KEY` не задан, backend использует `JWT_SECRET`, но отдельный ключ безопаснее.
- Docker-сборка frontend получает `VITE_*` переменные на этапе build, после их изменения нужно пересобрать образ `web`.
