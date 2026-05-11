# 🏦 Online Bank

Веб-приложение интернет-банка, написанное на **Go**. Позволяет пользователям управлять банковскими счетами: пополнять баланс, переводить средства другим пользователям и конвертировать валюты в режиме реального времени.

---

## 📋 Содержание

- [Возможности](#-возможности)
- [Стек технологий](#-стек-технологий)
- [Структура проекта](#-структура-проекта)
- [База данных](#-база-данных)
- [Установка и запуск](#-установка-и-запуск)
- [Конфигурация](#-конфигурация)
- [Маршруты приложения](#-маршруты-приложения)
- [API конвертации валют](#-api-конвертации-валют)

---

## ✨ Возможности

- **Регистрация и авторизация** — создание аккаунта, вход по email/паролю, сессия через cookie-токен
- **Личный кабинет** — просмотр баланса в трёх валютах (TJS, USD, EUR), редактирование профиля и аватара
- **Пополнение счёта** — зачисление средств на баланс в сомони (TJS)
- **Переводы** — отправка средств другим зарегистрированным пользователям
- **Конвертация валют** — обмен между TJS, USD и EUR по актуальному курсу через внешнее API
- **История транзакций** — полный журнал операций: пополнения, переводы, конвертации

---

## 🛠 Стек технологий

| Компонент      | Технология                         |
|----------------|------------------------------------|
| Язык           | Go 1.25                            |
| База данных    | PostgreSQL                         |
| Драйвер БД     | `github.com/lib/pq`                |
| Хэширование    | `golang.org/x/crypto/bcrypt`       |
| Шаблонизатор   | `html/template` (стандартная б-ка) |
| UI             | Bootstrap 5.3                      |
| Курсы валют    | [Open Exchange Rates API](https://openexchangerates.org) |

---

## 📁 Структура проекта

```
Online_Bank/
├── main.go                      # Точка входа, регистрация маршрутов
├── config.json                  # Конфигурация подключения к БД
├── go.mod
├── go.sum
│
├── config/
│   └── config.go                # Загрузка конфигурации из JSON
│
├── db/
│   └── postgres.go              # Подключение к PostgreSQL
│
├── internal/
│   ├── currency/
│   │   └── currency.go          # Получение курсов валют через Open Exchange Rates
│   └── user/
│       ├── handler.go           # HTTP-обработчики (регистрация, логин, дашборд и др.)
│       ├── service.go           # Бизнес-логика (валидация, токены, конвертация)
│       ├── repository.go        # Работа с БД (SQL-запросы)
│       └── model.go             # Модели данных (User, Transactions, AboutPerson)
│
├── templates/
│   ├── register.html
│   ├── login.html
│   ├── dashboard.html
│   ├── deposit.html
│   ├── transfer.html
│   ├── convert.html
│   ├── transactions.html
│   └── about.html
│
└── uploads/                     # Загруженные аватары пользователей
    └── default-avatar.jpg
```

---

## 🗄 База данных

Приложение использует три основные таблицы:

**`users`** — аккаунты пользователей
```sql
CREATE TABLE users (
    id           SERIAL PRIMARY KEY,
    name         TEXT NOT NULL,
    email        TEXT UNIQUE NOT NULL,
    password     TEXT NOT NULL,           -- bcrypt-хэш
    balance_tjs  NUMERIC DEFAULT 100.0,
    balance_usd  NUMERIC DEFAULT 0.0,
    balance_eur  NUMERIC DEFAULT 0.0,
    created_at   TIMESTAMP
);
```

**`profiles`** — профили пользователей
```sql
CREATE TABLE profiles (
    user_id     INT REFERENCES users(id),
    full_name   TEXT,
    bio         TEXT,
    avatar_path TEXT DEFAULT 'uploads/default-avatar.jpg',
    updated_at  TIMESTAMP
);
```

**`user_tokens`** — сессионные токены
```sql
CREATE TABLE user_tokens (
    token       TEXT PRIMARY KEY,
    user_id     INT REFERENCES users(id),
    created_at  TIMESTAMP
);
```

**`transactions`** — история операций
```sql
CREATE TABLE transactions (
    id          SERIAL PRIMARY KEY,
    user_id     INT REFERENCES users(id),
    type        TEXT,          -- 'deposit', 'transfer', 'conversion'
    amount      NUMERIC,
    currency    TEXT,
    description TEXT,
    created_at  TIMESTAMP
);
```

---

## 🚀 Установка и запуск

### Предварительные требования

- Go 1.21+
- PostgreSQL 14+

### 1. Клонирование репозитория

```bash
git clone https://github.com/your-username/online_bank.git
cd online_bank
```

### 2. Создание базы данных

```sql
CREATE DATABASE online_bank;
```

Затем выполните SQL-скрипты для создания таблиц (см. раздел [База данных](#-база-данных)).

### 3. Настройка конфигурации

Создайте файл `config.json` в корне проекта:

```json
{
  "db_user": "postgres",
  "db_password": "your_password",
  "db_name": "online_bank"
}
```

### 4. Установка зависимостей

```bash
go mod download
```

### 5. Запуск приложения

```bash
go run main.go
```

Приложение будет доступно по адресу: **http://localhost:8080/login**

---

## ⚙️ Конфигурация

| Параметр      | Файл          | Описание                             |
|---------------|---------------|--------------------------------------|
| `db_user`     | `config.json` | Имя пользователя PostgreSQL          |
| `db_password` | `config.json` | Пароль PostgreSQL                    |
| `db_name`     | `config.json` | Имя базы данных                      |
| `apiKey`      | `main.go`     | API-ключ Open Exchange Rates         |

> ⚠️ **Важно:** не коммитьте `config.json` с реальными учётными данными. Добавьте его в `.gitignore`.

---

## 🗺 Маршруты приложения

| Метод       | Маршрут         | Описание                                  | Требует авторизации |
|-------------|-----------------|-------------------------------------------|---------------------|
| GET / POST  | `/register`     | Страница регистрации                      | ❌                  |
| GET / POST  | `/login`        | Страница входа                            | ❌                  |
| GET         | `/dashboard`    | Личный кабинет с балансом                 | ✅                  |
| GET / POST  | `/deposit`      | Пополнение счёта                          | ✅                  |
| GET / POST  | `/transfer`     | Перевод другому пользователю              | ✅                  |
| GET / POST  | `/convert`      | Конвертация валюты                        | ✅                  |
| GET         | `/transactions` | История транзакций                        | ✅                  |
| GET / POST  | `/about`        | Просмотр и редактирование профиля         | ✅                  |
| GET         | `/logout`       | Выход из системы                          | ✅                  |

---

## 💱 API конвертации валют

Для получения актуальных курсов используется [Open Exchange Rates](https://openexchangerates.org). Логика конвертации с участием TJS реализована через USD как базовую валюту:

- `TJS → USD/EUR`: TJS конвертируется в USD по фиксированному коэффициенту, затем в целевую валюту через API
- `USD/EUR → TJS`: сначала в USD через API, затем в TJS

Для использования необходим бесплатный или платный `app_id` на сайте Open Exchange Rates.

---

## 📄 Лицензия

Проект разработан в учебных целях.
