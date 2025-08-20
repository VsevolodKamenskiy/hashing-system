# Система хеширования

Этот проект содержит два микросервиса на Go для вычисления и хранения
SHA3‑256 хешей. Docker Compose управляет сервисами и вспомогательной
инфраструктурой.

## Сервисы

### service1 – gRPC калькулятор хешей (stateless сервис)

`service1` предоставляет gRPC API `HasherService` на порту `50051`.
Метод `CalculateHashes` принимает список строк и возвращает их хеши SHA3‑256
в том же порядке.

Пример с использованием [grpcurl](https://github.com/fullstorydev/grpcurl):

```bash
grpcurl -plaintext \
-d '{"strings":["foo","bar"]}' \
localhost:50051 hasher.HasherService/CalculateHashes
```

Пример ответа:

```json
{
"hashes": [
"3338be0...", // hash of "foo"
"fcde2b2..."  // hash of "bar"
]
}
```

### service2 – HTTP API и хранилище (stateful севрис)

`service2` предоставляет HTTP API на порту `8080` и сохраняет хеши в
PostgreSQL. Для вычисления хешей он обращается к `service1` по gRPC.

Эндпоинты:

* `POST /send` – тело: JSON массив строк, возвращает массив объектов `{id, hash}`.
* `GET /check?ids=1&ids=2` – возвращает сохранённые хеши для указанных ID,
`204` если ничего не найдено.

Пример запроса:

```bash
# вычисление и сохранение хешей
curl -X POST http://localhost:8080/send \
-H 'Content-Type: application/json' \
-d '["hello","world"]'

# получение хешей по ID
curl "http://localhost:8080/check?ids=1&ids=2"
```

## Используемые технологии в проекте

- **Go** – язык реализации обоих сервисов.
- **gRPC** – удалённые вызовы между `service2` и `service1`.
- **HTTP/REST** – внешний API, предоставляемый `service2`.
- **PostgreSQL** – постоянное хранилище для рассчитанных хешей.
- **Docker и Docker Compose** – контейнеризация и оркестрация сервисов и зависимостей.
- **Consul** – сервис-дискавери.
- **Graylog** – централизованный логинг.
- **Prometheus** – сбор метрик.

## Запуск

Используйте Docker Compose для сборки и запуска обоих сервисов и их зависимостей
(PostgreSQL, Consul, Graylog и т.д.):

```bash
docker-compose up --build
```

`service1` будет доступен на `localhost:50051`, а `service2` будет слушать
по адресу `http://localhost:8080`.
