---
Document Id: PRD-RULES
Title: Продуктовые правила
Layer: L2 · Product — Продукт
Status: Draft
Version: 0.2.0
Last Updated: 2026-08-01
Reading Order: 2.11
Authority Rank: 10
Authority Scope: Смысл, границы, цели и правила продукта. Высшая содержательная власть.

Owners:
  - Product Owner
  - Education Lead

Depends On:
  - product/000-product-overview.md
  - product/001-vision-and-principles.md
  - product/002-product-boundaries.md
  - governance/decisions/002-pending-decisions.md

Required By:
  - features/000-feature-catalog.md
  - experience/010-interaction-rules.md
  - design/000-design-overview.md
  - governance/002-review-checklist.md

Defines:
  - Инварианты продукта
  - Запреты продукта
  - Правила приёмки продуктового решения

Must Not Define:
  - Правила оформления документов (принадлежит meta/)
  - Правила доменного моделирования (принадлежит domain/architecture/)
---

# Продуктовые правила

> **Статус: DRAFT.**
> В этой редакции заполнены только правила вертикального среза B.0 — управляемое создание ученика, приглашение и активация. Остальные продуктовые правила остаются вне области редакции.

---

## 1. Purpose · Назначение

Свод инвариантов продукта: правила, которые обязаны соблюдаться в любой фиче, на любом экране, в любом релизе. Проверочный лист для приёмки любого решения.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Education Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/000-product-overview.md`](../product/000-product-overview.md)
- [`product/001-vision-and-principles.md`](../product/001-vision-and-principles.md)
- [`product/002-product-boundaries.md`](../product/002-product-boundaries.md)
- [`governance/decisions/002-pending-decisions.md`](../governance/decisions/002-pending-decisions.md), PD-0030

**Зависит от этого документа (обновить при изменении):**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`experience/010-interaction-rules.md`](../experience/010-interaction-rules.md)
- [`design/000-design-overview.md`](../design/000-design-overview.md)
- [`governance/002-review-checklist.md`](../governance/002-review-checklist.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L2 · Product — Продукт**
- Позиция в порядке чтения: **2.11**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждое правило сформулировано как проверяемое утверждение (да/нет).
2. Каждое правило имеет id вида PR-XX и прямую ссылку на авторитетный Product-документ либо записанное решение, которое действительно поддерживает всё правило.
3. Правила не дублируют друг друга и не противоречат друг другу.
4. Каждое правило можно нарушить — правило, которое невозможно нарушить, удаляется.

Дополнительно применяются общие правила: [`meta/006-validation-rules.md`](../meta/006-validation-rules.md)

---

## 6. Content · Содержание

### 6.1. Область редакции B.0

Belcanto Product начинает сопровождать человека после того, как он стал учеником школы. Настоящая редакция ограничена правилами управляемого создания Student, приглашения и активации; она не определяет полный набор будущих продуктовых правил.

**Источники:** [`product/000-product-overview.md`](000-product-overview.md), раздел «Что такое Belcanto Product»; [`product/002-product-boundaries.md`](002-product-boundaries.md), раздел «Начало ответственности продукта»; [`PD-0030`](../governance/decisions/002-pending-decisions.md), пп. 1–13.

### 6.2. Правила управляемого доступа B.0

#### PR-01 · Нет публичной регистрации

Продукт НЕ ДОЛЖЕН предоставлять публичную команду создания Student или `signup`. Запрос без действующего приглашения НЕ ДОЛЖЕН активировать новый Account.

**Источник:** [`PD-0030`](../governance/decisions/002-pending-decisions.md), пп. 1, 4.

#### PR-02 · Ученика учреждает школа

Student и его учебная история создаются уполномоченным представителем школы до активации Account. Ученик активирует доступ к существующему Student и НЕ ДОЛЖЕН создавать учебную идентичность самостоятельно.

**Источник:** [`product/000-product-overview.md`](000-product-overview.md), раздел «Что такое Belcanto Product»; [`product/002-product-boundaries.md`](002-product-boundaries.md), раздел «Начало ответственности продукта»; [`PD-0030`](../governance/decisions/002-pending-decisions.md), пп. 1–3.

#### PR-03 · Создание требует delegation, а Invitation остаётся Owner-only

Роль Administrator без явно выданного Owner полномочия «Суперадминистратор» НЕ ДОЛЖНА позволять создание Student. Выпуск, перевыпуск и отзыв Invitation ДОЛЖНЫ оставаться доступны только Owner — в том числе при действующем delegation grant.

**Источник:** [`PD-0030`](../governance/decisions/002-pending-decisions.md), пп. 7–10.

#### PR-04 · Суперадминистратор является делегированным уровнем доступа

Полномочие уровня «Суперадминистратор» обозначает Administrator с действующим delegation grant, выданным Owner, а не отдельную персону. Только Owner ДОЛЖЕН иметь возможность выдать или отозвать это полномочие; его B.0 scope ограничен созданием Student и чтением состояния onboarding.

**Источник:** [`product/003-personas.md`](003-personas.md), вводное правило «роль определяет не права доступа»; [`PD-0030`](../governance/decisions/002-pending-decisions.md), пп. 7–10.

#### PR-05 · Учебная история независима от цифрового доступа

Отзыв Invitation, истечение Invitation, блокировка Account или отсутствие Account НЕ ДОЛЖНЫ удалять Student либо его учебную историю.

**Источник:** [`product/000-product-overview.md`](000-product-overview.md), раздел «История важнее текущего состояния»; [`PD-0030`](../governance/decisions/002-pending-decisions.md), п. 3.

#### PR-06 · Пароль принадлежит ученику

Ученик ДОЛЖЕН самостоятельно задать пароль в ходе активации. Сотрудник школы НЕ ДОЛЖЕН задавать, читать или передавать пароль ученика.

**Источник:** [`PD-0030`](../governance/decisions/002-pending-decisions.md), п. 5.

#### PR-07 · Invitation ограничено и одноразово

Invitation ДОЛЖНО быть персональным, одноразовым, отзывным и ограниченным сроком действия. Перевыпуск ДОЛЖЕН сделать прежнее неиспользованное Invitation недействительным.

**Источник:** [`PD-0030`](../governance/decisions/002-pending-decisions.md), п. 6.

#### PR-08 · Сначала учебная ценность, затем доступ

Invitation НЕ ДОЛЖНО выпускаться до публикации первого учебного ориентира закреплённым Teacher. После активации и повторного входа Student ДОЛЖЕН видеть только собственный ориентир.

**Источник:** [`product/000-product-overview.md`](000-product-overview.md), раздел «Развитие важнее развлечения»; [`PD-0030`](../governance/decisions/002-pending-decisions.md), пп. 11, 13.

#### PR-09 · Возрастная граница первого среза

B.0 ДОЛЖЕН допускать активацию только совершеннолетнего Student. Guardian/детский доступ остаётся отдельным будущим срезом и настоящей редакцией не определяется.

**Источник:** [`PD-0030`](../governance/decisions/002-pending-decisions.md), п. 12.

### 6.3. Таблица вывода

| Раздел | Класс | Источник | Версия источника |
|--------|-------|----------|------------------|
| 5.2 | A | [`meta/002-authority-model.md`](../meta/002-authority-model.md), раздел 6.3, шаг 2, и раздел 6.5: записанное решение является истиной до обновления производного документа | 1.1.1 |
| 6.1 | A | [`product/000-product-overview.md`](000-product-overview.md), «Что такое Belcanto Product»; [`product/002-product-boundaries.md`](002-product-boundaries.md), «Начало ответственности продукта» | 1.1.0; 1.0.0 |
| 6.1, 6.2 (PR-01–PR-09) | B | [`PD-0030`](../governance/decisions/002-pending-decisions.md) | `GOV-PENDING-DECISIONS` 1.4.0 |
| 6.2 (PR-04) | A | [`product/003-personas.md`](003-personas.md), вводное правило и «Основные роли» | 1.0.0 |

---

## История изменений

| Версия | Дата | Изменение |
|--------|------|-----------|
| 0.2.0 | 2026-08-01 | DRAFT: по PD-0030 добавлены проверяемые правила B.0 для управляемого создания Student, делегируемого Owner уровня доступа «Суперадминистратор», приглашения и активации; правило 5.2 расширено на записанные решения по модели авторитета и явно трассировано; добавлена таблица вывода. Требуется человеческий P7. |
| 0.1.0 | 2026-07-28 | Создан скелет документа. |
