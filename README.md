# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

## API

API документация доступна через интерактивный Swagger UI:

**Swagger UI:**
```text
http://localhost:8080/swagger/
```

### Основные маршруты для задач

- `POST /api/v1/tasks` - создать обычную задачу
- `GET /api/v1/tasks` - получить список задач
- `GET /api/v1/tasks/{id}` - получить задачу по ID
- `PUT /api/v1/tasks/{id}` - обновить задачу
- `DELETE /api/v1/tasks/{id}` - удалить задачу

### Маршруты для периодических задач

#### Управление правилами периодичности

- `POST /api/v1/recurrence-rules` - создать правило периодичности
- `GET /api/v1/recurrence-rules` - получить список всех правил (поддерживает `?enabled=true/false`)
- `GET /api/v1/recurrence-rules/{id}` - получить правило по ID
- `PUT /api/v1/recurrence-rules/{id}` - обновить правило
- `DELETE /api/v1/recurrence-rules/{id}` - удалить правило

#### Создание задач с периодичностью

- `POST /api/v1/tasks/with-recurrence` - создать периодическую задачу

## Функциональность периодических задач

Система поддерживает четыре типа периодичности задач:

### 1. Ежедневные (Daily)
Создает задачи каждый N-й день.

**Параметры:**
- `interval_days` (число): интервал между задачами (например, 1 = каждый день, 3 = каждый третий день)

**Пример:**
```json
{
  "name": "Ежедневный обход пациентов",
  "recurrence_type": "daily",
  "params": {
    "interval_days": 1
  }
}
```

### 2. Ежемесячные (Monthly)
Создает задачи на конкретные дни месяца.

**Параметры:**
- `month_days` (массив чисел 1-30): дни месяца для создания задач

**Пример:**
```json
{
  "name": "Отчетность",
  "recurrence_type": "monthly",
  "params": {
    "month_days": [1, 15, 30]
  }
}
```

### 3. На конкретные даты (Specific)
Создает задачи только на указанные даты.

**Параметры:**
- `specific_dates` (массив строк в формате "YYYY-MM-DD"): конкретные даты

**Пример:**
```json
{
  "name": "Специальные совещания",
  "recurrence_type": "specific",
  "params": {
    "specific_dates": ["2025-05-01", "2025-06-15", "2025-12-31"]
  }
}
```

### 4. Четные/Нечетные дни (Even/Odd)
Создает задачи только на четные или нечетные числа месяца.

**Типы:**
- `even` - четные дни (2, 4, 6, ..., 30)
- `odd` - нечетные дни (1, 3, 5, ..., 29, 31)

**Пример:**
```json
{
  "name": "Четные дни отчетность",
  "recurrence_type": "even",
  "params": {}
}
```

## Примеры использования


### Через curl

Для тестирования через командную строку используйте примеры ниже.

#### Пример 1: Создать периодическую задачу через существующее правило

```bash
curl -X POST http://localhost:8080/api/v1/recurrence-rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ежедневный обход",
    "recurrence_type": "daily",
    "params": {
      "interval_days": 1
    }
  }'
```

Затем создать задачу:
```bash
curl -X POST http://localhost:8080/api/v1/tasks/with-recurrence \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Обход пациентов",
    "description": "Проверить всех пациентов в отделении",
    "recurrence_rule_id": 1
  }'
```

### Пример 2: Создать задачу с inline правилом

```bash
curl -X POST http://localhost:8080/api/v1/tasks/with-recurrence \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Еженедельный отчет",
    "description": "Отчет по отделению",
    "recurrence_rule_data": {
      "name": "Еженедельно по понедельникам",
      "recurrence_type": "monthly",
      "params": {
        "month_days": [1, 8, 15, 22, 29]
      }
    }
  }'
```


### Graceful Shutdown
При остановке приложения:
1. Планировщик корректно останавливается
2. HTTP сервер закрывается с таймаутом 10 секунд
3. Активные локи в job_queue освобождаются (по истечении timeout'а lock)