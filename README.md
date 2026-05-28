# Онлайн-кинотеатр

Бэкенд онлайн-кинотеатра на Go с микросервисной архитектурой. Межсервисное взаимодействие через gRPC.

> **Статус:** в разработке. Реализованы sso, showcase и media сервисы. Планируется сервис транскодирования видео (ffmpeg).
> Внешнее API - через gateway на базе KrakenD.

### Сервисы

**sso** - авторизация и аутентификация
- Регистрация и вход
- JWT-токены, сессии в Redis
- Reset-токены для восстановления пароля

**showcase** - витрина контента
- Каталог фильмов с метаданными и постерами
- PostgreSQL

**media** - доставка контента
- Генерация presigned URL для медиафайлов
- Хранилище на базе SeaweedFS (S3-совместимый API)

**transcoder** *(в планах)*
- Транскодирование видео в необходимые форматы через ffmpeg

## Стек

- **Go 1.24**
- **gRPC** + protobuf (buf) - межсервисное взаимодействие
- **PostgreSQL** (pgx) - основное хранилище
- **Redis** - сессии и reset-токены
- **SeaweedFS** - объектное хранилище файлов (S3 API)
- **goose** - миграции базы данных
- **JWT** - аутентификация
- **Docker Compose** - локальный запуск окружения
- **mockery** - моки для тестов

## Запуск

**Требования:** Go 1.24+, Docker, Docker Compose

```bash
git clone https://github.com/IlyaYashkin/cinema.git
cd cinema

cp .env.example .env # Заполнить .env актуальными данными

docker compose up -d

goose up

go run ./cmd/sso/main.go
go run ./cmd/showcase/main.go
go run ./cmd/media/main.go
```

### Генерация кода из proto

```bash
make proto
```

## Лицензия

MIT