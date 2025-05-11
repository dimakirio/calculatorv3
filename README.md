# <img src="https://img.icons8.com/ios-filled/50/000000/calculator.png" width="32"/> Распределённый вычислитель арифметических выражений

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/doc/go1.21)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](#)
[![Issues](https://img.shields.io/github/issues/Egor213312/Sprint3)](https://github.com/Egor213312/Sprint3/issues)

---

Распределённый сервис вычисления арифметических выражений позволяет пользователям отправлять арифметические выражения по HTTP и получать результаты их вычислений. Все данные пользователей и выражения хранятся в SQLite. Для доступа к API требуется регистрация и аутентификация (JWT).

---

## 📋 Содержание
- [О проекте](#о-проекте)
- [Возможности](#возможности)
- [Установка и запуск](#установка-и-запуск)
- [Использование API](#использование-api)
- [Тестирование](#тестирование)
- [Переменные окружения](#переменные-окружения)
- [Контакты](#контакты)

---

## О проекте

Проект "Сервис подсчёта арифметических выражений" — это распределённая система для вычисления арифметических выражений с поддержкой многопользовательского режима, JWT-аутентификации и хранения истории вычислений в базе данных. Система написана на Go и легко масштабируется.

---

## Возможности
- Регистрация и аутентификация пользователей (JWT)
- Добавление арифметических выражений на вычисление
- Получение истории своих выражений
- Получение результата по ID выражения
- Обработка ошибок с понятными сообщениями
- Хранение данных в SQLite (переживает перезапуск)
- Примеры для Postman и curl
- Модульные тесты

---

## Установка и запуск

### Локально
```bash
# 1. Клонируйте репозиторий
 git clone https://github.com/dimakirio/calculatorv1.git
 cd calculatorv1
# 2. Установите зависимости
 go mod tidy
# 3. Запустите сервер
 go run ./cmd/main.go
```

### Через Docker
```bash
git clone https://github.com/dimakirio/calculatorv1.git
cd calculatorv1
docker-compose up --build
```

Сервер будет доступен на [http://localhost:8080](http://localhost:8080)

---

## Использование API

### 1. Регистрация пользователя
```bash
curl --location 'http://localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
    "login": "testuser",
    "password": "testpass123"
}'
```

### 2. Вход в систему (логин)
```bash
curl --location 'http://localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
    "login": "testuser",
    "password": "testpass123"
}'
```
**Ответ:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 3. Добавление выражения
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <ваш_JWT_токен>' \
--data '{
    "expression": "2 + 2 * 2"
}'
```

### 4. Получение списка выражений
```bash
curl --location 'http://localhost:8080/api/v1/expressions' \
--header 'Authorization: Bearer <ваш_JWT_токен>'
```

### 5. Получение выражения по ID
```bash
curl --location 'http://localhost:8080/api/v1/expressions/{id}' \
--header 'Authorization: Bearer <ваш_JWT_токен>'
```

---

## Примеры ошибок

- **Некорректное выражение:**
  - Запрос: `{"expression": "2 + * 2"}`
  - Ответ: `422 Unprocessable Entity`, JSON: `{ "error": "Invalid expression" }`
- **Неавторизованный доступ:**
  - Ответ: `401 Unauthorized`, JSON: `{ "error": "Invalid token" }`
- **Несуществующий ID:**
  - Ответ: `404 Not Found`, JSON: `{ "error": "Expression not found" }`

---

## Тестирование

Для запуска модульных тестов:
```bash
cd calculatorv1
# Запуск всех тестов
 go test ./internal/orchestrator
```

---

## Переменные окружения

| Переменная      | Описание                        | Значение по умолчанию |
|-----------------|----------------------------------|-----------------------|
| SERVER_PORT     | Порт сервера                     | 8080                  |
| LOG_LEVEL       | Уровень логирования             | info                  |
| JWT_SECRET      | Секрет для JWT                  | your-secret-key       |
| DB_PATH         | Путь к базе данных SQLite       | calc.db               |

---
