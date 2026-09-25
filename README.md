# tgBank
# tgBank

**tgBank** — backend-приложение на Go для управления арендой банковских карт через Telegram-бота.

Система предоставляет пользователю интерфейс в Telegram для регистрации, работы с банковскими картами и операциями аренды. Для выполнения асинхронных операций используется Apache Kafka, а взаимодействие с GSM-шлюзом GoIP позволяет отправлять и получать SMS.

## Основной стек

* **Go**
* **PostgreSQL**
* **pgx / connection pool**
* **Apache Kafka**
* **Redis**
* **Telegram Bot API**
* **GoIP**
* **Prometheus**
* **Docker / Docker Compose**
* **GitHub Actions**

## Архитектура

Приложение разделено на несколько логических слоёв:

* Telegram handlers — обработка пользовательских команд и callback-запросов;
* Service layer — бизнес-логика приложения;
* Repository layer — работа с PostgreSQL;
* GoIP client — взаимодействие с GSM-шлюзом;
* Kafka producer/consumer — асинхронная обработка команд;
* Metrics — сбор технических метрик приложения.

Основной пользовательский поток проходит через Telegram-бота, который передаёт запросы в service layer. Бизнес-логика взаимодействует с PostgreSQL через repository layer.

Отдельные команды для GoIP передаются через Kafka topic `goip.commands`, после чего consumer обрабатывает их и взаимодействует с GSM-шлюзом.

## Основные возможности

* Регистрация пользователей через Telegram;
* Управление пользователями и их статусами;
* Работа с банковскими картами;
* Аренда карт;
* Отслеживание срока аренды;
* Отправка SMS через GoIP;
* Асинхронная обработка GoIP-команд через Kafka;
* Работа с PostgreSQL;
* Миграции базы данных;
* Сбор метрик состояния connection pool PostgreSQL;
* Dockerized local environment.

## Асинхронное взаимодействие

Для операций с GoIP используется Kafka.

```text
Telegram Bot
     │
     ▼
Service
     │
     ▼
Kafka Producer
     │
     ▼
goip.commands
     │
     ▼
Kafka Consumer
     │
     ▼
GoIP
     │
     ▼
SMS
```

Такой подход позволяет отделить пользовательский поток от операций взаимодействия с внешним GSM-шлюзом и выполнять их асинхронно.

## Мониторинг

Приложение собирает метрики PostgreSQL connection pool через Prometheus.

Отслеживаются, в частности:

* количество активных соединений;
* количество свободных соединений.

Метрики обновляются периодически в отдельной goroutine.

## Infrastructure

Проект содержит Docker Compose configuration, Dockerfile, Prometheus configuration и GitHub Actions workflow.

## Цель проекта

Проект создан как практический backend-проект для изучения разработки сервисов на Go, работы с PostgreSQL, Kafka, внешними API, Telegram Bot API и асинхронными процессами.
