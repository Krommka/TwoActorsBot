* usecase/  Бизнес-логика
* pkg/
    * logger/  Логирование
    * prometheus/  Метрики
* docker-compose.yml  Оркестрация сервисов
## Команды Makefile

all: run        Запуск по умолчанию (docker compose)

local:          Локальный запуск без Docker

run:            Запуск контейнеров

stop:           Остановка контейнеров

build:          Сборка образов

restart:             Пересборка и перезапуск

## Настройка
* Скопируйте .env.template в .env

* Заполните переменные:

  TG_TOKEN - Токен Telegram бота

  KINOPOISK_API_KEY - Ключ API Кинопоиска

  REDIS_URL - Адрес Redis сервера
## Мониторинг
* Сервисы мониторинга:

  Prometheus (метрики): http://localhost:9090

  Grafana (дашборды): http://localhost:3000

  Loki (логи)
##  Требования
* Go 1.21+
* Docker
* Доступ к API Telegram Bot
* Ключ API Кинопоиска