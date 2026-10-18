# GameStats CS2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Реализовать SIS1 простым учебным кодом: анализ CS2 из JSON, задачи в памяти, состояния, ошибки и обязательные тесты.

**Architecture:** Один package main в sis1. Обычная функция считает статистику; TaskStore хранит задачи в map; main последовательно читает файлы и печатает отчёты через io.Writer.

**Tech Stack:** Go 1.27, только стандартная библиотека: encoding/json, io, os, fmt, errors, math, testing; strings.Builder для проверки вывода.

**Spec:** [Спецификация GameStats CS2](../specs/2026-10-18-gamestats-cs2-design.md); [исходное задание SIS1](../../requirements/SIS_CP1_SPECIFICATION.md).

## Global Constraints

- Из задания: `Use **Go 1.27**.` В sis1/go.mod: module gamestats, go 1.27.0. Сейчас локально установлен 1.26.3.
- `Expected application errors must be returned as errors. Do not use panic for normal application-level errors.`
- `Use at least one table-driven test where appropriate.` Использовать стандартный testing.
- Один JSON-файл — одна задача; данные и отчёты только в памяти; последовательная обработка; только стандартная библиотека.
- Вход, типы, имена констант, сигнатуры и формулы заданы в Spec. Все четыре JSON-поля обязательны, счётчики неотрицательные. Ничьи входят в знаменатель win rate.
- **Требование пользователя:** писать максимально просто, по примерам лекций, как студент, который только учит Go. Это обязательное ограничение реализации.
- Go-команды выполнять из sis1, Git-команды — из корня репозитория.

## Учебный стиль и источники

| Материал | Страницы и образец | Применение |
|---|---|---|
| [Week 2](</Users/grandmaster/golang/WEEK 2. ARRAYS and SLICES (1).pdf>) | стр. 2–3: slices, append, make, copy | []Match и отдельные копии входных данных |
| [Week 3](</Users/grandmaster/golang/WEEK 3. Strings and Maps (1).pdf>) | стр. 8–10: запись в map и поиск с ok | map[int]Task, поиск по ID |
| [Week 4](</Users/grandmaster/golang/WEEK 4. Structs and Methods.pdf>) | стр. 1: Student; стр. 10: Counter и pointer receiver | Match, Report, Task, методы изменения состояния |
| [Week 5](</Users/grandmaster/golang/WEEK 5. Interfaces and Generics (1).pdf>) | стр. 5: writeReport(io.Writer), os.Stdout, strings.Builder; стр. 18: сначала конкретные типы | интерфейс для реального вывода отчёта, обычные функции |

Писать обычные for/range, if и switch, понятные имена и явные if err != nil. Короткие функции должны делать одно понятное действие; проверки в тестах писать раздельно. Не вводить generics, фабрики, слои repository/service, callback-функции, подменяемые анализаторы и общие тестовые конструкторы. Достаточно одного package main в нескольких файлах.

JSON и testing необходимы по согласованному вводу и заданию, но в загруженных Week 2–5 JSON и unit-тесты не разобраны. Эти небольшие дополнения использовать через стандартную библиотеку и пояснить в README. Указатели в закрытой структуре JSON нужны только для различения отсутствующего поля и нуля.

## Review Focus

- Нулевые показатели разрешены; отсутствие поля и null отклоняются — задача 3.
- После массива разрешены пробелы, но не второй JSON-документ — задача 3.
- Ничьи и ноль смертей дают правильные дробные показатели без деления на смерти — задача 1.
- Изменение исходного среза или полученной задачи не меняет данные в map — задача 2.
- Реальное переполнение суммы даёт failed и ошибку; повторная обработка запрещена — задачи 1–2.

## Карта файлов

| Файл | Ответственность |
|---|---|
| sis1/go.mod | Модуль gamestats, Go 1.27.0 |
| sis1/stats.go, sis1/stats_test.go | Match, Report, проверка и расчёты |
| sis1/tasks.go, sis1/tasks_test.go | Task, TaskStore, map, ID, состояния, копирование |
| sis1/json.go, sis1/json_test.go | Декодирование сессии |
| sis1/main.go, sis1/main_test.go | Запуск и понятный текстовый отчёт |
| sis1/examples/session.json, sis1/examples/draw.json | Контрольная сессия и нулевой draw из Spec |
| sis1/README.md, корневой README.md | Описание проекта, запуск и ссылка на SIS1 |

