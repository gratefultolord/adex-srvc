# AdEx Service

Мини-сервис маршрутизации auction-запросов к DSP-партнёрам с предварительной фильтрацией (pretargeting), параллельной отправкой запросов и общим таймаутом.

## Что делает сервис

Сервис принимает auction-запрос:

```json
{
  "request_id": "test-auction-1",
  "country": "RU",
  "device_type": "mobile",
  "bid_floor": 1.5,
  "categories": ["news"]
}
```

После этого он:

1. Загружает конфигурации DSP из PostgreSQL.
2. Фильтрует партнёров по правилам pretargeting.
3. Параллельно отправляет запрос всем подходящим DSP.
4. Использует общий таймаут для всего fan-out.
5. Не завершает весь auction ошибкой, если отдельный DSP недоступен, вернул ошибку или не уложился в таймаут.
6. Возвращает агрегированный результат.

Пример ответа:

```json
{
  "request_id": "test-auction-1",
  "matched_dsps": ["DSP Alpha", "DSP Beta", "DSP Gamma"],
  "sent": 3,
  "succeeded": 1,
  "duration_ms": 216
}
```

## Pretargeting

DSP участвует в запросе, если одновременно выполняются условия:

- `is_enabled = true`;
- страна запроса входит в `countries`;
- если `countries` пуст, разрешены все страны;
- тип устройства входит в `device_types`;
- если `device_types` пуст, разрешены все типы устройств;
- `bid_floor >= min_bid_floor`;
- ни одна категория запроса не входит в `blocked_categories`.

## Архитектура

```text
HTTP handler
    ↓
  usecase
   ↙   ↘
storage  DSP client
   ↓        ↓
PostgreSQL HTTP DSPs
```

### Handler

Отвечает за HTTP-уровень:

- декодирование JSON;
- валидацию входных данных;
- преобразование HTTP DTO в `usecase.Input`;
- вызов usecase;
- формирование HTTP-ответа.

### Usecase

Содержит основную бизнес-логику:

- получение DSP из storage;
- pretargeting;
- создание общего timeout context;
- параллельный fan-out запросов;
- агрегацию результатов.

Usecase зависит от интерфейсов, а не от конкретных реализаций storage или HTTP-клиента.

### Storage

Работает с PostgreSQL через `sqlx` и `lib/pq`.

Storage отвечает за получение данных и преобразование DB-модели в модель usecase. Бизнес-фильтрация в SQL не выполняется.

### DSP Client

Отвечает за один HTTP-запрос к одному DSP.

Параллелизм и общий таймаут управляются на уровне usecase.

## Технологии

- Go
- PostgreSQL
- `sqlx`
- `lib/pq`
- `golang-migrate`
- Docker
- Docker Compose
- `zap`
- `net/http`

## Запуск

Самый простой способ запустить весь стенд:

```bash
docker compose up --build
```

Docker Compose поднимает:

- PostgreSQL;
- контейнер миграций;
- AdEx Service;
- `DSP Alpha` — успешный ответ `200`;
- `DSP Beta` — ответ `500`;
- `DSP Gamma` — ответ `200` с задержкой `500ms`.

После запуска API доступен на:

```text
http://localhost:8080
```

## Healthcheck

```bash
curl -i http://localhost:8080/health
```

Ожидаемый ответ:

```json
{
  "status": "ok"
}
```

## Пример auction-запроса

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "test-auction-1",
    "country": "RU",
    "device_type": "mobile",
    "bid_floor": 1.5,
    "categories": ["news"]
  }'
```

Для seed-конфигурации запрос матчится с тремя DSP:

- `DSP Alpha` отвечает успешно;
- `DSP Beta` возвращает `500`;
- `DSP Gamma` отвечает позже общего таймаута;
- `DSP Delta` отключён.

Поэтому ожидаемый результат выглядит примерно так:

```json
{
  "request_id": "test-auction-1",
  "matched_dsps": ["DSP Alpha", "DSP Beta", "DSP Gamma"],
  "sent": 3,
  "succeeded": 1,
  "duration_ms": 200
}
```

`duration_ms` может незначительно отличаться от `200` из-за накладных расходов планировщика, HTTP и обработки результата.

## Пример запроса без подходящих DSP

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "test-auction-2",
    "country": "US",
    "device_type": "desktop",
    "bid_floor": 0.1,
    "categories": ["gambling"]
  }'
```

Пример ответа:

```json
{
  "request_id": "test-auction-2",
  "matched_dsps": [],
  "sent": 0,
  "succeeded": 0,
  "duration_ms": 1
}
```

## Конфигурация

Пример переменных окружения находится в `.env.example`.

| Переменная | Назначение | Пример |
| --- | --- | --- |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5432/adex?sslmode=disable` |
| `HTTP_ADDR` | адрес HTTP-сервера | `:8080` |
| `LOG_LEVEL` | уровень логирования | `info` |
| `DSP_TIMEOUT` | общий timeout для fan-out к DSP | `200ms` |
| `SHUTDOWN_TIMEOUT` | timeout graceful shutdown | `10s` |

При запуске через Docker Compose `DATABASE_URL` использует hostname `postgres`, а не `localhost`.

## Миграции

Миграции находятся в директории:

```text
migrations/
```

Для локального PostgreSQL их можно применить через `golang-migrate`:

```bash
migrate \
  -path ./migrations \
  -database "postgres://postgres:postgres@localhost:5432/adex?sslmode=disable" \
  up
```

При запуске через Docker Compose миграции применяются автоматически отдельным контейнером.

## Тесты

Запуск всех тестов:

```bash
go test ./...
```

Дополнительные проверки:

```bash
go vet ./...
go build ./...
```

В проекте покрыты тестами:

- правила pretargeting;
- фильтрация партнёров;
- ошибка storage;
- отсутствие подходящих DSP;
- успешные DSP-вызовы;
- partial failure;
- timeout;
- параллельный fan-out;
- формирование и отправка HTTP-запроса в DSP;
- обработка неуспешных HTTP-статусов;
- соблюдение `context deadline`.

## Graceful shutdown

Приложение обрабатывает системные сигналы и использует `http.Server.Shutdown`.

При завершении сервер перестаёт принимать новые соединения и получает ограниченное время на корректное завершение активных запросов.

Timeout настраивается через:

```text
SHUTDOWN_TIMEOUT
```

## Структура проекта

```text
cmd/
  app/
    main.go
  mockdsp/
    main.go

internal/
  api/
    handlers/
      auction/
  client/
    dsp/
  storage/
  usecases/
    auction/

migrations/
Dockerfile
docker-compose.yml
.env.example
```

## Основные решения

### Один общий timeout

Timeout создаётся один раз на уровне usecase и используется всеми параллельными DSP-запросами.

Таким образом медленный DSP не может увеличить общее время auction сверх заданного лимита.

### Partial failure

Ошибка отдельного DSP не считается ошибкой всего auction.

Например, если один партнёр вернул `500`, другой успешно ответил, а третий не уложился в timeout, сервис всё равно возвращает агрегированный результат.

### Параллельный fan-out

Запросы к подходящим DSP выполняются конкурентно.

Каждый DSP получает один запрос в отдельной goroutine, а результаты собираются через buffered channel.

### Dependency inversion

Интерфейсы storage и DSP client объявлены в usecase-пакете — там, где они используются.

Благодаря этому бизнес-логика не зависит от конкретного SQL-драйвера или HTTP-реализации и легко тестируется с fake-зависимостями.

## Остановка стенда

```bash
docker compose down
```

Для удаления также PostgreSQL volume:

```bash
docker compose down -v
```
