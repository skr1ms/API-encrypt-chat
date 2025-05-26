# EncryptChat Application

Зашифрованный мессенджер с защищенной передачей сообщений, использующий современные криптографические алгоритмы.

## Технологии

### Backend
- Go (Gin)
- PostgreSQL
- WebSockets
- JWT аутентификация
- Криптография: AES, ECDSA, RSA, HMAC

### Frontend
- React
- TypeScript
- Vite
- Redux Toolkit
- React Router
- shadcn/ui (Tailwind CSS)
- Crypto.js, Elliptic, Node-forge (Криптография на стороне клиента)

## Запуск с помощью Docker

Приложение настроено для запуска с использованием Docker Compose, что позволяет легко развернуть все компоненты:
- Frontend (React)
- Backend (Go)
- База данных (PostgreSQL)

### Требования
- Docker
- Docker Compose

### Инструкция по запуску

1. Клонировать репозиторий
2. Перейти в корневую папку проекта
3. **Убедиться, что Docker Desktop запущен**
4. Запустить приложение одной командой:

```bash
# На Linux/Mac
./start.sh

# На Windows
start.bat
```

Или вручную с помощью Docker Compose:

```bash
docker-compose up --build
```

### Доступ к приложению

После запуска контейнеров приложение будет доступно:
- Frontend: http://localhost
- Backend API: http://localhost/api
- WebSocket: ws://localhost/ws

## Структура проекта

```
.
├── backend/               # Go backend
│   ├── cmd/server/        # Точка входа сервера
│   ├── internal/          # Внутренние пакеты и модули
│   │   ├── adapters/      # Адаптеры и обработчики
│   │   ├── crypto/        # Криптографические модули
│   │   ├── domain/        # Бизнес-логика и модели
│   │   └── infrastructure/ # Инфраструктурные компоненты
│   └── pkg/               # Общие пакеты
├── frontend/             # React frontend
│   ├── src/              # Исходный код
│   │   ├── app/          # Ядро приложения
│   │   ├── components/   # Компоненты UI
│   │   ├── features/     # Функциональные модули
│   │   ├── pages/        # Страницы приложения
│   │   └── shared/       # Общие утилиты и компоненты
└── docker-compose.yml    # Конфигурация Docker Compose
```

## Безопасность

- End-to-end шифрование сообщений
- Хранение паролей с использованием безопасного хеширования
- Защита против XSS и CSRF атак
- Защищённые соединения с использованием TLS
- Подписывание сообщений для проверки подлинности
