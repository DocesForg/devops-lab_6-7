# DevOps: лабораторные работы 6–7

Вариант 6: микросервис сотрудников (Employee) и CI/CD с проверками безопасности.

[![CI](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/ci.yml/badge.svg)](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/ci.yml)
[![Docker](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/docker.yml/badge.svg)](https://github.com/DocesForg/devops-lab_6-7/actions/workflows/docker.yml)

## Возможности сервиса

- Создание сотрудника: `POST /employees`.
- Получение сотрудника: `GET /employees/{id}`.
- Изменение должности: `PATCH /employees/{id}/position`.
- Проверка работоспособности: `GET /healthz`.
- Хранилище в памяти процесса; данные не переживают перезапуск.
- HTTP JSON API, логирование запросов, graceful shutdown.

Поля Employee: `id`, `full_name`, `position`, `salary`, `hired_on` (Unix timestamp).

## Запуск

Требуется Go 1.21 или новее:

```bash
go test -race ./...
go vet ./...
go run .
```

Сервер слушает `localhost:8080`. Адрес можно изменить: `go run . -http.addr=:8081`.

## Примеры запросов

Создать сотрудника:

```bash
curl -i -X POST http://localhost:8080/employees \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Иван Иванов","position":"Инженер","salary":80000}'
```

Ответ содержит сгенерированный `id`, например `{"id":"1"}`.

Получить данные (подставьте ID из ответа):

```bash
curl http://localhost:8080/employees/1
```

Изменить должность:

```bash
curl -i -X PATCH http://localhost:8080/employees/1/position \
  -H 'Content-Type: application/json' \
  -d '{"position":"Ведущий инженер"}'
```

Проверить health endpoint:

```bash
curl http://localhost:8080/healthz
```

## CI/CD (лабораторная №7, вариант 6)

Workflow `.github/workflows/ci.yml` запускается на push/PR в `main` и вручную. Матрица проверяет Go 1.21 и 1.22: сборка, тесты с race detector и coverage, `go vet`. Дополнительно выполняются `govulncheck` и `gosec` для security-проверок; отчёт покрытия сохраняется как artifact.

Workflow `.github/workflows/docker.yml` собирает multi-stage Docker image. Для публикации в Docker Hub при push нужно добавить repository secrets `DOCKERHUB_USERNAME` и `DOCKERHUB_TOKEN` в Settings → Secrets and variables → Actions. При pull request образ только собирается, но не публикуется.

## Отчёты

- [Лабораторная работа №6](lab6.md)
- [Лабораторная работа №7](lab7.md)

## Ограничения текущей реализации

Данные хранятся только в памяти и очищаются при перезапуске. Для реальной эксплуатации следует добавить постоянное хранилище, аутентификацию и авторизацию. gRPC-транспорт из задания №6 пока не подключён; основной реализованный транспорт — HTTP.
