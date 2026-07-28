---
Document Id: EXP-STATES
Title: Состояния и пустые состояния
Layer: L5 · Experience — Опыт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 5.9
Authority Rank: 40
Authority Scope: Структура информации, навигация, пользовательские сценарии.

Owners:
  - Design Lead
  - Product Owner

Depends On:
  - experience/003-screen-catalog.md
  - design/010-content-and-voice.md

Required By:
  - features/001-feature-template.md
  - design/006-component-language.md

Defines:
  - Перечень типов состояний
  - Требования к каждому состоянию

Must Not Define:
  - Тексты конкретных экранов
  - Визуальные токены (принадлежит design/011)
---

# Состояния и пустые состояния

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить, что видит пользователь, когда данных нет, данные загружаются, произошла ошибка или доступ ограничен. Пустое состояние — часть продукта, а не исключение.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Design Lead
- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`experience/003-screen-catalog.md`](../experience/003-screen-catalog.md)
- [`design/010-content-and-voice.md`](../design/010-content-and-voice.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/001-feature-template.md`](../features/001-feature-template.md)
- [`design/006-component-language.md`](../design/006-component-language.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L5 · Experience — Опыт**
- Позиция в порядке чтения: **5.9**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Для каждого экрана из experience/003 определены все применимые состояния.
2. Каждое пустое состояние предлагает следующее действие.
3. Ни одно состояние не обвиняет пользователя (запрет product/000).

Дополнительно применяются общие правила: [`meta/006-validation-rules.md`](../meta/006-validation-rules.md)

---

## 6. Content · Содержание

<!-- SCAFFOLD: содержание не написано.
     Перед заполнением прочитать документы из раздела 3.
     После заполнения: Status → Draft, обновить meta/007-ownership-map.md. -->

_Содержание не написано._

---

## История изменений

| Версия | Дата | Изменение |
|--------|------|-----------|
| 0.1.0 | 2026-07-28 | Создан скелет документа. |
