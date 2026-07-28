---
Document Id: EXP-JRN-ADMIN
Title: Целевой сценарий: путь администратора
Layer: L5 · Experience — Опыт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 5.7
Authority Rank: 40
Authority Scope: Структура информации, навигация, пользовательские сценарии.

Owners:
  - Product Owner
  - Education Lead

Depends On:
  - product/003-personas.md
  - school/004-current-admin-workflow.md
  - experience/003-screen-catalog.md

Required By:
  - features/000-feature-catalog.md
  - roadmap/001-mvp-scope.md

Defines:
  - Целевой путь администратора
  - Шаги, экраны и ценность каждого шага

Must Not Define:
  - Текущий путь в школе (принадлежит school/)
  - Поведение отдельных фич (принадлежит features/catalog/)
---

# Целевой сценарий: путь администратора

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Описать целевой путь администратора в продукте — от первого входа до регулярного использования, с указанием экранов, событий и ценности на каждом шаге.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Education Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/003-personas.md`](../product/003-personas.md)
- [`school/004-current-admin-workflow.md`](../school/004-current-admin-workflow.md)
- [`experience/003-screen-catalog.md`](../experience/003-screen-catalog.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`roadmap/001-mvp-scope.md`](../roadmap/001-mvp-scope.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L5 · Experience — Опыт**
- Позиция в порядке чтения: **5.7**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый шаг ссылается на экран SCR-XXX из experience/003-screen-catalog.md.
2. Каждый шаг объясняет ценность для пользователя.
3. Документ описывает целевое состояние; текущее состояние остаётся в school/.
4. Каждый шаг связан хотя бы с одним доменным событием.

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
