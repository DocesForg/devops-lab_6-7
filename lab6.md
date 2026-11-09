# Лабораторная работа №8. Основы Docker

**Вариант:** 6 — «Свой микросервис + PostgreSQL»  
**Репозиторий:** [DocesForg/devops-lab_6-7](https://github.com/DocesForg/devops-lab_6-7)  
**Сервис:** Employee Service (Go)

## Цель работы

Изучить основы Docker, освоить управление образами и контейнерами, публикацию портов, подключение томов и настройку сетевого взаимодействия. Подготовить микросервис из лабораторной работы №6 к запуску в контейнере и рассмотреть его совместное развёртывание с PostgreSQL.

## 1. Исходный микросервис

В качестве приложения используется Employee Service, разработанный ранее. Сервис предоставляет операции над сотрудниками:

- создание сотрудника;
- получение сотрудника по идентификатору;
- изменение должности сотрудника;
- проверка состояния HTTP-сервера.

Для взаимодействия доступны два транспорта:

| Интерфейс | Адрес внутри контейнера | Назначение |
|---|---:|---|
| HTTP | 8080 | REST API |
| gRPC | 9090 | RPC API |

В проекте используется Go kit: HTTP- и gRPC-транспорты обращаются к общему сервисному слою через endpoints. Доступ к данным вынесен за интерфейс `EmployeeRepository`.

**Важно о текущем состоянии исходного кода:** на ветке `main` метод `main.go` создаёт `NewMemoryRepository()`. Это хранилище находится в оперативной памяти процесса. В текущей версии репозитория PostgreSQL-реализация репозитория и чтение параметров подключения к БД ещё не добавлены. Поэтому запуск PostgreSQL-контейнера сам по себе не означает, что приложение использует БД. Этот пункт отмечен отдельно, чтобы результаты отчёта соответствовали фактическому состоянию кода.

## 2. Подготовка Docker

Для выполнения практической части используется Docker Desktop с включённым WSL 2 backend (Windows) либо установленный Docker Engine (Linux).

Проверка доступности Docker:

```powershell
docker version
docker info
docker run hello-world
```

Команда `docker version` должна показать сведения о клиенте и сервере Docker. Команда `hello-world` проверяет, что Docker может загрузить образ и запустить контейнер.

**Скриншот 1 — проверка Docker и запуск hello-world**

> СКРИНШОТ ЗДЕСЬ

## 3. Сборка образа микросервиса

В корне репозитория находится `Dockerfile`. Он использует многостадийную сборку:

1. На стадии `builder` используется образ `golang:1.27-alpine`, копируются исходный код и файлы Go-модуля, затем компилируется исполняемый файл.
2. На финальной стадии используется компактный образ `alpine:3.22`. В него копируется только собранный бинарный файл и сертификаты.
3. Приложение запускается от непривилегированного пользователя `app`.
4. Объявлены порты 8080 (HTTP) и 9090 (gRPC).

Сборка выполняется из корня репозитория:

```powershell
docker build -t employee-service:v1 .
```

Проверка результата:

```powershell
docker images
docker image inspect employee-service:v1
```

Файл `.dockerignore` исключает из контекста сборки каталог Git, GitHub Actions, файл покрытия и Markdown-документы:

```text
.git
.github
coverage.out
*.md
```

Это уменьшает объём данных, передаваемых Docker daemon при сборке, и не включает в образ файлы, не нужные для запуска приложения.

**Скриншот 2 — сборка образа и вывод docker images**

> СКРИНШОТ ЗДЕСЬ

## 4. Запуск контейнера и проверка API

Запуск контейнера с публикацией обоих портов:

```powershell
docker run -d --name employee-api -p 8080:8080 -p 9090:9090 employee-service:v1
```

Проверка состояния и логов:

```powershell
docker ps
docker logs employee-api
docker port employee-api
```

Проверка HTTP health endpoint:

```powershell
curl.exe http://localhost:8080/healthz
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

Создание сотрудника:

```powershell
curl.exe -i -X POST http://localhost:8080/employees -H "Content-Type: application/json" -d "{\"full_name\":\"Ivan Ivanov\",\"position\":\"Engineer\",\"salary\":80000}"
```

В ответе возвращается идентификатор созданного сотрудника. Его следует подставить в запрос получения:

```powershell
curl.exe http://localhost:8080/employees/1
```

Изменение должности:

```powershell
curl.exe -i -X PATCH http://localhost:8080/employees/1/position -H "Content-Type: application/json" -d "{\"position\":\"Senior Engineer\"}"
```

Идентификатор в примерах приведён для случая, когда это первая созданная запись. В реальном запуске необходимо использовать ID из ответа на POST.

**Скриншот 3 — работа контейнера и проверка HTTP API**

> СКРИНШОТ ЗДЕСЬ

## 5. Создание Docker-сети

Для связи приложения и PostgreSQL создаётся пользовательская bridge-сеть:

```powershell
docker network create appnet
docker network ls
docker network inspect appnet
```

Пользовательская сеть позволяет контейнерам обращаться друг к другу по имени контейнера благодаря встроенному DNS Docker. Поэтому приложению следует указывать адрес БД как `db:5432`, а не использовать `localhost`: внутри контейнера `localhost` указывает на сам контейнер приложения.

**Скриншот 4 — созданная сеть appnet**

> СКРИНШОТ ЗДЕСЬ

## 6. Запуск PostgreSQL с постоянным томом

Для хранения данных PostgreSQL создаётся именованный том:

```powershell
docker volume create employee-db-data
docker volume ls
```

Запуск PostgreSQL 16 в сети `appnet` под именем `db`:

```powershell
docker run -d --name db --network appnet -e POSTGRES_DB=employees -e POSTGRES_USER=employee_user -e POSTGRES_PASSWORD=change_me -v employee-db-data:/var/lib/postgresql/data -p 5432:5432 postgres:16
```

Проверка запуска:

```powershell
docker ps
docker logs db
docker volume inspect employee-db-data
```

Подключение к БД и создание таблицы:

```powershell
docker exec -it db psql -U employee_user -d employees
```

SQL-команды в консоли PostgreSQL:

```sql
CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    position TEXT NOT NULL,
    salary NUMERIC(12, 2) NOT NULL CHECK (salary >= 0),
    hired_on BIGINT NOT NULL
);

