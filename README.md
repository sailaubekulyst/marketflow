# Marketflow

Сервис агрегации цен с HTTP API и хранением в SQLite.

## Запуск

```bash
go run ./cmd/main.go
```

Сервер слушает порт **8080**.

## API

- **GET** `/prices/highest/{symbol}` — максимальная цена по символу
- **GET** `/prices/highest/{exchange}/{symbol}` — максимальная цена по бирже и символу  

Опциональный query-параметр: `period` (например, `1h`, `24h`).

## Данные

SQLite-база `marketflow.db` создаётся в текущей директории при первом запуске.

## Остановка

Graceful shutdown по **Ctrl+C** или **SIGTERM**: завершение активных запросов (таймаут 5 сек), затем остановка приложения.
