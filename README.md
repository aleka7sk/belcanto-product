# Belcanto Product — Source of Truth

Это репозиторий **продуктовой документации**, а не кода. Здесь нет приложения, интерфейса и реализации.

Здесь находится всё, что определяет продукт Belcanto: зачем он существует, для кого, из чего состоит, как устроен, как выглядит и в каком порядке создаётся.

> **Главное правило.** Если решение не описано здесь — его не существует.
> Если два документа противоречат друг другу — читай [`meta/002-authority-model.md`](meta/002-authority-model.md).

---

## С чего начать

| Кто вы | Что читать |
|--------|------------|
| Новый инженер | [`docs/project/project-constitution.md`](docs/project/project-constitution.md) → [`meta/000-documentation-system.md`](meta/000-documentation-system.md) → далее маршрут в [`meta/001-reading-order.md`](meta/001-reading-order.md), раздел 6.3 |
| Дизайнер | маршрут «Дизайнер» в [`meta/001-reading-order.md`](meta/001-reading-order.md) |
| Методист | маршрут «Преподаватель / методист» там же |
| Руководитель | маршрут «Руководитель» там же |
| **AI-агент** | [`CLAUDE.md`](CLAUDE.md) → [`meta/standard/008-agent-change-protocol.md`](meta/standard/008-agent-change-protocol.md) |

---

## 1. Структура репозитория