INSERT INTO employees (full_name, position, salary, hired_on)
VALUES ('Ivan Ivanov', 'Engineer', 80000, EXTRACT(EPOCH FROM NOW())::BIGINT);

SELECT * FROM employees;
\q
```

Именованный том хранит файлы БД вне жизненного цикла контейнера. При удалении и повторном создании контейнера с тем же томом данные должны сохраниться.

Проверка сохранности данных:

```powershell
docker rm -f db
docker run -d --name db --network appnet -e POSTGRES_DB=employees -e POSTGRES_USER=employee_user -e POSTGRES_PASSWORD=change_me -v employee-db-data:/var/lib/postgresql/data -p 5432:5432 postgres:16
docker exec db psql -U employee_user -d employees -c "SELECT * FROM employees;"
```

**Скриншот 5 — PostgreSQL, таблица employees и сохранение записи после пересоздания контейнера**

> СКРИНШОТ ЗДЕСЬ

## 7. Подключение микросервиса к PostgreSQL

По условию варианта 6 приложение должно использовать PostgreSQL через слой репозитория. В проекте уже определён интерфейс `EmployeeRepository`, поэтому для выполнения требования необходимо добавить реализацию этого интерфейса на основе PostgreSQL и передавать её в `NewEmployeeService` вместо `NewMemoryRepository()`.

Параметры подключения следует получать из переменных окружения, например:

| Переменная | Значение для Docker-сети |
|---|---|
| DB_HOST | `db` |
| DB_PORT | `5432` |
| DB_NAME | `employees` |
| DB_USER | `employee_user` |
| DB_PASSWORD | пароль, заданный при запуске PostgreSQL |

После добавления PostgreSQL-репозитория контейнер приложения должен запускаться в той же сети:

```powershell
docker run -d --name api --network appnet -p 8080:8080 -p 9090:9090 -e DB_HOST=db -e DB_PORT=5432 -e DB_NAME=employees -e DB_USER=employee_user -e DB_PASSWORD=change_me employee-service:v1
```

Проверка DNS-разрешения имени БД из контейнера приложения (если в образ приложения добавлена утилита `nslookup` или `ping`, либо используется отдельный диагностический контейнер):

```powershell
docker run --rm --network appnet alpine:3.22 sh -c "apk add --no-cache bind-tools >/dev/null && nslookup db"
```

Проверка сетевого подключения к PostgreSQL:

```powershell
docker run --rm --network appnet postgres:16 pg_isready -h db -p 5432 -U employee_user -d employees
```

**Статус выполнения этого пункта в текущей версии репозитория:** команды демонстрируют настройку контейнеров и сети, но приложение пока не использует указанные переменные окружения и продолжает создавать in-memory репозиторий. Для полного соответствия варианту необходимо реализовать PostgreSQL-репозиторий, пересобрать образ и повторить проверку создания и получения сотрудника через API с подтверждением записи SQL-запросом в PostgreSQL.

**Скриншот 6 — проверка доступности PostgreSQL из сети appnet**

> СКРИНШОТ ЗДЕСЬ

**Скриншот 7 — создание сотрудника через API и проверка записи в PostgreSQL**

> СКРИНШОТ ЗДЕСЬ

## 8. Очистка ресурсов

После завершения проверки можно остановить и удалить контейнеры:

```powershell
docker rm -f employee-api
docker rm -f db
docker network rm appnet
```

Том `employee-db-data` намеренно не удаляется, чтобы сохранить данные. Если данные больше не нужны, том можно удалить отдельно:

```powershell
docker volume rm employee-db-data
```

Просмотр использования дискового пространства:

```powershell
docker system df
```

**Скриншот 8 — использование дискового пространства Docker**

> СКРИНШОТ ЗДЕСЬ

## 9. Вывод

В ходе работы рассмотрена контейнеризация Go-микросервиса Employee Service. Изучены сборка Docker-образа, публикация HTTP- и gRPC-портов, запуск и диагностика контейнеров, создание пользовательской сети и именованного тома PostgreSQL. Рассмотрен способ организации связи приложения с БД по DNS-имени контейнера `db` и сохранения данных при пересоздании контейнера.

В исходной версии проекта используется `MemoryRepository`, поэтому интеграция PostgreSQL на уровне приложения остаётся необходимым этапом для полного выполнения индивидуального задания варианта 6. Сам факт запуска PostgreSQL-контейнера не подтверждает выполнение операций чтения и записи через репозиторий микросервиса.
