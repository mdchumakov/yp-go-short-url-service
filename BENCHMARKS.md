# Бенчмарки производительности

Этот документ описывает бенчмарки, доступные в проекте для измерения производительности ключевых компонентов системы.

## Запуск бенчмарков

### Все бенчмарки

Для запуска всех бенчмарков:

```bash
go test -bench=. -benchmem ./...
```

### Конкретный пакет

Для запуска бенчмарков в конкретном пакете:

```bash
go test -bench=. -benchmem ./internal/service/urls/shortener
```

### Конкретный бенчмарк

Для запуска конкретного бенчмарка:

```bash
go test -bench=BenchmarkShortenURLBase62 -benchmem ./internal/service/urls/shortener
```

## Доступные бенчмарки

### Генерация коротких URL

#### BenchmarkShortenURLBase62
Бенчмарк для функции генерации коротких URL из длинного URL.

```bash
go test -bench=BenchmarkShortenURLBase62 -benchmem ./internal/service/urls/shortener
```

#### BenchmarkToBase62
Бенчмарк для функции конвертации числа в base62.

```bash
go test -bench=BenchmarkToBase62 -benchmem ./internal/service/urls/shortener
```

#### BenchmarkShortenURLBase62Parallel
Параллельный бенчмарк для генерации коротких URL (тестирует конкурентность).

```bash
go test -bench=BenchmarkShortenURLBase62Parallel -benchmem ./internal/service/urls/shortener
```

### Сервис сокращения URL

#### BenchmarkShortURL_NewURL
Бенчмарк для создания нового короткого URL через сервис.

```bash
go test -bench=BenchmarkShortURL_NewURL -benchmem ./internal/service/urls/shortener
```

#### BenchmarkShortURL_ExistingURL
Бенчмарк для случая, когда URL уже существует в базе данных.

```bash
go test -bench=BenchmarkShortURL_ExistingURL -benchmem ./internal/service/urls/shortener
```

#### BenchmarkShortURLsByBatch
Бенчмарк для пакетного создания URL (10 URL).

```bash
go test -bench=BenchmarkShortURLsByBatch -benchmem ./internal/service/urls/shortener
```

#### BenchmarkShortURLsByBatch_Large
Бенчмарк для большого пакета URL (100 URL).

```bash
go test -bench=BenchmarkShortURLsByBatch_Large -benchmem ./internal/service/urls/shortener
```

### Сервис извлечения URL

#### BenchmarkExtractLongURL
Бенчмарк для извлечения длинного URL по короткому.

```bash
go test -bench=BenchmarkExtractLongURL -benchmem ./internal/service/urls/extractor
```

#### BenchmarkExtractUserURLs
Бенчмарк для извлечения всех URL пользователя.

```bash
go test -bench=BenchmarkExtractUserURLs -benchmem ./internal/service/urls/extractor
```

## Параметры бенчмарков

### -benchmem
Показывает статистику по использованию памяти:
- `B/op` - байт на операцию
- `allocs/op` - количество аллокаций на операцию

### -benchtime
Устанавливает время выполнения бенчмарка:

```bash
go test -bench=. -benchtime=10s ./internal/service/urls/shortener
```

### -count
Количество запусков каждого бенчмарка:

```bash
go test -bench=. -count=5 ./internal/service/urls/shortener
```

### -cpu
Количество используемых CPU:

```bash
go test -bench=. -cpu=1,2,4,8 ./internal/service/urls/shortener
```

## Интерпретация результатов

### Пример вывода

```
BenchmarkShortenURLBase62-8    1000000    1200 ns/op    256 B/op    2 allocs/op
```

- `BenchmarkShortenURLBase62-8` - имя бенчмарка и количество CPU
- `1000000` - количество итераций
- `1200 ns/op` - среднее время выполнения одной операции в наносекундах
- `256 B/op` - среднее количество байт, выделенных на операцию
- `2 allocs/op` - среднее количество аллокаций на операцию

### Анализ производительности

1. **Время выполнения (ns/op)**
   - Меньше = лучше
   - Сравнивайте результаты до и после оптимизаций

2. **Использование памяти (B/op)**
   - Меньше = лучше
   - Обратите внимание на неожиданно большие значения

3. **Количество аллокаций (allocs/op)**
   - Меньше = лучше
   - Большое количество аллокаций может указывать на проблемы производительности

## Сравнение результатов

Для сравнения результатов до и после изменений используйте `benchstat`:

```bash
# Установка benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Сбор результатов до изменений
go test -bench=. -benchmem ./internal/service/urls/shortener > scripts/profiles/before.txt

# Сбор результатов после изменений
go test -bench=. -benchmem ./internal/service/urls/shortener > scripts/profiles/after.txt

# Сравнение
benchstat before.txt after.txt
```
