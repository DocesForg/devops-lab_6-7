# DevOps: лабораторные работы 6–8

Вариант 6: микросервис сотрудников (Employee), CI/CD и контейнерное развёртывание с PostgreSQL.

[![CI](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/ci.yml/badge.svg)](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/ci.yml)
[![Docker](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/docker.yml/badge.svg)](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/docker.yml)

## Возможности сервиса

- Создание сотрудника: HTTP `POST /employees` и gRPC `EmployeeService/Create`.
- Получение сотрудника: HTTP `GET /employees/{id}` и gRPC `EmployeeService/GetByID`.
- Изменение должности: HTTP `PATCH /employees/{id}/position` и gRPC `EmployeeService/ChangePosition`.
- Проверка HTTP-состояния: `GET /healthz`.
- Go kit endpoints и HTTP/gRPC-транспорты используют общий бизнес-сервис.
- HTTP слушает порт `8080`, gRPC — порт `9090`.
- Реализованы два хранилища: `MemoryRepository` для локального запуска без настроенной БД и `PostgresRepository` для постоянного хранения.
- PostgreSQL-репозиторий автоматически создаёт таблицу `employees` при подключении.
- Есть логирование HTTP-запросов и корректное завершение серверов.

## Запуск без PostgreSQL

Требуется Go 1.26 или новее.

```bash
go mod tidy
go test -race ./...
go vet ./...
go run .
```

Если переменная `DB_HOST` не задана, приложение использует in-memory хранилище. Данные в нём теряются при перезапуске.

## Запуск с PostgreSQL (лабораторная №8, вариант 6)

Рекомендуемый способ — Docker Compose. Он создаёт общую сеть, запускает PostgreSQL с именем `db`, сохраняет данные в named volume и запускает API после прохождения healthcheck БД.

```bash
docker compose up --build -d
docker compose ps
docker compose logs api
```

По умолчанию для локальной учебной среды используется пароль `change_me_local`. Для изменения задайте `DB_PASSWORD` в окружении перед запуском.

Проверить API:

```bash
curl http://localhost:8080/healthz
curl -i -X POST http://localhost:8080/employees \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Ivan Ivanov","position":"Engineer","salary":80000}'
```

Используйте полученный из POST идентификатор:

```bash
curl http://localhost:8080/employees/1
curl -i -X PATCH http://localhost:8080/employees/1/position \
  -H 'Content-Type: application/json' \
  -d '{"position":"Senior Engineer"}'
```

Проверить, что данные действительно записаны в PostgreSQL:

```bash
docker compose exec db psql -U employee_user -d employees \
  -c "SELECT id, full_name, position, salary, hired_on FROM employees;"
```

Проверить сетевое имя БД из отдельного контейнера в той же сети:

```bash
docker run --rm --network appnet postgres:16 \
  pg_isready -h db -p 5432 -U employee_user -d employees
```

Если Compose создал сеть с другим именем, узнайте его командой `docker network ls`.

Проверить сохранность данных после пересоздания контейнеров:

```bash
docker compose down
docker compose up -d
docker compose exec db psql -U employee_user -d employees -c "SELECT * FROM employees;"
```

Named volume `employee-db-data` не удаляется командой `docker compose down`, поэтому данные остаются. Для полного удаления вместе с данными используйте `docker compose down -v`.

Для ручного запуска приложение поддерживает переменные `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` и `DB_SSLMODE`. В Docker-сети значение `DB_HOST` должно быть равно имени контейнера PostgreSQL — `db`, а не `localhost`.

PostgreSQL integration test запускается при наличии `TEST_DATABASE_URL`:

```bash
TEST_DATABASE_URL="host=localhost port=5432 dbname=employees user=employee_user password=change_me_local sslmode=disable" go test -run TestPostgresRepositoryCRUD -v
```

## HTTP-примеры

Создать сотрудника:

```bash
curl -i -X POST http://localhost:8080/employees \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Иван Иванов","position":"Инженер","salary":80000}'
```

Получить сотрудника:

```bash
curl http://localhost:8080/employees/1
```

Изменить должность:

```bash
curl -i -X PATCH http://localhost:8080/employees/1/position \
  -H 'Content-Type: application/json' \
  -d '{"position":"Ведущий инженер"}'
```

## gRPC-примеры

Для вызовов установите `grpcurl`. Серверу не требуется включать reflection:

```bash
grpcurl -plaintext -import-path . -proto employee.proto \
  -d '{"full_name":"Ivan Ivanov","position":"Engineer","salary":80000}' \
  localhost:9090 employee.v1.EmployeeService/Create
```

```bash
grpcurl -plaintext -import-path . -proto employee.proto \
  -d '{"id":"1"}' localhost:9090 employee.v1.EmployeeService/GetByID

grpcurl -plaintext -import-path . -proto employee.proto \
  -d '{"id":"1","position":"Senior Engineer"}' \
  localhost:9090 employee.v1.EmployeeService/ChangePosition
```

## CI/CD (лабораторная №7)

- `.github/workflows/ci.yml` проверяет сборку, тесты с race detector, coverage, `go vet`, `govulncheck` и `gosec`.
- `.github/workflows/docker.yml` собирает multi-stage Docker image и публикует его в Docker Hub при наличии secrets `DOCKERHUB_USERNAME` и `DOCKERHUB_TOKEN`.

## Отчёты

- [Лабораторная работа №6](l6.md)
- [Лабораторная работа №7](l7.md)
- [Отчёт по Docker (вариант 6)](lab6.md)