```
belcanto-product/
│
├── README.md                    ← вы здесь: карта репозитория
├── CLAUDE.md                    ← контракт для AI-агента
│
├── docs/
│   └── project/                 Project · Нормативный контекст проекта (вне L0–L11)
│       └── project-constitution.md
│
├── meta/                        L0 · Как устроена документация  (привязка)
│   ├── 000-documentation-system.md
│   ├── 001-reading-order.md
│   ├── 002-authority-model.md
│   ├── 003-document-template.md
│   ├── 004-anti-duplication-rules.md
│   ├── 005-anti-contradiction-rules.md
│   ├── 006-validation-rules.md
│   ├── 007-ownership-map.md
│   ├── 008-lifecycle-and-versioning.md
│   ├── 009-contribution-workflow.md
│   ├── 010-migration-register.md
│   ├── 011-conformance-statement.md
│   ├── 012-authoring-binding.md
│   │
│   └── standard/                L0 · Как документы производятся  (переиспользуемый стандарт)
│       ├── 000-standard-overview.md
│       ├── 001-document-lifecycle.md
│       ├── 002-entry-criteria.md
│       ├── 003-exit-criteria.md
│       ├── 004-validation-checklist.md
│       ├── 005-cross-reference-checklist.md
│       ├── 006-traceability.md
│       ├── 007-ownership-and-transfer.md
│       ├── 008-agent-change-protocol.md
│       ├── 009-conflict-detection.md
│       ├── 010-responsibility-budget.md
│       ├── 011-decomposition-rules.md
│       ├── 012-diagram-rules.md
│       ├── 013-example-rules.md
│       ├── 014-terminology-rules.md
│       ├── 015-migration-rules.md
│       ├── 016-conformance-and-binding.md
│       ├── 017-authoring-pipeline.md     ← конвейер производства
│       └── 018-document-class-profiles.md
│
├── school/                      L1 · Как школа работает СЕГОДНЯ
│   ├── 000-school-overview.md
│   ├── 001-education-model.md
│   ├── 002-current-student-journey.md
│   ├── 003-teacher-workflow.md
│   ├── 004-current-admin-workflow.md
│   ├── 005-current-owner-workflow.md
│   ├── 006-current-tooling-landscape.md
│   └── 007-open-questions.md
│
├── product/                     L2 · Зачем существует продукт
│   ├── 000-product-overview.md          ← главный документ продукта
│   ├── 001-vision-and-principles.md
│   ├── 002-product-boundaries.md
│   ├── 003-personas.md
│   ├── 004-core-concepts.md
│   ├── 005-domain-glossary.md           ← указатель на language/
│   ├── 006-business-goals.md
│   ├── 007-success-metrics.md
│   ├── 008-value-propositions.md
│   ├── 009-product-rules.md
│   └── 010-risks-and-assumptions.md
│
├── language/                    L3 · Что означает каждое слово
│   ├── 000-language-overview.md
│   ├── 001-ubiquitous-language.md       ← единственный каталог терминов
│   ├── 002-naming-rules.md
│   ├── 003-ui-terminology.md
│   ├── 004-forbidden-terms.md
│   └── 005-localization-policy.md
│
├── domain/                      L4 · Как устроена предметная область
│   ├── 000-domain-overview.md
│   ├── 001-domain-reading-order.md
│   ├── architecture/            правила моделирования
│   ├── aggregates/              агрегаты
│   ├── entities/                сущности
│   ├── value-objects/           объекты-значения
│   ├── commands/                команды
│   ├── events/                  доменные события
│   ├── services/                доменные сервисы
│   └── policies/                правила расчёта и реакции
│
├── experience/                  L5 · Как продукт устроен для пользователя
│   ├── 000-experience-overview.md
│   ├── 001-information-architecture.md
│   ├── 002-navigation-model.md
│   ├── 003-screen-catalog.md            ← реестр экранов SCR-XXX
│   ├── 004-journey-student.md
│   ├── 005-journey-teacher.md
│   ├── 006-journey-admin.md
│   ├── 007-journey-owner.md
│   ├── 008-states-and-empty-states.md
│   ├── 009-notification-experience.md
│   └── 010-interaction-rules.md
│
├── features/                    L6 · Что продукт умеет
│   ├── 000-feature-catalog.md           ← реестр фич F-XXX
│   ├── 001-feature-template.md
│   ├── 002-feature-lifecycle.md
│   ├── 003-feature-dependency-map.md
│   └── catalog/                 F-XXX-slug.md — по одному файлу на фичу
│
├── design/                      L7 · Как продукт выглядит и звучит
│   ├── 000-design-overview.md
│   ├── 001-design-principles.md
│   ├── 002-color-system.md
│   ├── 003-typography.md
│   ├── 004-spacing-and-layout.md
│   ├── 005-iconography.md
│   ├── 006-component-language.md
│   ├── 007-motion-language.md
│   ├── 008-sound-and-haptics.md
│   ├── 009-accessibility.md
│   ├── 010-content-and-voice.md
│   └── 011-design-tokens.md             ← единственное место конкретных значений
│
├── roadmap/                     L8 · В каком порядке делаем
│   ├── 000-roadmap-overview.md
│   ├── 001-mvp-scope.md
│   ├── 002-release-plan.md
│   ├── 003-milestones.md
│   └── 004-backlog-and-out-of-scope.md
│
├── governance/                  L9 · Кто и как принимает решения
│   ├── 000-governance-overview.md
│   ├── 001-change-management.md
│   ├── 002-review-checklist.md
│   └── decisions/
│       ├── 000-decision-log.md          ← журнал решений PDR-XXXX
│       ├── 001-decision-template.md
│       └── records/             PDR-XXXX-slug.md — по одному файлу на решение
│
├── ai/                          L10 · Как AI-агент работает с репозиторием
│   ├── 000-ai-overview.md
│   ├── 001-agent-reading-protocol.md
│   ├── 002-authority-resolution.md
│   ├── 003-writing-rules.md
│   ├── 004-task-playbooks.md
│   ├── 005-prohibited-actions.md
│   └── 006-ai-in-product-boundary.md    ← AI ВНУТРИ продукта (не про агентов)
│
└── assets/                      L11 · Материалы (ненормативные)
    ├── 000-assets-overview.md
    ├── images/
    ├── diagrams/
    ├── exports/
    └── references/
        ├── 000-references.md
        └── external/
```

---

