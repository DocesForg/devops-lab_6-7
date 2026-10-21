# DevOps: лабораторные работы 6–7

Вариант 6: микросервис сотрудников (Employee) и CI/CD с проверками безопасности.

[![CI](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/ci.yml/badge.svg)](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/ci.yml)
[![Docker](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/docker.yml/badge.svg)](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/docker.yml)

## Возможности сервиса

- Создание сотрудника: HTTP `POST /employees` и gRPC `EmployeeService/Create`.
- Получение сотрудника: HTTP `GET /employees/{id}` и gRPC `EmployeeService/GetByID`.
- Изменение должности: HTTP `PATCH /employees/{id}/position` и gRPC `EmployeeService/ChangePosition`.
- Проверка работоспособности HTTP: `GET /healthz`.
- Go kit endpoints и адаптеры HTTP/gRPC используют общий бизнес-сервис.
- HTTP слушает порт `8080`, gRPC — порт `9090`.
- Хранилище находится в памяти процесса; данные не переживают перезапуск.
- Логирование HTTP-запросов и корректное завершение обоих серверов.

Поля Employee: `id`, `full_name`, `position`, `salary`, `hired_on` (Unix timestamp). Контракт gRPC описан в `employee.proto`; сообщения запроса/ответа — protobuf `google.protobuf.Struct`.

## Запуск

Требуется Go 1.25 или новее:

```bash
go mod download
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

В этом контракте gRPC payload использует protobuf `Struct`, поэтому поля передаются как JSON-объект. При расширении API можно заменить `Struct` типизированными protobuf-сообщениями и добавить сгенерированный код.

## CI/CD (лабораторная №7, вариант 6)

Workflow `.github/workflows/ci.yml` запускается на push/PR в `main` и вручную. Матрица использует поддерживаемые ветки Go 1.25.x и stable, выполняет сборку, тесты с race detector и coverage, `go vet`, `govulncheck` и `gosec`. Профиль покрытия сохраняется как artifact.

Workflow `.github/workflows/docker.yml` собирает multi-stage Docker image. Публикация в Docker Hub выполняется при push, если заданы repository secrets `DOCKERHUB_USERNAME` и `DOCKERHUB_TOKEN`. Если секреты недоступны, workflow всё равно проверяет сборку образа, но явно пропускает публикацию. Имя образа приводится к допустимому нижнему регистру: `docesforg/devops-lab-6-7`.

## Отчёты

- [Лабораторная работа №6](lab6.md)
- [Лабораторная работа №7](lab7.md)

## Ограничения

Хранилище in-memory очищается при перезапуске. Для эксплуатации нужны постоянное хранилище, аутентификация, авторизация и TLS для gRPC. gRPC использует protobuf `Struct` вместо типизированных сообщений.