### Task 1: Структуры и расчёты обычной функцией

**Files:** Create sis1/go.mod, sis1/stats.go, sis1/stats_test.go.

**Interfaces:**
- Consumes: вход и контрольный отчёт из Spec.
- Produces: Match, Report, строковые ResultWin/ResultLoss/ResultDraw, validateMatches([]Match) error, AnalyzeMatches([]Match) (Report, error).
- Tests: package main; небольшой helper sampleMatches() []Match возвращает новый литерал трёх контрольных матчей. Другие тестовые данные задавать обычными литералами.

- [ ] **Step 1: Подготовить модуль.** Выбрать/установить Go 1.27.x; GOTOOLCHAIN=local go version должен показать эту версию. Создать go.mod с module gamestats и go 1.27.0.
- [ ] **Step 2: Написать тесты до расчётов.** TestAnalyzeMatches проверяет весь контрольный отчёт: counts 3/2/1/0, totals 48/36/18, averages 16/12/6, win rate 200.0/3.0. Пример простых раздельных проверок:

```go
report, err := AnalyzeMatches(sampleMatches())
if err != nil {
    t.Fatal(err)
}
if report.TotalKills != 48 {
    t.Errorf("total kills: got %d, want 48", report.TotalKills)
}
if math.Abs(report.WinRate-200.0/3.0) > 1e-9 {
    t.Errorf("unexpected win rate: %f", report.WinRate)
}
```

  Table-driven TestAnalyzeCases: один draw с нулями → одна ничья и остальные показатели 0; win + draw с нулями → win rate 50; win(1,0,0) + loss(0,0,0) → averages 0.5/0/0 и win rate 50. TestAnalyzeInvalidInput отклоняет nil, пустой срез, пустой/неверный исход и каждый отрицательный счётчик. TestAnalyzeOverflow проверяет math.MaxInt + 1 отдельно для kills/deaths/assists: нужна ошибка, а не отрицательная сумма. В проверках ошибок достаточно err == nil / err != nil.
- [ ] **Step 3: Проверить падение.** go test . -run '^TestAnalyze' -v → FAIL из-за отсутствующих типов или функции.
- [ ] **Step 4: Реализовать stats.go.** Валидация обычными условиями, подсчёты одним for range и switch по исходу. До каждого сложения сравнивать счётчик с math.MaxInt минус текущая сумма. Затем явно посчитать float64-средние и win rate. Возвращать Report и обычный error, без собственного интерфейса анализатора.
- [ ] **Step 5: Проверить успех.** go test . -run '^TestAnalyze' -v → PASS.
- [ ] **Step 6: Зафиксировать.** git add sis1/go.mod sis1/stats.go sis1/stats_test.go; git commit -m "feat(sis1): add simple match statistics".

### Task 2: Хранилище и методы состояния

**Files:** Create sis1/tasks.go, sis1/tasks_test.go.

**Interfaces:**
- Consumes: Match, Report, validateMatches и AnalyzeMatches.
- Produces: Task, TaskStore, четыре Status-константы; NewTaskStore() *TaskStore; методы Create([]Match) (Task, error), Get(int) (Task, error), Process(int) (Task, error) на *TaskStore; (*Task).start() error; copyTask(Task) Task.
- TaskStore содержит map[int]Task и nextID. Report хранится в Task значением, поэтому копируется вместе с ним; copyTask отдельно копирует только срез Input через make/copy.

- [ ] **Step 1: Написать падающие тесты.** TestCreateTask проверяет ID 1, pending, Input и нулевой Report; второе успешное создание получает ID 2. Table-driven TestCreateInvalid проверяет доменные ошибки из задачи 1: err != nil, задача не появилась, следующий успешный Create получает ID 1. TestGetTask возвращает созданную задачу; TestTaskNotFound проверяет ошибки Get/Process для неизвестного ID.

  TestTaskStart непосредственно проверяет метод, используемый в Process:

```go
task := Task{Status: StatusPending}
if err := task.start(); err != nil {
    t.Fatal(err)
}
if task.Status != StatusProcessing {
    t.Errorf("got status %s, want processing", task.Status)
}
if err := task.start(); err == nil {
    t.Error("expected error when starting twice")
}
```

  TestProcessTask проверяет completed и полный контрольный отчёт после Process и повторного Get. TestProcessOverflow создаёт задачу с допустимыми по отдельности math.MaxInt и 1, затем проверяет ошибку Process, failed и нулевой Report. TestRepeatProcessing проверяет запрет повторной обработки completed/failed и неизменность сохранённых данных.

  TestIndependentTasks создаёт контрольную сессию и нулевой draw: разные ID, результаты не смешиваются, вторая остаётся pending до своей обработки. TestTaskCopies изменяет исходный срез и данные возвращённых Create/Get/Process задач: новый Get должен вернуть исходные сохранённые Input, Status и Result. Никаких spy-анализаторов и callback-функций.
- [ ] **Step 2: Проверить падение.** go test . -run '^(TestCreate|TestGet|TestTask|TestProcess|TestRepeat|TestIndependent)' -v → FAIL до реализации хранилища.
- [ ] **Step 3: Реализовать tasks.go.** make(map[int]Task), nextID=1; поиск task, ok := s.tasks[id]. Create сначала проверяет и копирует ввод. start проверяет pending и присваивает processing; Process сохраняет это состояние до AnalyzeMatches, затем сохраняет completed с Report либо failed с нулевым Report. Все результаты методов возвращать через copyTask. Ошибки — простые errors.New/fmt.Errorf; методы изменения данных — с pointer receiver.
- [ ] **Step 4: Проверить успех.** go test . -v → PASS, включая расчёты.
- [ ] **Step 5: Зафиксировать.** git add sis1/tasks.go sis1/tasks_test.go; git commit -m "feat(sis1): add tasks and state transitions".

### Task 3: JSON с понятными проверками

**Files:** Create sis1/json.go, sis1/json_test.go.

**Interfaces:**
- Consumes: Match, Result-константы и validateMatches.
- Produces: DecodeSession(r io.Reader) ([]Match, error). Функция только читает матчи, задачи не создаёт.

- [ ] **Step 1: Написать падающие тесты.** Table-driven TestDecodeSession принимает контрольный JSON, все нули, каждый исход и пробелы после массива. Сравнить длину среза и поля Match обычными проверками. TestDecodeSessionInvalid отклоняет пустой файл, [], null, {}, обрезанный JSON, [null], пропуск/null каждого обязательного поля, killz вместо kills, неверный/пустой исход, строку/дробное/отрицательное число, число вне int, второй документ. Вход для тестов — strings.NewReader с небольшими JSON-строками.
- [ ] **Step 2: Проверить падение.** go test . -run '^TestDecodeSession' -v → FAIL до появления функции.
- [ ] **Step 3: Реализовать json.go.** Одна закрытая структура с *string и тремя *int, Decoder с DisallowUnknownFields. Проверить nil-поля простыми if, преобразовать в []Match и вызвать validateMatches. После первого массива повторный Decode должен вернуть io.EOF. При ошибке вернуть понятный error с индексом матча, если он известен; сложные цепочки обёрток не нужны. Короткий комментарий поясняет, почему здесь указатели, а в Match обычные int.
- [ ] **Step 4: Проверить успех.** go test . -v → PASS.
- [ ] **Step 5: Зафиксировать.** git add sis1/json.go sis1/json_test.go; git commit -m "feat(sis1): read sessions from JSON".

### Task 4: Запуск и интерфейс по примеру лекции

**Files:** Create sis1/main.go, sis1/main_test.go, sis1/examples/session.json, sis1/examples/draw.json, sis1/README.md; Modify корневой README.md.