## 2. Слои и их вопросы

Каждый слой отвечает на **один** вопрос. Документ не может отвечать на вопрос чужого слоя.

| Область / слой | Каталог | Вопрос | Характер | Ранг |
|----------------|---------|--------|----------|------|
| Project | `docs/project/` | Ради чего ведётся проект и в каком архитектурном порядке движется работа? | нормативный контекст проекта, вне L0–L11 | **4** |
| L0 Standard | `meta/standard/` | Как документы производятся? | нормативный (процесс) | **3** |
| L0 Meta | `meta/` | Как устроена документация? | нормативный (форма) | 5 |
| L1 Reality | `school/` | Как школа работает сегодня? | **описательный** | 90 |
| L2 Product | `product/` | Зачем существует продукт? | нормативный (смысл) | 10 |
| L3 Language | `language/` | Что означает каждое слово? | нормативный (термины) | 20 |
| L4 Domain | `domain/` | Как устроена предметная область? | нормативный (модель) | 30 |
| L5 Experience | `experience/` | Как продукт устроен для пользователя? | нормативный (структура) | 40 |
| L6 Features | `features/` | Что продукт умеет? | нормативный (поведение) | 50 |
| L7 Design | `design/` | Как выглядит и звучит? | нормативный (форма) | 60 |
| L8 Delivery | `roadmap/` | В каком порядке делаем? | нормативный (план) | 70 |
| L9 Governance | `governance/` | Кто и как решает? | процесс | **0** |
| L10 AI | `ai/` | Как работает агент? | производный | 100 |
| L11 Assets | `assets/` | Где материалы? | ненормативный | 110 |

`Project` — **область авторитета**, а не слой документации: она вне L0–L11 и слоя L12 не образует. Владельцы: [`meta/000`](meta/000-documentation-system.md) §6.2.2 — положение и граф вывода; [`meta/002`](meta/002-authority-model.md) §6.1 — ранг и зона; основания — PD-0027 и PD-0028.

**Направление знания:**

```
school → product → language → domain → experience → features → design → roadmap
факты     смысл      слова      модель    структура    поведение   форма    порядок
```

Документ ссылается на свой или предыдущий слой. Ссылка вперёд допустима только в поле `Required By`.

---

## 3. Порядок чтения

Канонический порядок — [`meta/001-reading-order.md`](meta/001-reading-order.md).

Короткая версия для нового инженера:

```
 1. docs/project/project-constitution.md  ради чего и в каком порядке ведётся проект
 2. meta/000-documentation-system.md      как устроена документация
 3. meta/002-authority-model.md           кто главнее при конфликте
 4. school/000, school/001                как школа работает сегодня
 5. product/000 → 001 → 002 → 003         зачем продукт, границы, роли
 6. language/001                          язык продукта
 7. domain/000 → domain/001               предметная область
 8. experience/001 → 002 → 003            структура, навигация, экраны
 9. features/000                          что продукт умеет
10. design/000, design/009                форма и доступность
11. roadmap/001                           первая версия
```

**Правило порядка.** Документ нельзя читать раньше любого документа из его поля `Depends On`.

**Для задачи** читать не всё, а по формуле:

```
meta/002 + документ-владелец понятия + его Depends On + его Required By
```

Владельца понятия искать в [`meta/007-ownership-map.md`](meta/007-ownership-map.md).

---

## 4. Форма документа

Каждый документ имеет обязательный frontmatter и разделы 1–6. Полностью — [`meta/003-document-template.md`](meta/003-document-template.md).

Ключевые поля и их назначение:

| Поле | Зачем |
|------|-------|
| `Defines` | понятия, которыми документ **владеет** — механизм против дублирования |
| `Must Not Define` | чужие зоны — защита от расползания документа |
| `Depends On` | что прочитать **до** |
| `Required By` | что обновить **после** — механизм каскада |
| `Authority Rank` / `Authority Scope` | разрешение конфликтов |
| `Status` | можно ли опираться на документ сегодня |

