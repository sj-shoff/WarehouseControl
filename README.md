# Warehouse Control System

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)

**Warehouse Control System** — это система управления складским учётом с полным аудитом изменений и ролевой моделью доступа. Сервис предоставляет HTTP API и встроенный SPA-фронтенд, а аутентификация делегируется отдельному SSO-сервису.

---

## Важное предупреждение

Система **полностью зависит от SSO-сервиса**.  
Перед запуском Warehouse Control убедитесь, что SSO запущен и готов к работе:

```bash
cd sso
make docker-up
make migrate-up
```

SSO должен быть доступен по адресу, указанному в `SSO_GRPC_ADDR` (по умолчанию `sso:44044` для Docker).  
Подробная инструкция по запуску SSO находится в [`sso/README.md`](sso/README.md).

---

## Возможности

- **Управление товарами** — создание, чтение, обновление, удаление (в т.ч. массовое).
- **Полный аудит** — каждое изменение фиксируется в таблице `items_history` через триггеры PostgreSQL (включая `INSERT`, `UPDATE`, `DELETE`).
- **Визуальное сравнение версий** — интерфейс показывает изменения полей с подсветкой.
- **Ролевая модель**:
  - `admin` — полный доступ, управление пользователями;
  - `manager` — создание и редактирование товаров;
  - `viewer` — только просмотр.
- **Экспорт истории** в CSV.
- **Поиск** товаров по названию и SKU.
- **Индикация низких остатков** (quantity < 5).
- **Rate limiting** (5 запросов/сек на IP).
- **JWT-аутентификация** через SSO + автоматический refresh токенов.

---

## Архитектура

### База данных

Две основные таблицы:

- **`items`** — товары (id, name, sku, quantity, price, category, location, created_at, updated_at).
- **`items_history`** — аудит (id, item_id, action, old_data (jsonb), new_data (jsonb), changed_by, changed_at).

**Триггер `log_item_changes`** автоматически заполняет `items_history` при любых изменениях в `items`.  
Имя пользователя (`changed_by`) передаётся через сессионную переменную `warehouse_control.changed_by`, которую устанавливает приложение перед транзакцией.

> 💡 Использование триггеров для аудита — учебный приём, демонстрирующий альтернативный подход к логированию. В высоконагруженных системах такая логика часто выносится на уровень приложения.

### Фронтенд

SPA на чистом JavaScript, стилизация через TailwindCSS (CDN).  

### Коммуникация с SSO

- При старте сервис регистрируется в SSO (bootstrap).
- Аутентификация проксируется через gRPC-клиент.
- JWT проверяются локально через middleware с общим секретом (`JWT_SECRET`).

---

## Быстрый старт

### Требования

- Go 1.24+
- Docker & Docker Compose
- Запущенный SSO-сервис (см. выше)

### Шаги

1. **Клонируйте репозиторий** (если ещё не сделали этого).
2. **Настройте окружение**:
   ```bash
   cp .env.example .env
   ```
   (Отредактируйте, особенно важны `SSO_GRPC_ADDR` и `JWT_SECRET`.)
3. **Запустите сервис и БД**:
   ```bash
   make docker-up
   ```
4. **Примените миграции**:
   ```bash
   make migrate-up
   ```