**Interfaces:**
- Consumes: DecodeSession, NewTaskStore, Create, Process, Get, Task.
- Produces: run(paths []string, out io.Writer) error; writeReport(w io.Writer, task Task) error; main передаёт os.Args[1:] и os.Stdout, печатает ошибку в stderr и завершает процесс с кодом 1.
- Использование io.Writer соответствует Week 5, стр. 5 и выполняет требование SIS1 об интерфейсе.

- [ ] **Step 1: Написать падающие тесты.** TestWriteReport использует strings.Builder и проверяет текст из Spec, включая Win rate: 66.67% и Average kills: 16.00; для Task без completed ожидается ошибка и пустой вывод. TestWriteReportError передаёт уже закрытый временный файл: ошибка записи должна вернуться. TestRunMultipleSessions записывает два входа в t.TempDir, вызывает run и проверяет два отчёта с ID 1/2 и правильными значениями. Table-driven TestRunInvalidInput: нет аргументов, файл не существует, JSON неверный, ввод невалиден — ошибка; ошибка конкретного файла содержит его путь. TestRunStopsBeforeOutput: корректный и неверный файлы дают ошибку и пустой stdout.
- [ ] **Step 2: Проверить падение.** go test . -run '^(TestWriteReport|TestRun)' -v → FAIL до реализации запуска.
- [ ] **Step 3: Реализовать main.go и примеры.** Один TaskStore на запуск, обычный цикл по путям, os.Open, DecodeSession, закрытие файла и Create → Process → Get. Закрытие файла выполнять после чтения каждого входа, не откладывать все закрытия до конца длинного цикла. После успеха всех входов пройти по срезу завершённых задач и вызвать writeReport. Формат отчёта из Spec выводить через fmt.Fprintf с проверкой ошибки записи.

  session.json содержит три контрольных матча, draw.json — один draw с нулями. README объясняет запуск, JSON, формулы, состояния, хранение в памяти и связь с лекциями. Отдельно кратко разобрать make/copy, pointer receiver, task, ok и io.Writer — то, что нужно уметь объяснить на защите. Корневой README получает ссылку на sis1/README.md.
- [ ] **Step 4: Проверить весь проект.** Последовательно выполнить:

```bash
gofmt -w .
gofmt -l .
go build ./...
go vet ./...
go test ./...
go run . examples/session.json examples/draw.json
```

  gofmt ничего не перечисляет; build/vet/test успешны; запуск печатает два отчёта с контрольными значениями. go run . examples/does-not-exist.json должен дать ошибку в stderr и пустой stdout. Этот отсутствующий файл не создавать. На этапе планирования проверки приложения ещё не запускались.
- [ ] **Step 5: Зафиксировать.** git add sis1/main.go sis1/main_test.go sis1/examples sis1/README.md README.md; git commit -m "feat(sis1): add console reports and examples".

## Покрытие SIS1

| Обязательная категория тестов из §8 | Проверка |
|---|---|
| 1. Успешное создание | TestCreateTask |
| 2. Неверный ввод | TestCreateInvalid, TestDecodeSessionInvalid |
| 3. Получение задачи | TestGetTask |
| 4. Неизвестный ID | TestTaskNotFound |
| 5. Успешная обработка | TestProcessTask, TestRunMultipleSessions |
| 6. Правильный результат | TestAnalyzeMatches, TestAnalyzeCases |
| 7. Состояние после обработки | TestProcessTask, TestProcessOverflow |
| 8. Недопустимая операция | TestRepeatProcessing, TestTaskStart |
| 9. Независимые задачи | TestIndependentTasks, TestTaskCopies |

§3 задания: создание, ID, память, поиск, ошибки, состояния и независимость — задача 2; расчёты и отчёт — задачи 1–2. §4–5 — задачи 1 и 3; §6 — задача 2 и README; §7 — структуры/методы, map, срезы, ошибки и io.Writer в задачах 1–4; §9 — модуль в задаче 1 и итоговые команды/коммиты в задаче 4.

## Передача к реализации

Для небольшого учебного проекта рекомендован Native: последовательно писать код в текущем чате и проверить всё изменение отдельным обзором в конце. Альтернатива Subagent-driven допускается после выбора пользователем. До запуска реализации пользователь рассматривает обновлённый план. Приложение в рамках правки плана не создаётся.
