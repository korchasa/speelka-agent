# ADR: Устранение дублирования конфигурационных структур

## Goal
Избавить проект от дублирующихся структур конфигурации и централизовать их определение.

## Overview
- В проекте определены похожие структуры для конфигурации в `internal/types/configuration.go` (`Configuration`, `ConfigAgent`, `AgentLLMConfig`, `LLMRetryConfig` и др.) и бизнес-структуры в `internal/types` (`AgentConfig`, `LLMConfig`, `MCPServerConfig`, `MCPConnectorConfig`, `RetryConfig` и др.).
- Для преобразования между ними используются методы `ToAgentConfig`, `ToLLMConfig`, `ToMCPServerConfig`, `ToMCPConnectorConfig`, что усложняет сопровождение и увеличивает риск рассинхронизации полей.

## Definition of Done
- Создан пакет `internal/config` с единым набором структур конфигурации (raw и бизнес).
- Удалены дублирующие типы из `internal/types/configuration.go` и бизнес-типов.
- Добавлены теги `json`/`yaml` к бизнес-структурам.
- Менеджер конфигурации обновлён для загрузки напрямую в новые типы.
- Методы `To*Config` и связанные тесты удалены.
- Все компоненты и тесты используют единые типы конфигурации.
- Все тесты проходят (`./run test`) и проверка (`./run check`) выполнена.

## Solution
1. Создать пакет `internal/config` и перенести туда:
   - Структуры raw-конфига: `Configuration`, `RuntimeConfig`, `RuntimeLogConfig`, `RuntimeTransportConfig`, `RuntimeStdioConfig`, `RuntimeHTTPConfig`, `ConfigAgent`, `AgentChatConfig`, `AgentToolConfig`, `AgentLLMConfig`, `AgentConnectionsConfig`, `LLMRetryConfig`, `ConnectionRetryConfig`.
   - Бизнес-структуры: `AgentConfig`, `LLMConfig`, `MCPServerConfig`, `MCPConnectorConfig`, `HTTPConfig`, `StdioConfig`, `RetryConfig`.
2. Аннотировать все структуры тегами `json`/`yaml` для прямого разбора.
3. Удалить методы `To*Config` и их тесты.
4. Переписать менеджер конфигурации (`internal/configuration/manager.go`) на загрузку и merge напрямую в новые структуры из `internal/config`.
5. Обновить все места использования типов и тесты в `internal/*`.
6. Запустить `./run test` и `./run check`.

## Consequences
- Единое определение конфигурации упрощает поддержку и расширение.
- Уменьшается дублирование кода и риск рассинхронизации полей.
- Потребуется масштабный рефакторинг и обновление большого числа зависимостей.

## Alternative Solutions

1. Использовать type alias в текущем пакете `internal/types`:
   - Например, `type AgentConfig = ConfigAgent` и т.п., чтобы не дублировать структуру, а лишь переименовывать.
   - Не требует новых пакетов, но сохраняет одно определение.

2. Объединить raw и business-структуры в одном типе:
   - Дать структурам теги `json`/`yaml` и использовать их напрямую во всех слоях.
   - Убрать методы `To*Config`, оставить один набор типов с валидацией при загрузке.

3. Воспользоваться встраиванием (struct embedding):
   - Вставить raw-конфиг как вложенную структуру в бизнес-тип, расширив его методами.
   - Позволяет хранить все поля в одном месте, но логика слоёв отделена.

4. Автоматическая генерация кода (go generate + шаблоны):
   - Объявить одну «истинную» структуру и генерировать алиасы/обёртки для разных контекстов.
   - Минимизирует ручное дублирование, но добавляет дополнительную сборку.

5. Сконцентрировать все типы конфигурации в существующем пакете `internal/configuration`:
   - Перенести бизнес-структуры в `internal/configuration`, используя его без создания нового пакета.
   - Пакет `internal/types` оставить для интерфейсов и примитивов.