5. **Откройте браузер** по адресу [http://localhost:8037](http://localhost:8037).

---

## Конфигурация

Основные переменные окружения (полный список см. в [`.env.example`](warehouse-control/.env.example)):

- `SERVER_PORT` — порт HTTP-сервера (по умолчанию `8037`).
- `SERVER_READ_TIMEOUT` — таймаут чтения запроса (по умолчанию `30s`).
- `SERVER_WRITE_TIMEOUT` — таймаут записи ответа (по умолчанию `30s`).
- `SERVER_IDLE_TIMEOUT` — таймаут неактивного соединения (по умолчанию `60s`).
- `SERVER_SHUTDOWN_TIMEOUT` — таймаут корректного завершения (по умолчанию `10s`).
- `POSTGRES_HOST` — хост базы данных склада (для Docker — `warehouse-db`).
- `POSTGRES_PORT` — порт базы данных внутри контейнера (по умолчанию `5432`).
- `POSTGRES_EXTERNAL_PORT` — порт базы данных для внешнего подключения (используется для миграций, по умолчанию `5432`).
- `POSTGRES_USER` — пользователь базы данных (по умолчанию `postgres`).
- `POSTGRES_PASSWORD` — пароль базы данных (обязателен).
- `POSTGRES_DB_WAREHOUSE` — имя базы данных склада (по умолчанию `warehouse_control`).
- `POSTGRES_MAX_OPEN_CONNS` — максимальное количество открытых соединений (по умолчанию `10`).
- `POSTGRES_MAX_IDLE_CONNS` — максимальное количество неактивных соединений (по умолчанию `5`).
- `POSTGRES_CONN_MAX_LIFETIME` — время жизни соединения (по умолчанию `5m`).
- `JWT_SECRET` — секрет для проверки токенов (должен совпадать с SSO).
- `JWT_EXP_HOURS` — время жизни токена в часах (по умолчанию `24`).
- `SSO_GRPC_ADDR` — адрес SSO-сервиса (по умолчанию `localhost:44044` для локальной разработки, для Docker — `sso:44044`).
- `SSO_CLIENT_TIMEOUT` — таймаут запросов к SSO (по умолчанию `10s`).
- `SSO_APP_NAME` — имя приложения для SSO (по умолчанию `warehouse_control`).
- `SSO_APP_ID` — идентификатор приложения в SSO (заполняется автоматически при bootstrap).
- `SSO_APP_SECRET` — секрет приложения (должен совпадать с настроенным в SSO).
- `INIT_ADMIN_USERNAME` — логин администратора для bootstrap SSO.
- `INIT_ADMIN_PASSWORD` — пароль администратора для bootstrap SSO.
- `RETRIES_ATTEMPTS` — количество попыток при ошибке БД (по умолчанию `3`).
- `RETRIES_DELAY_MS` — задержка между попытками в миллисекундах (по умолчанию `100`).
- `RETRIES_BACKOFF` — множитель экспоненциальной задержки (по умолчанию `1.5`).
- `RATE_LIMIT_ENABLED` — включение ограничителя запросов (по умолчанию `true`).
- `RATE_LIMIT_RATE` — количество запросов в секунду (по умолчанию `5`).
- `RATE_LIMIT_CAPACITY` — ёмкость токенов (по умолчанию `10`).

---

## HTTP API

Базовый URL: `http://localhost:8037`  
Все запросы (кроме `/auth/login` и `/auth/refresh`) требуют заголовок:  
`Authorization: Bearer <access_token>`

### Аутентификация

- `POST /auth/login` — прокси в SSO, тело `{username, password}` → возвращает `{access_token, refresh_token}`.
- `POST /auth/refresh` — обновление токенов, тело `{refresh_token}`.
- `POST /auth/register` — регистрация нового сотрудника (только `admin`), тело `{username, password, role}`.

### Товары

- `GET /items?search=&limit=&offset=` — список товаров.
- `POST /items` — создание товара (поля: `name`, `sku`, `quantity`, `price`, `category`, `location`).
- `GET /items/:id` — получить товар.
- `PUT /items/:id` — обновить товар.
- `DELETE /items/:id` — удалить товар (`admin`/`manager`).
- `DELETE /items/bulk` — массовое удаление (`admin`), тело `{ids: [...]}`.

### Аудит

- `GET /history` — глобальная история (параметры: `item_id`, `action`, `username`, `date_from`, `date_to`, `limit`).
- `GET /history/item/:id` — история конкретного товара.
- `GET /history/diff/:id` — сравнение версий (по ID записи истории).
- `GET /history/export` — выгрузка CSV (те же параметры).

### Примеры запросов

**Создание товара:**

```bash
curl -X POST http://localhost:8037/items \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"iPhone 16 Pro","sku":"APPL-IP16P-001","quantity":45,"price":124990.00,"category":"Смартфоны","location":"A-12-03"}'
```

**Получение истории товара:**

```bash
curl http://localhost:8037/history/item/42 -H "Authorization: Bearer YOUR_TOKEN"
```

**Экспорт истории в CSV:**

```bash
curl http://localhost:8037/history/export?limit=10000 -H "Authorization: Bearer YOUR_TOKEN" --output history.csv
```

---

## Безопасность

- **JWT-аутентификация** — middleware проверяет токен, извлекает роль.
- **RBAC** — middleware `RequireRole` ограничивает доступ к эндпоинтам.
- **Rate limiting** — 5 запросов/сек на IP (настраивается).
- **Audit user** — имя пользователя передаётся в БД через `set_config`, триггер записывает его в историю.
- **Валидация** всех входных данных через [validator](https://github.com/go-playground/validator).

---

## Миграции

Управление миграциями осуществляется через [goose](https://github.com/pressly/goose). Файлы лежат в `migrations/`.

- `make migrate-up` — применить все миграции.
- `make migrate-down` — откатить последнюю миграцию.

---

## Устранение неполадок

- **`sso bootstrap failed`** — проверьте, запущен ли SSO, совпадают ли `SSO_GRPC_ADDR` и `SSO_APP_SECRET`. Оба сервиса должны быть в одной сети Docker (`warehouse-net`).
- **`relation items_history does not exist`** — не применены миграции — выполните `make migrate-up`. Убедитесь, что миграции применяются к правильной базе данных `warehouse_control`, а не `sso`.
- **Триггер не записывает `changed_by`** — убедитесь, что в коде репозитория перед выполнением SQL-запроса вызывается `setAuditUser`, который делает `SELECT set_config('warehouse_control.changed_by', ...)`. Проверьте, что пользователь авторизован и данные извлекаются из контекста.
- **Интерфейс не загружается** — откройте консоль браузера (F12), проверьте пути к статике (должны отдаваться корректно по пути `static/templates/index.html`). При локальной разработке убедитесь, что CORS не блокирует запросы (в данной сборке статику отдаёт сам Go-сервер).
- **`invalid token`** — секрет `JWT_SECRET` не совпадает с SSO или рассинхронизировано время на серверах. Проверьте синхронизацию времени.

---

## Команды Makefile

- `make run` — запуск сервиса локально (требуется БД и SSO)
- `make build` — компиляция бинарного файла в `bin/`
- `make docker-up` — запуск контейнеров (сервис + БД)
- `make docker-down` — остановка контейнеров
- `make migrate-up` — применение миграций
- `make migrate-down` — откат последней миграции
- `make lint` — запуск линтера (golangci-lint)

---