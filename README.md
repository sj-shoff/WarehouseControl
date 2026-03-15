# Система управления складом Warehouse Control

**Warehouse Control** — мини-система для управления инвентарем с полным CRUD, аудитом через PostgreSQL-триггеры, ролевым доступом и JWT-аутентификацией через отдельный SSO-сервис (gRPC).  

**Особенности:**
- Refresh tokens (долгоживущие токены обновления)
- Rate limiting (токен-бакет)
- Graceful shutdown
- Автоматическое обновление токена на фронтенде
- История изменений через триггеры (образовательный антипаттерн)

## Архитектура

- **Warehouse Control** — REST API (Gin) + фронтенд (HTML/JS/Tailwind)
- **SSO** — gRPC-сервис аутентификации
- **БД** — PostgreSQL (отдельные БД: `warehouse_control` и `sso`)
- **Фронтенд** — статический, обслуживается Warehouse

## Функции

- CRUD товаров (с ролевыми ограничениями)
- Полная история изменений (INSERT/UPDATE/DELETE) с diff
- Экспорт истории в CSV (с фильтрами)
- Роли: **Admin** / **Manager** / **Viewer**
- JWT + Refresh tokens
- Rate limiting + Prometheus метрики
- Авто-обновление токена на фронтенде

## Установка

### Требования
- Go 1.24+
- Docker + Docker Compose
- Goose (для миграций)

### 1. Клонирование и подготовка
```bash
git clone <repo>
cd warehouse-control-project
cp warehouse-control/.env.example warehouse-control/.env
cp sso/.env.example sso/.env

2. Запуск через Docker (рекомендуется)
Bash# Запуск Postgres + Redis (если нужен)
cd infrastructure && make run   # если есть отдельный infra

# Запуск SSO
cd sso && make run

# Запуск Warehouse
cd warehouse-control && make run
3. Миграции
Bashcd sso && make migrate-up
cd warehouse-control && make migrate-up
4. Создание приложения и пользователей
SQL-- В БД sso выполните:
INSERT INTO apps (id, name, secret) 
VALUES (1, 'warehouse', 'super_secret_jwt_key_change_in_production');
Bash# Создание пользователей
grpcurl -plaintext -d '{"username":"admin","password":"123","role":"admin","app_id":1}' localhost:44044 sso.Auth/Register
grpcurl -plaintext -d '{"username":"manager","password":"123","role":"manager","app_id":1}' localhost:44044 sso.Auth/Register
grpcurl -plaintext -d '{"username":"viewer","password":"123","role":"viewer","app_id":1}' localhost:44044 sso.Auth/Register
Доступ

Warehouse API + UI: http://localhost:8037
SSO gRPC: localhost:44044

Эндпоинты API (Warehouse)
Аутентификация

POST /auth/login → {access_token, refresh_token, username, role, expires_at}
POST /auth/refresh → {access_token, refresh_token}

Товары (требует access_token)

GET /items?limit=10&offset=0&search=...
POST /items (Manager/Admin)
GET /items/:id
PUT /items/:id (Manager/Admin)
DELETE /items/:id (Manager/Admin)
DELETE /items/bulk (только Admin)

История

GET /history?item_id=...&action=...&username=...&date_from=...&date_to=...
GET /history/item/:id
GET /history/export (CSV с фильтрами)

.env файлы
warehouse-control/.env.example — см. выше в предыдущем сообщении.
sso/.env.example — см. выше в предыдущем сообщении.