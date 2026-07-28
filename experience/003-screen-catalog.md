---
Document Id: EXP-SCREENS
Title: Каталог экранов
Layer: L5 · Experience — Опыт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 5.4
Authority Rank: 40
Authority Scope: Структура информации, навигация, пользовательские сценарии.

Owners:
  - Product Owner
  - Design Lead

Depends On:
  - experience/001-information-architecture.md
  - experience/002-navigation-model.md
  - language/003-ui-terminology.md

Required By:
  - features/000-feature-catalog.md
  - design/006-component-language.md
  - experience/008-states-and-empty-states.md

Defines:
  - Реестр экранов и их идентификаторов
  - Назначение каждого экрана
  - Сущности и действия на экране

Must Not Define:
  - Вёрстку и визуальный дизайн (принадлежит design/)
  - Логику фич (принадлежит features/catalog/)
---

# Каталог экранов

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Единый реестр всех экранов продукта: идентификатор, роль, назначение, отображаемые сущности, доступные действия.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Design Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`experience/001-information-architecture.md`](../experience/001-information-architecture.md)
- [`experience/002-navigation-model.md`](../experience/002-navigation-model.md)
- [`language/003-ui-terminology.md`](../language/003-ui-terminology.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`design/006-component-language.md`](../design/006-component-language.md)
- [`experience/008-states-and-empty-states.md`](../experience/008-states-and-empty-states.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L5 · Experience — Опыт**
- Позиция в порядке чтения: **5.4**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый экран имеет id вида SCR-XXX.
2. Каждый экран отвечает хотя бы на один вопрос из product/000: что сейчас / что дальше / что получилось.
3. Каждое действие на экране соответствует команде из domain/commands/000-domain-command-catalog.md.
4. Каждый экран достижим по experience/002-navigation-model.md.

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
