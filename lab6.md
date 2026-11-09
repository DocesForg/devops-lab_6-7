# Лабораторная работа №8. Основы Docker

**Вариант:** 6 — «Свой микросервис + PostgreSQL»  
**Проект:** Employee Service (Go)  
**Репозиторий:** [DocesForg/devops-lab_6-7](https://github.com/DocesForg/devops-lab_6-7)

## Цель работы

Изучить принципы работы Docker, освоить управление контейнерами, портами, томами и сетями, научиться запускать готовые образы и подготовить собственный микросервис к контейнеризации.

## 1. Исходный микросервис

В качестве приложения используется Employee Service, разработанный в лабораторной работе №6. Он предоставляет операции создания сотрудника, получения сотрудника по идентификатору и изменения должности. Для взаимодействия доступны HTTP API на порту 8080 и gRPC API на порту 9090.

Архитектура разделена на транспортный, сервисный и репозиторный слои. HTTP- и gRPC-транспорты используют общие Go kit endpoints и бизнес-сервис. Для хранения данных реализованы два репозитория: `MemoryRepository` и `PostgresRepository`. В рамках варианта 6 контейнер приложения подключается к PostgreSQL.

## 2. Установка и проверка Docker

Для выполнения работы используется Docker Desktop с включённым WSL 2 backend (Windows) либо Docker Engine (Linux).

Проверка:

```powershell
docker version
docker info
docker run hello-world
```

Команда `docker version` должна показать сведения о клиенте и сервере Docker. При первом запуске `hello-world` Docker загружает образ, создаёт контейнер и выводит сообщение об успешном запуске.

**Скриншот 1 — docker version и docker run hello-world**

> СКРИНШОТ ЗДЕСЬ

## 3. Работа с образами

Загрузка образов nginx и PostgreSQL:

```powershell
docker pull nginx
docker pull nginx:1.25-alpine
docker pull postgres:16
docker images
docker inspect nginx:latest
```

Для демонстрации того, что используемый контейнером образ нельзя удалить, запускается nginx:

```powershell
docker run -d --name test-nginx nginx
docker rmi nginx
```

Docker сообщит, что образ используется контейнером. После проверки контейнер и дополнительный тег можно удалить:

```powershell
docker rm -f test-nginx
docker rmi nginx:1.25-alpine
```

**Скриншот 2 — docker images после загрузки nginx**

> СКРИНШОТ ЗДЕСЬ

## 4. Управление контейнерами

Запуск nginx в фоновом режиме и проверка его состояния:

```powershell
docker run -d --name my-nginx nginx
docker ps
docker ps -a
docker logs my-nginx
docker logs -f my-nginx
```

Остановка, повторный запуск и удаление:

```powershell
docker stop my-nginx
docker start my-nginx
docker restart my-nginx
docker stop my-nginx
docker rm my-nginx
```

Ключ `-d` запускает контейнер в фоне, а `--name` задаёт ему удобное имя. Команда `docker logs` показывает стандартный вывод контейнера.

**Скриншот 3 — docker ps и docker logs my-nginx**

> СКРИНШОТ ЗДЕСЬ

## 5. Публикация портов и работа внутри контейнера

Запуск веб-сервера nginx с публикацией внутреннего порта 80 на порт 8080 хоста:

```powershell
docker run -d --name web -p 8080:80 nginx
docker port web
curl.exe http://localhost:8080
docker exec web hostname
docker inspect web
```

Параметр `-p 8080:80` означает, что обращения к порту 8080 хоста перенаправляются на порт 80 контейнера. Команда `docker exec` позволяет выполнить команду внутри работающего контейнера.

**Скриншот 4 — открытие nginx через localhost:8080 и docker exec web hostname**

> СКРИНШОТ ЗДЕСЬ

## 6. Тома Docker и сохранение данных PostgreSQL

Создание именованного тома и запуск PostgreSQL 16:

```powershell
docker volume create pgdata
docker run -d --name pg-test -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=testdb -v pgdata:/var/lib/postgresql/data -p 5432:5432 postgres:16
docker volume ls
docker volume inspect pgdata
```

Создание таблицы и тестовой записи:

```powershell
docker exec -it pg-test psql -U postgres -d testdb
```

В консоли PostgreSQL:

```sql
CREATE TABLE volume_check (
    id SERIAL PRIMARY KEY,
    value TEXT NOT NULL
);
INSERT INTO volume_check (value) VALUES ('volume survives container removal');
SELECT * FROM volume_check;
\q
```

Удаление контейнера и повторный запуск с тем же томом:

```powershell
docker rm -f pg-test
docker run -d --name pg-test -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=testdb -v pgdata:/var/lib/postgresql/data -p 5432:5432 postgres:16
docker exec pg-test psql -U postgres -d testdb -c "SELECT * FROM volume_check;"
```

Таблица и запись сохраняются, потому что файлы базы данных расположены в именованном томе Docker, жизненный цикл которого не зависит от контейнера.

**Скриншот 5 — docker volume inspect pgdata и запись после пересоздания контейнера**

> СКРИНШОТ ЗДЕСЬ

## 7. Сети Docker

Создание пользовательской bridge-сети и запуск nginx в ней:

```powershell
docker network create mynet
docker run -d --name nginx1 --network mynet nginx
docker run --rm --network mynet alpine:3.22 sh -c "apk add --no-cache curl && curl http://nginx1"
docker network inspect mynet
```

Пользовательская bridge-сеть предоставляет DNS: контейнеры могут обращаться друг к другу по именам. Это надёжнее, чем фиксировать IP-адреса контейнеров.

**Скриншот 6 — docker network inspect mynet и запрос к nginx1 по имени**

> СКРИНШОТ ЗДЕСЬ

## 8. Собственный образ микросервиса

Для сборки используется Dockerfile проекта. Он реализует многостадийную сборку: на первом этапе компилируется Go-приложение, на финальный образ копируется только бинарный файл и сертификаты. Контейнер запускается от непривилегированного пользователя. Порты 8080 и 9090 доступны для HTTP и gRPC.

В корне проекта создан файл `.dockerignore`:

```text
.git
.github
coverage.out
*.md
```

Он исключает из контекста сборки Git-метаданные, конфигурацию GitHub Actions, файл покрытия и Markdown-документы. Это сокращает передаваемый контекст и не позволяет ненужным файлам попадать в сборку.

Сборка и запуск:

```powershell
docker build -t employee-service:v1 .
docker images
docker run -d --name employee-api -p 8080:8080 -p 9090:9090 employee-service:v1
curl.exe http://localhost:8080/healthz
```

Проверка API:

```powershell
curl.exe -i -X POST http://localhost:8080/employees -H "Content-Type: application/json" -d "{\"full_name\":\"Ivan Ivanov\",\"position\":\"Engineer\",\"salary\":80000}"
curl.exe http://localhost:8080/employees/1
```

Для сравнения контекста сборки можно временно выполнить сборку с отдельным пустым файлом исключений и сравнить строки `transferring context` в выводе `docker build`. Размер готового образа смотрят через `docker images employee-service`. Сам .dockerignore влияет прежде всего на контекст сборки; он не гарантирует уменьшение размера финального образа, если исключённые файлы и так не копируются в Dockerfile.

**Скриншот 7 — сборка employee-service и проверка HTTP API**

> СКРИНШОТ ЗДЕСЬ

## 9. Индивидуальное задание. Вариант 6 — микросервис + PostgreSQL

### 9.1. Реализация PostgreSQL-репозитория

В проект добавлен файл `postgres.go` с реализацией интерфейса `EmployeeRepository` на базе `database/sql` и драйвера `github.com/lib/pq`.

Реализованы операции:

- подключение к PostgreSQL;
- автоматическое создание таблицы `employees` при инициализации репозитория;
- добавление сотрудника с возвратом сгенерированного ID;
- получение сотрудника по ID;
- изменение должности;
- преобразование ситуации отсутствующей записи в `ErrNotFound`.

Параметры подключения задаются переменными окружения: `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_SSLMODE`. Если `DB_HOST` не задан, для локального запуска без БД используется `MemoryRepository`. В контейнерном сценарии задаётся `DB_HOST=db`.

При запуске приложение ожидает доступности PostgreSQL до 45 секунд, затем создаёт репозиторий и таблицу. Таким образом, таблица создаётся приложением автоматически, без ручного SQL.

### 9.2. Развёртывание двух контейнеров в одной сети

Для воспроизводимого запуска добавлен `compose.yaml`. Он описывает два сервиса:

- `db` — PostgreSQL 16 с именем контейнера `employee-db`, healthcheck и именованным томом `employee-db-data`;
- `api` — Employee Service, собираемый из Dockerfile. Он подключается к PostgreSQL по DNS-имени `db`.

Оба сервиса находятся в пользовательской сети `appnet`. Публикуются HTTP-порт 8080 и gRPC-порт 9090 приложения. Порт PostgreSQL не требуется публиковать на хост для взаимодействия контейнеров.

Запуск из корня репозитория:

```powershell
docker compose up --build -d
docker compose ps
docker compose logs api
```

Проверка, что API отвечает:

```powershell
curl.exe http://localhost:8080/healthz
```

Проверка доступности PostgreSQL из той же сети:

```powershell
docker run --rm --network appnet postgres:16 pg_isready -h db -p 5432 -U employee_user -d employees
```

Если имя сети отличается, его можно узнать командой `docker network ls`.

### 9.3. Проверка чтения и записи через API

Создание сотрудника:

```powershell
curl.exe -i -X POST http://localhost:8080/employees -H "Content-Type: application/json" -d "{\"full_name\":\"Ivan Ivanov\",\"position\":\"Engineer\",\"salary\":80000}"
```

В ответе возвращается ID. Получение записи (в примере ID равен 1; используйте фактическое значение из ответа):

```powershell
curl.exe http://localhost:8080/employees/1
```

Изменение должности:

```powershell
curl.exe -i -X PATCH http://localhost:8080/employees/1/position -H "Content-Type: application/json" -d "{\"position\":\"Senior Engineer\"}"
```

Проверка записи непосредственно в PostgreSQL:

```powershell
docker compose exec db psql -U employee_user -d employees -c "SELECT id, full_name, position, salary, hired_on FROM employees;"
```

Таким образом, запись создаётся через HTTP API, сохраняется в PostgreSQL и читается тем же микросервисом через репозиторий.

**Скриншот 8 — успешный запуск db и api в одной сети**

> СКРИНШОТ ЗДЕСЬ

**Скриншот 9 — создание и чтение сотрудника через API**

> СКРИНШОТ ЗДЕСЬ

**Скриншот 10 — SELECT из PostgreSQL с записью, созданной через API**

> СКРИНШОТ ЗДЕСЬ

### 9.4. Проверка сохранности данных

Остановка и удаление контейнеров без удаления тома:

```powershell
docker compose down
docker compose up -d
docker compose exec db psql -U employee_user -d employees -c "SELECT * FROM employees;"
```

Данные сохраняются, так как PostgreSQL хранит файлы БД в named volume `employee-db-data`. Команда `docker compose down` удаляет контейнеры и сеть, но не именованный том. Удаление данных выполняется отдельно через `docker compose down -v`.

**Скриншот 11 — запись в таблице после пересоздания контейнеров**

> СКРИНШОТ ЗДЕСЬ

## 10. Проверка и очистка ресурсов

Проверка занимаемого Docker пространства:

```powershell
docker system df
```

Очистка контейнеров, образов и сетей, созданных в работе:

```powershell
docker compose down
docker rm -f web nginx1 2>$null
docker network rm mynet 2>$null
docker rm -f pg-test 2>$null
docker volume rm pgdata
docker system df
```

Не следует использовать `docker system prune -a --volumes` без необходимости: эта команда может удалить неиспользуемые тома с данными.

**Скриншот 12 — docker system df до и после очистки**

> СКРИНШОТ ЗДЕСЬ

## 11. Вывод

В ходе лабораторной работы изучены основные операции Docker: загрузка образов, запуск и управление контейнерами, публикация портов, выполнение команд внутри контейнера, использование томов и пользовательских сетей. Микросервис Employee Service подготовлен к контейнеризации и запускается совместно с PostgreSQL в общей сети. Реализован PostgreSQL-репозиторий с автоматическим созданием таблицы и операциями создания, чтения и изменения сотрудника. Данные PostgreSQL сохраняются в именованном томе при пересоздании контейнера.
