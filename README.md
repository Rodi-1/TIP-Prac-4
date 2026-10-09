# Практическая работа 4 — CRUD-сервис списка задач на chi

Сортов Семён Александрович, ЭФМО-01-26. Дата: 09.10.2026.

CRUD-сервис на chi v5.3.2 с задачами в памяти и сохранением в tasks.json. Есть все пять операций CRUD, title 3–100 символов, пагинация page/limit, фильтр done, логирование и CORS. Доступны `/api/tasks` и версия `/api/v1/tasks`. Требования: Go 1.23+ и Git. tasks.json создаётся при первом изменении; приложение читает его при запуске. Файл хранится в текущей рабочей папке, поэтому запускать нужно из корня проекта.

## Запуск и сборка

Команды выполняются из этой папки. Каждая работа запускается отдельно: по умолчанию все используют порт 8080.

```powershell
go mod download
go run .
# Сборка и проверки
go build -o ./bin/server.exe .
go fmt ./...
go vet ./...
go test ./...
# После остановки go run
.\bin\server.exe
```

## Примеры запросов

```powershell
curl.exe -i http://localhost:8080/health
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/tasks -ContentType "application/json" -Body '{"title":"Выучить chi"}'
curl.exe -i http://localhost:8080/api/tasks
curl.exe -i http://localhost:8080/api/tasks/1
Invoke-RestMethod -Method Put -Uri http://localhost:8080/api/tasks/1 -ContentType "application/json" -Body '{"title":"Выучить chi глубже","done":true}'
curl.exe -i "http://localhost:8080/api/tasks?done=true&page=1&limit=10"
curl.exe -i -X DELETE http://localhost:8080/api/tasks/1
curl.exe -i http://localhost:8080/api/v1/tasks
```

## Конфигурация

При пагинации задачи отсортированы по ID. Без page/limit возвращается весь список; при указании одного параметра значения по умолчанию — page=1 и limit=10. done принимает true или false; фильтр применяется до пагинации. Некорректные данные → 400, неизвестная задача → 404, ошибка сохранения → 500, DELETE → 204 без тела. tasks.json не включён в Git, чтобы не передавать текущие данные.

## Структура

```text
.gitignore
go.mod
go.sum
internal/task/handler.go
internal/task/model.go
internal/task/repo.go
main.go
pkg/middleware/cors.go
pkg/middleware/logger.go
testing.md
bin/server.exe (собранная программа, исключена из Git)
```