Разделы:

```
1. Purpose · Назначение          зачем документ существует
2. Owner · Владелец              кто отвечает
3. Dependencies · Зависимости    читать до / обновить после
4. Reading Order                 место в порядке чтения
5. Validation Rules              проверяемые условия «да / нет»
6. Content · Содержание          ← единственное место знания
```

Разделы 1–5 — контракт, действуют с момента создания файла. Раздел 6 — само знание.

---

## 5. Правила против противоречий

Полностью — [`meta/005-anti-contradiction-rules.md`](meta/005-anti-contradiction-rules.md).

**Меры, встроенные в систему:**

| # | Мера |
|---|------|
| M1 | Каждое понятие определяется **ровно в одном** документе |
| M2 | Знание течёт в одну сторону; нижний слой не переопределяет верхний |
| M3 | Изменение обязано пройти каскад `Required By` |
| M4 | `Defines` / `Must Not Define` очерчивают зону документа |
| M5 | `school/` описывает настоящее, остальные слои — целевое; расхождение между ними — замысел, а не ошибка |
| M6 | Конкретные значения существуют только в `design/011-design-tokens.md` |
| M7 | Изменение смысла документа ранга ≤ 30 требует PDR |

**Алгоритм при обнаружении противоречия** ([`meta/002`](meta/002-authority-model.md), раздел 6.3). Предмет спора относится к слою **или к области авторитета**:

```
1. Есть принятый PDR по предмету спора?        → PDR есть истина
2. Предмет в зоне ровно одного документа?      → истина в нём
3. Один документ объявляет предмет в Defines?  → истина в нём
4. Ранги различаются?                          → истина у меньшего ранга
5. Иначе → ОСТАНОВИТЬСЯ: Status: Disputed на обоих + открыть PDR
```

**Запрещено** молчаливо выбирать «более правильный» вариант. Противоречие либо разрешается алгоритмом, либо эскалируется записью решения.

---

## 6. Правила против дублирования

Полностью — [`meta/004-anti-duplication-rules.md`](meta/004-anti-duplication-rules.md).

| # | Правило |
|---|---------|
| **D1** | Каждое понятие определяется в одном документе — **владельце смысла**, объявляющем его в `Defines` |
| **D2** | Владелец выбирается по вопросу: *что значит слово* → `language/`; *как устроено* → `domain/`; *где видит пользователь* → `experience/`; *как работает* → `features/`; *как выглядит* → `design/`; *когда сделаем* → `roadmap/` |
| **D3** | Дубликат — второй ответ на **тот же** вопрос. Разные аспекты одного понятия дубликатом не являются |
| **D4** | Разрешённые формы повторения: **ссылка**, **цитата ≤ 2 предложений с атрибуцией**, **документ-указатель** |
| **D5** | Вместо копирования данных — **таблица связей** только из идентификаторов |
| **D6** | Совпадающие копии → удалить лишнюю. Расходящиеся копии → это противоречие, см. раздел 5 |
| **D7** | Новый файл создаётся **последним**: сначала искать существующий документ. Документ без `Defines` не имеет права на существование |

**Реестр «понятие → владелец»:** [`meta/007-ownership-map.md`](meta/007-ownership-map.md). Проверять **перед** тем, как что-то писать.

---

## 6.5. Стандарт производства документации

Правила выше отвечают на вопрос «как устроен корпус». Каталог [`meta/standard/`](meta/standard/000-standard-overview.md) отвечает на вопрос «как в него добавляют» — это переиспользуемый, продуктово-нейтральный стандарт производства документации.

