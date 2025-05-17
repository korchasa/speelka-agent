# Refactoring and Test Coverage Plan (internal/app_mcp, mcp_connector, mcp_server, types)

## Цели
- Уменьшить размер длинных функций, повысить читаемость и поддерживаемость.
- Выделить внутренние (приватные) функции для изоляции логики и простого unit-тестирования.
- Улучшить покрытие тестами edge-cases и ошибок без сложных моков.
- Не выносить функции в отдельные файлы — только приватные внутри текущих файлов.

---

## 1. internal/app_mcp/app.go

### Кандидаты для рефакторинга
- `HandleCall`
- `DispatchMCPCall`

### План действий
- Вынести проверки и преобразования в приватные функции:
  - `validateToolName(toolName string, config types.ConfigurationManagerSpec) error`
  - `extractUserInput(arguments map[string]interface{}, argName string) (string, error)`
  - `buildDirectCallResult(answer string, meta types.MetaInfo, err error) types.DirectCallResult`
- Вынести логику логирования в приватные функции (например, `logHandleCallStep`), если есть повторяющиеся вызовы.
- Покрыть приватные функции unit-тестами (edge-cases: пустые значения, невалидные типы, неправильные имена инструментов).
- Основные функции оставить короткими, только с "каркасом" вызова приватных функций.

---

## 2. internal/mcp_connector/mcp_connector.go

### Кандидаты для рефакторинга
- `ExecuteTool`
- `callToolWithTimeout`
- `filterAllowedTools`

### План действий
- Вынести:
  - Поиск клиента: `findServerAndClientForTool(toolName string) (string, client.MCPClient, error)`
  - Таймаут: `callToolWithTimeout(ctx, mcpClient, call, callTimeout) (*mcp.CallToolResult, error, bool)` (уже есть, но можно упростить)
  - Логирование: `logToolExecutionStart`, `logToolTimeout`, `logToolError` (оставить приватными)
  - Фильтрацию инструментов: `filterAllowedTools(serverID string, tools []mcp.Tool, srvCfg types.MCPServerConnection) []mcp.Tool` (оставить приватной)
- Покрыть приватные функции unit-тестами (например, фильтрация инструментов, обработка таймаутов, edge-cases поиска клиента).
- Основные функции сделать "тонкими" — только orchestration.

---

## 3. internal/mcp_server/mcp_server.go

### Кандидаты для рефакторинга
- `Serve`
- `serveDaemon`
- `serveStdioWithContext`
- `buildTools`

### План действий
- Вынести проверки и инициализацию серверов в приватные функции:
  - `initSSEServer()`
  - `initStdioServer()`
- Вынести создание инструментов:
  - `buildMainTool()`
  - `buildLoggingTool()`
- Покрыть приватные функции unit-тестами (например, создание инструментов, обработка ошибок инициализации).
- Основные функции оставить "тонкими".

---

## 4. internal/types

### Кандидаты для рефакторинга
- Вспомогательные функции (например, парсинг уровней логов, преобразования)

### План действий
- Вынести парсинг и преобразования в приватные функции (если есть длинные функции).
- Покрыть edge-cases unit-тестами (например, невалидные уровни логов).

---

## Общие рекомендации
- Все новые приватные функции должны быть покрыты unit-тестами (edge-cases, ошибки, пустые значения).
- Не выносить функции в отдельные файлы — только приватные внутри текущих файлов.
- После рефакторинга — убедиться, что все тесты проходят и покрытие увеличилось.
- Документировать все изменения в этом whiteboard.

---

## Следующие шаги
1. Начать с internal/app_mcp/app.go: выделить приватные функции, добавить тесты.
2. Перейти к internal/mcp_connector/mcp_connector.go: аналогично.
3. Затем internal/mcp_server/mcp_server.go.
4. Проверить internal/types на предмет вспомогательных функций.
5. После каждого этапа — обновить этот план и зафиксировать прогресс.
