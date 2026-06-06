# Cold Plasma — система онлайн-записи для бизнеса

Fullstack-сервис онлайн-записи клиентов, работающий в продакшене.
Backend на Go, SPA на React + TypeScript, PostgreSQL. Спроектирован, написан и
развёрнут в одиночку.

🔗 **Живой проект:** [plasmaglow.ru](https://plasmaglow.ru)

> Репозиторий открыт для ознакомления с кодом и архитектурой.
> Это рабочий коммерческий продукт — инструкции по развёртыванию и `.env`
> намеренно не публикуются.

---

## Что это

Платформа, через которую клиенты записываются на процедуры онлайн, подтверждают
телефон по SMS, общаются с AI-консультантом и копят бонусы, а администратор
получает уведомления о новых заявках в Telegram и управляет всем из админ-панели.

Система закрывает реальную задачу бизнеса: убрать запись «по звонку»,
автоматизировать уведомления и квалификацию заявок (тип лида: `cold` / `warm` / `hot`).

## Ключевые возможности

- **Аутентификация:** регистрация, вход по JWT, восстановление пароля, верификация email
- **Телефон:** обязательная привязка российского номера, нормализация к `+7XXXXXXXXXX`,
  подтверждение по SMS-коду, запрет повторной привязки одного номера
- **Онлайн-запись:** выбор процедуры, нескольких вариантов даты/времени, комментарий мастеру
- **Бонусная система:** баланс и история бонусов в личном кабинете, ручное начисление и списание
- **Telegram-бот:** уведомления о новых записях, напоминания, отчёт по выручке;
  трёхслойная архитектура (transport / service / repository)
- **AI-консультант:** чат-помощник на базе LLM (Gemini по умолчанию, OpenAI опционально)
  с фолбэком моделей и шифрованием истории переписки
- **Авторизация через VK ID**
- **Админ-панель:** управление записями и бонусами, каналы уведомлений (Telegram / SMS);
  страница `/admin` скрыта от не-администраторов как 404
- **PDF-памятка** по уходу после процедуры (генерация на лету)

## Технологии

**Backend**
- Go 1.25, [Gin](https://github.com/gin-gonic/gin)
- PostgreSQL ([pgx/v5](https://github.com/jackc/pgx)), SQL-миграции
- JWT ([golang-jwt/v5](https://github.com/golang-jwt/jwt)), bcrypt (`golang.org/x/crypto`)
- Структурное логирование (`slog`), unit-тесты (`go test`)
- Генерация PDF ([gofpdf](https://github.com/jung-kurt/gofpdf))

**Frontend**
- React 18, TypeScript, Vite
- Tailwind CSS, Zustand, Framer Motion

**Инфраструктура**
- Docker Compose, Makefile, nginx
- Деплой на российский хостинг, домен + HTTPS

**Внешние интеграции**
- Telegram Bot API · VK ID · SMS-провайдер (через адаптер) · SMTP · LLM (Gemini / OpenAI)

## Архитектура

Чистая слоёная архитектура с разделением ответственности
(`transport` → `service` → `repository`):

```
cold-plasma-server/backend/
├── cmd/api/                 точка входа
├── config/                  конфигурация из окружения
├── migrations/              SQL-миграции схемы БД
├── internal/
│   ├── transport/
│   │   ├── router/          маршруты /api/v1 (+ группы auth, admin)
│   │   ├── handler/         HTTP-обработчики
│   │   └── middleware/      auth, проверка прав
│   ├── service/             бизнес-логика (auth, запись, бонусы, верификация)
│   ├── repository/postgres/ слой доступа к данным
│   ├── models/              доменные модели
│   ├── security/            шифрование чувствительных данных
│   ├── ai/                  интеграция с LLM (Gemini / OpenAI)
│   ├── email/               отправка писем (SMTP)
│   ├── sms/                 отправка SMS (адаптер под провайдера)
│   └── telegram/            бот: transport / service / repository
└── pkg/
    ├── jwt/                 выпуск и валидация токенов
    └── utils/
```

Принципы:
- внешние сервисы (SMS, Telegram, AI, email) спрятаны за интерфейсами —
  провайдера можно заменить, не трогая бизнес-логику;
- секреты — только из переменных окружения, в коде их нет;
- покрытие тестами критичных участков (шифрование, нормализация телефона, AI-клиент).

## API (обзор)

REST, базовый префикс `/api/v1`. Часть поверхности — как иллюстрация объёма:

**Публичные**
```
GET  /healthz
GET  /api/v1/procedures            POST /api/v1/chat
GET  /api/v1/pdf/care-memo
POST /api/v1/auth/register         POST /api/v1/auth/login
POST /api/v1/auth/forgot-password  POST /api/v1/auth/reset-password
POST /api/v1/auth/vk/exchange      GET  /api/v1/auth/verify
```

**С JWT**
```
GET  /api/v1/me
POST /api/v1/auth/phone/send-code  POST /api/v1/auth/phone/verify
POST /api/v1/bookings              GET  /api/v1/bookings
GET  /api/v1/bonus/balance         GET  /api/v1/bonus/logs
POST /api/v1/bonus/spend
```

**Админские**
```
GET  /api/v1/admin/chat-logs
```

## Моя роль

Проект разработан и доведён до продакшена самостоятельно: проектирование схемы БД и
архитектуры, весь backend на Go, фронтенд на React + TypeScript, интеграции
(Telegram, VK, SMS, email, LLM), а также деплой, домен и HTTPS.