| Область | Документ |
|---------|----------|
| Жизненный цикл: `Scaffold → Draft → Review → Approved → Canonical → Superseded` | [DES-001](meta/standard/001-document-lifecycle.md) |
| Условия создания документа | [DES-002](meta/standard/002-entry-criteria.md) |
| Условия завершённости | [DES-003](meta/standard/003-exit-criteria.md) |
| Чек-лист валидации | [DES-004](meta/standard/004-validation-checklist.md) |
| Чек-лист перекрёстных ссылок | [DES-005](meta/standard/005-cross-reference-checklist.md) |
| Прослеживаемость утверждений | [DES-006](meta/standard/006-traceability.md) |
| Владение и передача | [DES-007](meta/standard/007-ownership-and-transfer.md) |
| Протокол AI-агента | [DES-008](meta/standard/008-agent-change-protocol.md) |
| Обнаружение конфликтов | [DES-009](meta/standard/009-conflict-detection.md) |
| Бюджет ответственности | [DES-010](meta/standard/010-responsibility-budget.md) |
| Декомпозиция и слияние | [DES-011](meta/standard/011-decomposition-rules.md) |
| Диаграммы · Примеры · Термины | [DES-012](meta/standard/012-diagram-rules.md) · [DES-013](meta/standard/013-example-rules.md) · [DES-014](meta/standard/014-terminology-rules.md) |
| Миграции | [DES-015](meta/standard/015-migration-rules.md) |
| Соответствие и привязка | [DES-016](meta/standard/016-conformance-and-binding.md) |
| **Конвейер производства** | [DES-017](meta/standard/017-authoring-pipeline.md) |
| **Профили классов документов** | [DES-018](meta/standard/018-document-class-profiles.md) |

### Конвейер производства

Продуктовый документ **не пишут напрямую**. Он производится десятью стадиями; файл появляется только на четвёртой.

```
P0 Запрос → P1 Область → P2 Источники → P3 Утверждения → P4 Структура → P5 Черновик
                                                                            ↓
P9 Канонизация ← P8 Интеграция ← P7 Рецензия ← P6 Самопроверка ←────────────┘
   Canonical       Approved        Review          Draft
```

Ключевая стадия — **P3**: перечень всего, что документ намерен утверждать, с источником по каждому пункту, утверждается **до написания прозы**. Несогласие с рамкой обнаруживается на 30 строках реестра, а не после недели работы над текстом.

| Свойство | Чем достигается |
|----------|-----------------|
| Проза пишется один раз | ворота P3; правка идёт через утверждение, а не через текст |
| Рамка проверяется рано | жёсткие ворота P1, P3 |
| Источник не «уплывёт» | закрепление версий на P2, обнаружение на P8 |
| Зона не расползётся | изменение `MINOR` входит на P3, а не на P5 |
| Нет колебаний между стадиями | правило двух возвратов |
| Отказ адресуется точно | возврат на стадию, где дефект **возник** |
| **AI-агент пишет прозу безопасно** | к P5 всё суждение израсходовано на P1–P3 |

Пять из шести стадий, где агент наиболее полезен, — P2, P3, P6, P8 — это сбор, классификация, проверка и каскад. Вердикт (P7) и канонизация (P9) остаются за человеком.

Один конвейер работает для всех видов документов; различается **профиль риска** ([DES-018](meta/standard/018-document-class-profiles.md)) — где именно данный класс ломается:

| Класс | Пример | Ищи в первую очередь |
|-------|--------|----------------------|
| C1 Замысел | `product/000…002` | утверждение, не запрещающее ничего |
| C2 Наблюдение | `school/` | реконструкцию, выданную за подтверждённое |
| C3 Определение | `language/001` | понятие, уже принадлежащее другому документу |
| C4 Модель | `domain/` | определение слова там, где нужна структура |
| C5 Опыт | `experience/` | шаг без источника; разрыв в пути |
| C6 Поведение | `features/catalog/` | правило, принадлежащее модели |
| C7 Форма | `design/` | величину вне `design/011` |
| C8 План | `roadmap/` | единицу вне реестра |
| C9 Реестр | `features/000` | описание вместо идентификатора |

