# DevOps: лабораторные работы 6–7

Вариант 6: микросервис сотрудников (Employee) и CI/CD с проверками безопасности.

[![CI](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/ci.yml/badge.svg)](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/ci.yml)
[![Docker](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/docker.yml/badge.svg)](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/docker.yml)

## Возможности сервиса

- Создание сотрудника: HTTP `POST /employees` и gRPC `EmployeeService/Create`.
- Получение сотрудника: HTTP `GET /employees/{id}` и gRPC `EmployeeService/GetByID`.
- Изменение должности: HTTP `PATCH /employees/{id}/position` и gRPC `EmployeeService/ChangePosition`.
- Проверка работоспособности HTTP: `GET /healthz`.
- Go kit endpoints и HTTP/gRPC-транспорты используют общий бизнес-сервис.
- HTTP слушает порт `8080`, gRPC — порт `9090`.
- Хранилище находится в памяти процесса; данные не переживают перезапуск.
- Есть логирование HTTP-запросов и корректное завершение обоих серверов.

Поля Employee: `id`, `full_name`, `position`, `salary`, `hired_on` (Unix timestamp). Контракт gRPC описан в `employee.proto`; типизированные Go-типы и клиент/серверный интерфейс хранятся в `gen/employee/v1`.

## Запуск

Требуется Go 1.26 или новее.

```bash
go mod tidy
go test -race ./...
go vet ./...
go run .
```

По умолчанию HTTP-сервер слушает `:8080`, а gRPC — `:9090`. Адреса можно изменить:

```bash
go run . -http.addr=:8081 -grpc.addr=:9091
```

## HTTP-примеры

Создать сотрудника:

```bash
curl -i -X POST http://localhost:8080/employees \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Иван Иванов","position":"Инженер","salary":80000}'
```

Получить сотрудника (подставьте ID из ответа):

```bash
curl http://localhost:8080/employees/1
```

Изменить должность:

```bash
curl -i -X PATCH http://localhost:8080/employees/1/position \
  -H 'Content-Type: application/json' \
  -d '{"position":"Ведущий инженер"}'
```

Проверить HTTP health endpoint:

```bash
curl http://localhost:8080/healthz
```

## gRPC-примеры

Для вызовов установите `grpcurl`. Команды используют файл контракта из репозитория; серверу не требуется включать reflection:

```bash
grpcurl -plaintext -import-path . -proto employee.proto \
  -d '{"full_name":"Иван Иванов","position":"Инженер","salary":80000}' \
  localhost:9090 employee.v1.EmployeeService/Create
```

Используйте полученный `id` в последующих запросах:

```bash
grpcurl -plaintext -import-path . -proto employee.proto \
  -d '{"id":"1"}' localhost:9090 employee.v1.EmployeeService/GetByID

grpcurl -plaintext -import-path . -proto employee.proto \
  -d '{"id":"1","position":"Ведущий инженер"}' \
  localhost:9090 employee.v1.EmployeeService/ChangePosition
```

Для повторной генерации protobuf-кода локально нужны `protoc` и Go generators; выполните `make generate`. Сгенерированные файлы хранятся в репозитории, отдельный GitHub Actions workflow для их автоматического коммита не используется.

## CI/CD (лабораторная №7, вариант 6)

Оставлены два workflow:

- `.github/workflows/ci.yml` запускается на push и pull request в `main`, а также вручную. Проверяет Go 1.27.1 и stable: сборку, тесты с race detector и coverage, `go vet`, `govulncheck` и `gosec`. Профиль покрытия сохраняется как artifact.
- `.github/workflows/docker.yml` собирает multi-stage Docker image. Публикация в Docker Hub выполняется при push, если заданы repository secrets `DOCKERHUB_USERNAME` и `DOCKERHUB_TOKEN`. Без секретов образ проверяется сборкой, но не публикуется. Имя образа: `docesforg/devops-lab-6-7`.

## Отчёты

- [Лабораторная работа №6](lab6.md)
- [Лабораторная работа №7](lab7.md)

## Ограничения

Хранилище in-memory очищается при перезапуске. Для эксплуатации нужны постоянное хранилище, аутентификация, авторизация и TLS для gRPC.
