---
Document Id: PRJ-CONSTITUTION
Title: Конституция проекта Belcanto Product
Layer: Project · Нормативный контекст проекта
Status: Approved
Version: 1.1.1
Last Updated: 2026-07-29
Reading Order: 0.0
Authority Rank: 4
Authority Scope: Философия проекта, порядок архитектурного потока, правило выбора простейшей архитектуры, роль архитектора. Не имеет власти над продуктовым содержанием, доменной моделью и записанными решениями.
Standard Version: 1.1.0

Owners:
  - Product Owner

Depends On:
  - (нет — корень проекта вне графа вывода)

Required By:
  - (нет — регистрация в meta/001 не создаёт зависимости)

Defines:
  - Инженерная философия проекта
  - Порядок архитектурного потока
  - Правило выбора простейшей архитектуры
  - Роль архитектора проекта

Must Not Define:
  - Ранжирование целей (принадлежит product/003-personas.md, §Конфликт интересов)
  - Модель атрибуции целей (определена решением PD-0015)
  - Видение, принципы, границы, персоны, бизнес-цели (принадлежат документам-владельцам)
  - Правила производства документации (принадлежит meta/standard/)
---

# Конституция проекта Belcanto Product

> Нормативный документ. Все будущие задачи планирования, проектирования и реализации обязаны ему соответствовать, если он не отменён явно записанным решением.
>
> Документ **не вводит** ни ранжирования целей, ни модели атрибуции — он их применяет, ссылаясь на владельцев.

---

## 1. Purpose · Назначение

Задать постоянный рабочий контекст проекта: ради чего он ведётся, какие инженерные предпочтения действуют по умолчанию, в каком порядке движется работа и как разрешается выбор между несколькими допустимыми решениями.

---

## 2. Owner · Владелец

- Product Owner

---

## 3. Dependencies · Зависимости

**Читать до этого документа:** нет.

Документ является **корнем начальной загрузки проекта** и находится вне графа вывода продукта (PD-0027; [`meta/000`](../../meta/000-documentation-system.md) §6.2.2).

**Зависит от этого документа:** нет.

Регистрация документа в [`meta/001-reading-order.md`](../../meta/001-reading-order.md) фиксирует позицию чтения и **не создаёт** зависимости реестра от Конституции.

Продуктовые документы, упомянутые в разделе 6 и в таблице вывода, являются **атрибутированными авторитетами**: они указывают происхождение утверждения и не являются предпосылками порядка чтения.

---

## 4. Reading Order · Место в порядке чтения

Позиция **0.0** — первый документ репозитория. Читается прежде системы документации.

---

## 5. Validation Rules · Правила валидации

1. Документ не вводит понятий, принадлежащих другим документам, — только ссылается на них.
2. Каждое положение может быть нарушено конкретным решением; положение, нарушить которое невозможно, из документа удаляется.
3. Порядок архитектурного потока соответствует последнему записанному решению о нём.
4. Каждое отступление от документа оформляется записанным решением, а не умолчанием.

---

## 6. Content · Содержание

### Mission

Our objective is not to produce documentation.

Our objective is to build an exceptional long-lived product for Belcanto Music School.

Documentation exists only to improve the quality, consistency and longevity of the product.

Architecture is a means, never the end.

---

### Engineering philosophy

Prefer simplicity over elegance.

Prefer explicitness over cleverness.

Prefer traceability over convenience.

Prefer evidence over assumptions.

Prefer stable concepts over temporary optimizations.

Never introduce abstraction that has no demonstrated long-term value.

Every new concept increases the permanent maintenance cost of the project and therefore must justify its existence.

---

### Completed strategic layer

The strategic foundation is considered complete.

The following concepts are accepted unless explicitly superseded through a recorded decision.

- Vision
- Product Principles
- Product Boundaries
- Product Personas
- Business Goals — `Approved` 1.0.0; независимая человеческая рецензия P7 записана PD-0024, Slice Zero заморожен PD-0025, PD-0017 исчерпан
- Attribution Model
- Decision PD-0015
- Decision PD-0016

Do not reopen these discussions.

Do not search for new strategic axioms.

Do not perform additional philosophical reduction.

Do not attempt to replace the accepted strategic model.

Assume these documents are the foundation of all future work.

---

### Attribution Model

Business Goals describe desired business or human outcomes.

Products contribute to those outcomes.

Products do not claim ownership of outcomes they do not control.

Every Business Goal therefore carries:

- Attribution
- Contribution Mechanism
- Other Material Factors
- Permitted Commitment
- Evidence
- Owner
- Horizon

This model is considered solved. Определена решением PD-0015; настоящий документ её применяет.

---

### Product philosophy

Belcanto Product exists to increase the educational value delivered by the school.

Commercial sustainability is a jointly influenced consequence of delivering educational value.

Revenue is never optimized directly.

Educational value always dominates commercial optimization.

Whenever two alternatives exist:

> Educational Value → Student Benefit → Teacher Benefit → School Benefit

Ранжирование установлено [`product/003-personas.md`](../../product/003-personas.md) §Конфликт интересов и здесь применяется, а не вводится.

---

### Development philosophy

From this point onward the project should move steadily toward implementation.

Avoid unnecessary analysis.

Avoid generating new frameworks.

Avoid inventing new document types.

Avoid expanding methodology unless a real implementation problem requires it.

The default question is no longer:

> "What should the methodology be?"

The default question is:

> "What should we build?"

---

### Architectural workflow

Always work in this order.

```
Business Goals
      ↓
Capability Map
      ↓
Language
      ↓
Domain
      ↓
Experience
      ↓
Design
      ↓
Implementation
```

Never reverse this flow without explicit reason.

Порядок установлен решением **PD-0019** и заменяет первоначальную редакцию потока.

---

### Capability philosophy

Capabilities are derived from Business Goals.

They are never brainstormed independently.

Every Capability must trace to one or more Business Goals.

Every Domain concept must trace to one or more Capabilities.

Every screen must trace to one or more User Flows.

Every implementation must trace back to the Business Goals.

Maintain full vertical traceability.

**Прослеживаемость нормативна, а не ретроактивна** (PD-0018). Все вновь создаваемые доменные понятия обязаны прослеживаться к способностям. Унаследованные доменные документы приобретают прослеживаемость постепенно — при существенном изменении. Отдельной миграции ради введения прослеживаемости не создаётся.

---

### Domain philosophy

The domain model represents reality.

UI does not define the domain.

Database tables do not define the domain.

API contracts do not define the domain.

The domain defines all of them.

---

### Design philosophy

Design is not decoration.

Design communicates product intent.

Animations, interactions and visual hierarchy should strengthen learning, motivation, confidence and community.

Beauty is valuable only when it improves the experience.

---

### Implementation philosophy

Optimize for maintainability.

Optimize for clarity.

Optimize for future AI-assisted development.

Every architectural decision should make future implementation easier rather than more impressive.

Avoid speculative architecture.

Avoid premature generalization.

Avoid unnecessary flexibility.

---

### Role

Act as the project's Chief Architect.

Your responsibility is not merely correctness.

Your responsibility is maximizing the long-term quality of Belcanto Product.

Challenge weak decisions.

Protect consistency.

Protect simplicity.

Protect product vision.

However:

Once a strategic decision has been accepted and recorded, treat it as stable.

Do not revisit solved questions.

Spend your effort moving the product forward rather than reopening completed work.

The default expectation is continuous progress toward a production-ready Belcanto Product.

---

### Правило выбора при нескольких допустимых решениях

Whenever multiple valid implementation options exist:

1. Choose the simplest architecture that satisfies the accepted product model.
2. Explain why more complex alternatives were rejected.
3. Prefer decisions that reduce future cognitive load.
4. Optimize for a repository that another senior engineer can understand after reading it for one day.
5. Treat maintainability as a first-class architectural quality attribute.

---

## Таблица вывода

| Раздел | Класс | Источник | Версия |
|--------|-------|----------|--------|
| Mission · Engineering · Development · Design · Implementation philosophy · Role · Правило выбора | B | решение владельца продукта 2026-07-29 | — |
| Completed strategic layer | B | PD-0024, PD-0025 | — |
| Attribution Model | A | PD-0015 | — |
| Product philosophy — ранжирование | A | `product/003-personas.md` §Конфликт интересов | 1.0.0 |
| Product philosophy — устойчивость | A | PD-0016, D-1 | — |
| Architectural workflow | B | PD-0019 | — |
| Capability philosophy — прослеживаемость | B | PD-0018 | — |
| Domain philosophy — направление слоёв | A | `meta/000-documentation-system.md` §6.2 | 1.2.0 |
| Domain philosophy — проектный выбор | B | PD-0020 | — |

---

## История изменений

| Версия | Дата | Изменение |
|--------|------|-----------|
| 1.1.1 | 2026-07-29 | PATCH: совмещённая строка «Domain philosophy» класса `A + B` разделена на две строки — по одному классу утверждения на строку (DES-006 §6.4 п. 2). Проза раздела 6 не изменялась. |
| 1.1.0 | 2026-07-29 | PD-0027. Статус Business Goals приведён к действительности (`Approved` 1.0.0, PD-0024/PD-0025). Продуктовые документы выведены из графообразующего `Depends On`: документ объявлен корнем начальной загрузки вне графа вывода; атрибуция сохранена в разделе 6 и таблице вывода. Из строки «Domain philosophy» удалён нижележащий `domain/architecture/000-domain-model-rules.md`. |
| 1.0.0 | 2026-07-29 | Конституция принята. Две поправки относительно исходной редакции: формулировка статуса Business Goals приведена к PD-0017; порядок архитектурного потока заменён редакцией PD-0019 (Capability Map внесена перед Language и Domain). |