Соответствие каталогов классам, роли на стадиях и место хранения артефактов — [`meta/012-authoring-binding.md`](meta/012-authoring-binding.md).

**Для 68 скелетов стадии P0–P4 уже пройдены.** Их разделы 1–5 — это карта области и структура. Наполнение каждого начинается со стадии **P2**, а не с нуля.

**Стандарт и привязка.** Стандарт задаёт требования в абстрактных терминах («вышестоящий документ», «порог», «владеющая роль»). Значения даёт **привязка** — документы `meta/` вне `standard/`. Стандарт переносится в другой репозиторий целиком; привязка создаётся заново.

**Стандарт читается перед первым авторством, а не перед первым чтением.** Тому, кто только читает документацию, он не нужен.

**Текущее соответствие: уровень не заявлен.** Не автоматизированы проверки, не наполнены реестры терминов и решений, конвейер задан, но ещё ни разу не пройден целиком. Полный разбор с перечнем непокрытого и четырьмя зарегистрированными отклонениями — [`meta/011-conformance-statement.md`](meta/011-conformance-statement.md).

---

## 7. Как внести изменение

Полностью — [`meta/009-contribution-workflow.md`](meta/009-contribution-workflow.md).

```
1. Определить понятие → найти владельца (meta/007)
2. Прочитать владельца + его Depends On
3. Проверить, нет ли уже такого знания
4. Изменить ОДИН документ, только раздел 6
5. Повысить Version, обновить Last Updated и историю
6. Пройти каскад Required By          ← чаще всего пропускают
7. Обновить реестры (meta/001, meta/007)
8. Проверить по meta/006 и governance/002
9. Изменён смысл документа ранга ≤ 30 → создать PDR
```

Изменение не завершено, пока не выполнены все девять шагов.

---

## 8. Текущее состояние

Значения состояний — [`meta/standard/001-document-lifecycle.md`](meta/standard/001-document-lifecycle.md). Прежние `Accepted` и `Deprecated` выведены из употребления миграцией M-0001.

| Область / слой | Состояние |
|----------------|-----------|
| `Project · docs/project/` | `Approved` — Конституция проекта 1.1.1 |
| `meta/standard/` | `Approved` — стандарт производства описан |
| `meta/` | `Approved` — система документации описана |
| `school/` | частично `Approved` (M-0001, партия 2); `006`, `007` — `Scaffold` |
| `product/` | `000`–`003` `Approved` (M-0001, партия 2); `004`–`010` — `Scaffold` |
| `language/` | `Scaffold` |
| `domain/` | `Approved` (M-0001, партия 3), наполнен |
| `experience/` | `Scaffold` |
| `features/` | `Scaffold` |
| `design/` | `Scaffold` |
| `roadmap/` | `Scaffold` |
| `governance/` | `Scaffold` |
| `ai/` | `Scaffold` |
| `assets/` | `Scaffold` |

`Scaffold` означает: разделы 1–5 заполнены и уже действуют, раздел 6 не написан.

### Незавершённая миграция формы

34 документа, созданные до введения системы (`school/`, `product/000`–`003`, весь `domain/`), имеют старый frontmatter. Им недостаёт полей `Defines`, `Must Not Define`, `Authority Rank`, `Authority Scope`, `Reading Order`, `Required By` — то есть именно тех, на которых держатся защита от дублирования и каскад обновлений.

До завершения миграции правила [`meta/004`](meta/004-anti-duplication-rules.md) и [`meta/005`](meta/005-anti-contradiction-rules.md) применяются к этим документам вручную.

Порядок миграции: `product/000`–`003` → `school/` → `domain/`. Каждый мигрированный документ вносится в [`meta/007-ownership-map.md`](meta/007-ownership-map.md) в том же изменении.

**Открытые противоречия**, требующие решения до наполнения зависимых документов — [`meta/005-anti-contradiction-rules.md`](meta/005-anti-contradiction-rules.md), раздел 6.6.
