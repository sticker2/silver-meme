Распределенный калькулятор
Описание проекта
Распределенный калькулятор — это многопользовательская система для вычисления математических выражений с поддержкой персистентности данных. Проект реализует клиент-серверную архитектуру, где calc_service обрабатывает запросы пользователей через HTTP API, а agent_service выполняет вычисления через gRPC. Данные (пользователи, выражения, задачи) хранятся в SQLite, обеспечивая восстановление состояния после перезагрузки.
Основной функционал

Регистрация и авторизация: Пользователи могут регистрироваться и входить в систему, получая JWT-токен для авторизации.
Вычисление выражений: Пользователи отправляют математические выражения (например, 2 + 3 * 4), которые обрабатываются распределенно.
Хранение данных: Все выражения и результаты сохраняются в базе данных SQLite.
Распределенные вычисления: agent_service выполняет задачи, получаемые через HTTP от calc_service, и возвращает результаты через gRPC.
Многопользовательский режим: Каждый пользователь видит только свои выражения.

Проект размещен на GitHub: https://github.com/sticker2/silver-meme/tree/main/DistributedCalc.
Требования

Go 1.21 или выше
SQLite (встроен, не требует установки)
Docker и Docker Compose (для запуска в контейнерах)
protoc (для генерации gRPC-кода)

Установка и запуск
Локальный запуск

Склонируйте репозиторий:
git clone https://github.com/sticker2/silver-meme.git
cd silver-meme/DistributedCalc


Сгенерируйте gRPC-код:
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/grpc/calc.proto


Соберите сервисы:
go build -o calc_service ./cmd/calc_service
go build -o agent_service ./cmd/agent_service


Запустите сервисы в отдельных терминалах:
./calc_service
./agent_service


calc_service запускает HTTP-сервер на :8080.
agent_service запускает gRPC-сервер на :50051 и обрабатывает задачи.



Запуск с Docker Compose

Убедитесь, что файлы docker-compose.yml, Dockerfile.calc и Dockerfile.agent находятся в корне проекта. Если их нет, создайте их (см. ниже).

Запустите:
docker-compose up --build

Это создаст и запустит контейнеры для calc_service и agent_service. Доступ к API: http://localhost:8080.


Файлы для Docker
Если docker-compose.yml отсутствует, создайте его в корне проекта:
version: '3.8'

services:
  calc_service:
    build:
      context: .
      dockerfile: Dockerfile.calc
    ports:
      - "8080:8080"
    volumes:
      - db_data:/app
    environment:
      - TIME_ADDITION_MS=100
      - TIME_SUBTRACTION_MS=100
      - TIME_MULTIPLICATIONS_MS=100
      - TIME_DIVISIONS_MS=100
    networks:
      - calc_network
    depends_on:
      - agent_service

  agent_service:
    build:
      context: .
      dockerfile: Dockerfile.agent
    ports:
      - "50051:50051"
    volumes:
      - db_data:/app
    environment:
      - COMPUTING_POWER=4
      - TIME_ADDITION_MS=100
      - TIME_SUBTRACTION_MS=100
      - TIME_MULTIPLICATIONS_MS=100
      - TIME_DIVISIONS_MS=100
    networks:
      - calc_network

volumes:
  db_data:

networks:
  calc_network:
    driver: bridge

Создайте Dockerfile.calc:
FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o calc_service ./cmd/calc_service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/calc_service .
EXPOSE 8080
CMD ["./calc_service"]

Создайте Dockerfile.agent:
FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o agent_service ./cmd/agent_service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/agent_service .
EXPOSE 50051
CMD ["./agent_service"]

Использование API
Регистрация пользователя
curl --location 'http://localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "testuser",
  "password": "testpass"
}'


Успех: {"message":"user registered"} (200 OK)
Ошибка (пользователь существует): {"code":400,"message":"user already exists"} (400 Bad Request)

Вход пользователя
curl --location 'http://localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "testuser",
  "password": "testpass"
}'


Успех: {"token":"..."} (200 OK)
Ошибка (неверный логин/пароль): {"code":401,"message":"invalid credentials"} (401 Unauthorized)

Отправка выражения
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <your-jwt-token>' \
--data '{
  "expression": "2 + 3 * 4"
}'


Успех: {"id":1} (200 OK)
Ошибка (неверный токен): {"code":401,"message":"invalid token"} (401 Unauthorized)
Ошибка (некорректное выражение): {"code":400,"message":"invalid expression"} (400 Bad Request)

Получение списка выражений
curl --location 'http://localhost:8080/api/v1/expressions' \
--header 'Authorization: Bearer <your-jwt-token>'


Успех: [{"id":1,"expression":"2 + 3 * 4","result":14,"status":"completed"}] (200 OK)
Ошибка (неверный токен): {"code":401,"message":"invalid token"} (401 Unauthorized)

Получение конкретного выражения
curl --location 'http://localhost:8080/api/v1/expression?id=1' \
--header 'Authorization: Bearer <your-jwt-token>'


Успех: {"id":1,"expression":"2 + 3 * 4","result":14,"status":"completed"} (200 OK)
Ошибка (выражение не найдено): {"code":404,"message":"expression not found"} (404 Not Found)

Тестирование
Проект включает модульные и интеграционные тесты (если они реализованы). Для запуска:
go test ./...

Примечание: На момент написания README тесты отсутствуют в репозитории. Если тесты требуются, обратитесь к автору.
Критерии реализации

Многопользовательский режим: Каждый пользователь работает со своими выражениями, аутентификация через JWT.
Персистентность: Данные хранятся в SQLite (calc.db), сохраняются после перезагрузки.
gRPC: Взаимодействие между calc_service и agent_service реализовано через gRPC (CalcService).
Тесты: Модульные и интеграционные тесты (если реализованы) покрывают функционал.

Устранение неполадок

Ошибка компиляции: Убедитесь, что gRPC-код сгенерирован (protoc выполнен).
Ошибка запуска: Проверьте, что порты :8080 и :50051 свободны.
Логи: Смотрите calc_service.log и agent_service.log для диагностики.
