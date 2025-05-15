# План рефакторинга Configuration с использованием TDD

1. Red: Написать тесты в `internal/types/configuration_test.go`:
   - Проверить Unmarshal YAML/JSON с новой inline-структурой `Configuration`.
   - Убедиться, что поля `Runtime` и `Agent` правильно разбираются.
2. Green: Реализовать inline-структуры:
   - Заменить `RuntimeConfig`, `ConfigAgent` и вложенные типы на анонимные inline-структуры в `Configuration`.
   - Сохранить существующие теги `json`/`yaml`.
3. Refactor:
   - Переименовать методы `ToAgentConfig` → `GetAgentConfig`, `ToLLMConfig` → `GetLLMConfig`, `ToMCPServerConfig` → `GetMCPServerConfig`, `ToMCPConnectorConfig` → `GetMCPConnectorConfig`.
   - Удалить неиспользуемые типы `RuntimeConfig`, `ConfigAgent` и вложенные конфиги.
   - Обновить тесты: сравнение по значениям, без анонимных структур.
   - Обновить документацию в `documents/implementation.md` и `documents/file_structure.md`.
   - Обновить архитектурное описание в `documents/architecture.md` (если требуется).
4. Final Check:
   - Запустить `./run test` и `./run check`.
   - Исправить ошибки и зафиксировать изменения в Git.

# Выполнено
- Вся структура Configuration переведена на inline-структуры.
- Удалены устаревшие типы (RuntimeConfig, ConfigAgent и вложенные).
- Тесты и документация обновлены.
- Все тесты и проверки успешно пройдены (`./run test`, `./run check`).

# Следующий шаг
- Рефакторинг internal/configuration/manager.go и всех зависимых модулей для поддержки inline-структур (если потребуется дополнительная оптимизация или упрощение).
