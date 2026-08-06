# Belcanto Product — полный production implementation prompt для Claude Code

## Конфигурация запуска

- Model: **Fable 5**
- Effort: **Max**
- Mode: **Default**
- Session: **новая implementation-сессия**; старую сессию можно возобновлять только после повторной проверки repository preflight ниже
- Repository: `https://github.com/aleka7sk/belcanto-product`
- Verified baseline snapshot (2026-08-06): branch `agent/l1-internal-scheduling`, commit `d447c2c3cc45e507cf2c80eba054c32053a9904d`
- Implementation branch: `agent/production-v2`
- Working directory: фактический корень текущего checkout этого репозитория, полученный через `git rev-parse --show-toplevel`; не используй абсолютный `/workspace/scratch/...` из другой среды
- Figma file: `Belcanto Mobile — Premium Product Design v1.0`
- Figma file key: `yXE7a9vAyWdbU9iLnjFmXf`
- Canonical handoff: https://www.figma.com/design/yXE7a9vAyWdbU9iLnjFmXf?node-id=293-20

---

Ты работаешь над **полной production-версией Belcanto**, а не над MVP, прототипом или набором статических экранов.

Твоя задача — самостоятельно реализовать в текущем репозитории весь продуктовый контур, заданный Figma Pages `19–37`. Экранный implementation scope — ровно 242 канонических состояния Pages `21–34`; Pages `19–20` и `37` задают contract/handoff, Page `35` — shared components, Page `36` — только flow semantics. Реализация включает backend, PostgreSQL, OpenAPI, mobile data layer, permissions, production UI, offline/recovery, observability и тесты. После анализа не останавливайся на плане: последовательно выполни все незаблокированные фазы до готового результата.

## 1. Непосредственно перед изменениями

1. Выполни обязательный repository preflight **до любых изменений**:

   ```bash
   git rev-parse --show-toplevel
   git remote get-url origin
   git status --short --branch
   git branch --show-current
   git rev-parse HEAD
   git log --oneline --decorate -20
   git diff --stat
   git diff --cached --stat
   git fetch --prune origin
   git rev-parse origin/main
   git rev-parse origin/agent/l1-internal-scheduling
   git rev-list --left-right --count HEAD...origin/agent/l1-internal-scheduling
   ```

   Проверенный remote snapshot на `2026-08-06`:

   - `origin/main` = `3086ca9a8c5e34408251311c8eae7357619b792f`;
   - `origin/agent/l1-internal-scheduling` = `d447c2c3cc45e507cf2c80eba054c32053a9904d`.

   Это reference snapshot, а не разрешение вслепую откатиться к старому SHA. Если fetched remote HEAD изменился, сначала прочитай новые commits/diff и отрази влияние в gap matrix. Если существуют незапушенные commits, незакоммиченные или untracked пользовательские файлы, не переключай ветку, не stash'ь, не очищай и не перезаписывай их. До записи кода определи, пересекаются ли они с production scope; при пересечении или неясности остановись и сообщи точный конфликт.

   Если более новой релевантной работы нет и worktree чистый, используй именно `origin/agent/l1-internal-scheduling`, а не `main`, как baseline. Создай `agent/production-v2`, если такой ветки ещё нет; если она уже существует локально или на origin, сначала проверь её историю и продолжай с последнего подтверждённого checkpoint — не создавай параллельную реализацию и не перезаписывай существующую ветку.

   В preflight-отчёте зафиксируй resolved repository path, remote URL, текущую ветку, baseline SHA, worktree status и divergence counts.

   Все команды preflight обязаны завершиться успешно. `origin` обязан указывать именно на `aleka7sk/belcanto-product` через допустимый HTTPS или SSH URL. Если remote не совпадает, branch/ref отсутствует, baseline не является предком существующей `agent/production-v2`, local и remote implementation branches разошлись либо команда проверки завершилась ошибкой — остановись **до** branch creation, install, generation, migration, commit и push.

2. Не используй destructive git-команды, не делай reset/rebase/force-push/stash/clean и не открывай или не сливай PR без отдельного указания. После каждого полностью проверенного vertical-slice checkpoint делай осмысленный commit и push в `agent/production-v2`; не коммить заведомо падающее промежуточное состояние. Непосредственно перед каждым push повтори `git fetch --prune origin`, проверь, что текущая ветка — ровно `agent/production-v2`, staged diff относится только к текущему slice, `git diff --cached --check` проходит, baseline остаётся предком ветки, а push является fast-forward относительно `origin/agent/production-v2` (если remote branch уже существует). При любом несоответствии остановись без push.

3. Прочитай полностью и соблюдай в таком порядке:

   - `CLAUDE.md`
   - `docs/project/project-constitution.md`
   - `meta/standard/008-agent-change-protocol.md`
   - `meta/002-authority-model.md`
   - `meta/007-ownership-map.md`
   - `meta/001-reading-order.md`, особенно обязательное чтение по типу задачи
   - `meta/006-validation-rules.md`
   - `meta/011-conformance-statement.md`
   - `meta/standard/006-traceability.md`
   - `meta/standard/009-conflict-detection.md`
   - `governance/decisions/002-pending-decisions.md`
   - затрагиваемые `school/`, `product/`, `language/`, `domain/`, `experience/`, `features/`, `design/`

4. Если меняешь нормативные документы, дополнительно прочитай `meta/004-anti-duplication-rules.md`, `meta/005-anti-contradiction-rules.md`, `meta/009-contribution-workflow.md`, `meta/012-authoring-binding.md` и `meta/standard/017-authoring-pipeline.md`.

5. Явно учти решения:

   - `PD-0002` — PEOS остаётся внешним инженерным эталоном; не копируй несуществующий корпус PEOS и не выдумывай его требования;
   - `PD-0003` — production vertical slices предпочтительнее горизонтальной реализации;
   - `PD-0030` — закрытая activation и server-side permission boundaries;
   - `PD-0031` — реализация живёт в этом monorepo; сохраняй Expo/React Native + modular Go API + PostgreSQL, не вводи микросервисы/auth-service/event bus без отдельного принятого решения;
   - `PD-0032` — Product владеет смыслом, Language словами, Domain структурой и инвариантами;
   - `PD-0033` — актуальная Figma является визуальным источником.

6. Полностью прочитай Approved product foundation:

   - `product/000-product-overview.md`;
   - `product/001-product-vision.md`;
   - `product/001-vision-and-principles.md`;
   - `product/002-product-boundaries.md`;
   - `product/003-personas.md`;
   - `product/006-business-goals.md`;
   - `product/011-capability-map.md`.

   Для учебных возможностей прочитай все применимые Approved owners и их `Depends On / Required By`, включая `domain/architecture/000-domain-model-rules.md`, `domain/entities/000-entity-catalog.md`, `domain/lesson.md`, `domain/homework.md`, `domain/progress.md`, `domain/assessment.md`, `domain/services/000-domain-service-catalog.md` и Approved policies.

7. Не считай документы со статусом `Scaffold` или `Draft` утверждёнными требованиями. В частности, `experience/**`, `features/**`, `design/**` и `roadmap/001-mvp-scope.md` не могут урезать production scope Figma. Draft используй только с явной risk note. `ai/001–ai/005` имеют статус `Superseded` по `PD-0026` и authority не имеют. Не исправляй попутно известные corpus warnings, если они не вызваны твоими изменениями.

8. Новая директива Product Owner: продукт больше не ограничен прежним MVP scope. Не используй старый out-of-scope список, чтобы исключить функции из Figma Pages `19–37`. При нормативном конфликте не переписывай Canonical/Approved документы молча: примени authority/change protocol. Для изменения класса C/D подготовь только предложение по DES-008; не создавай, не редактируй и не «принимай» PDR/decision record от имени владельца. Заблокируй зависимые slices до решения роли и продолжай все независимые незаблокированные vertical slices.

## 2. Источник истины в Figma

Сначала изучи весь файл и особенно Pages `19–37`. Используй Figma MCP/design context; не реализуй интерфейс только по двум screenshots или по памяти.

### Приоритет источников

1. Репозиторные constitution, security и authority rules определяют, **как безопасно менять систему**.
2. Page `37 · Handoff & Full Coverage` — текущий implementation contract и правило разрешения расхождений.
3. Pages `21–34` — канонические production UI, состояния и role projections.
4. Page `35` — reusable production component families и variants.
5. Page `36` задаёт только порядок шагов, recovery, deep-link и transition intent для 16 flows. Из-за ограничения Figma на cross-page destinations её узлы являются prototype-only same-page semantic replicas, связанными с каноническими screen keys Pages `21–34`. Не считай и не реализуй их как дополнительные screens/routes или визуальный источник.
6. Pages `19–20` — merge audit, product contract, IA, navigation и permissions.
7. Pages `03–08` — исходный premium visual/UX DNA и legacy feature inventory, уже сведённый в Page 19.
8. Pages `10–18` — история функциональной декомпозиции, а не production UI для копирования.

Не реализуй Pages `03–08` и `10–18` как параллельные версии приложения. Их полезный функционал уже должен иметь каноническое место в Pages `19–37`. Если найдёшь реальный legacy-only gap, сначала внеси его в traceability matrix и сопоставь с существующим production flow; не копируй старый экран буквально.

### Фактическое покрытие

| Page | Область | Production screens |
|---|---|---:|
| 21 | Student Today & Home | 8 |
| 22 | Profile, Progress, Skills & Achievements | 12 |
| 23 | Practice, Homework & Repertoire | 16 |
| 24 | Schedule, Lessons, Events & Reschedule | 32 |
| 25 | Teachers & School | 10 |
| 26 | Teacher Today & Lesson Journal | 19 |
| 27 | Teacher Students, Review & Analytics | 16 |
| 28 | Community, Chat & Safety | 19 |
| 29 | Administrator Operations | 33 |
| 30 | Owner Overview, Analytics & Governance | 24 |
| 31 | Notifications, Activity & Deep Links | 9 |
| 32 | Auth, Account, Security, Roles & Privacy | 31 |
| 33 | System States, Offline & Sync | 8 |
| 34 | Accessibility, Localization & Motion | 5 |
| **Всего** | | **242** |

Page 28 включает отдельную Teacher Community projection. Вход в Community сохраняет активную роль и соответствующий Bottom Navigation shell; это новый semantic state, а не новая глобальная роль.

Дополнительно:

- 9 production component families;
- 20 role-aware Bottom Navigation variants;
- 16 cross-role flows;
- 96 interactive flow steps;
- 22 handoff/coverage boards.

Все 242 Figma frames — это **покрытие экранов и состояний**, но не приказ создать 242 route-файла. Объединяй логически родственные frames в state-driven screens, sheets, dialogs и variants. Каждый semantic state должен быть достижим, протестирован и соответствовать каноническому user flow.

До кодинга создай рабочую traceability matrix:

`домен → роли/permissions → Figma frames → route/state → entities/storage → API → error/offline states → tests`

Не создавай новый нормативный продуктовый документ только ради этой матрицы, если репозиторный protocol этого не допускает. Веди её в task plan или разрешённом implementation artifact.

Сначала выполни reuse/gap audit существующих B.0/L.1 vertical slices на baseline. Сохраняй уже реализованные и проверенные invitation/activation, premium UI, internal scheduling, student-safe projections и iOS runtime fixes. Реализуй только подтверждённые gaps или необходимые эволюционные изменения; не переписывай рабочий слой лишь ради соответствия новому имени или структуре.

## 3. Визуальный и UX-контракт

Сохрани premium-характер Pages `03–08`, но реализуй более зрелую архитектуру Pages `19–37`:

- почти чёрный canvas и layered ink surfaces;
- Onest;
- violet — primary action/selection;
- cyan и gold — функциональные акценты;
- magenta — редкий semantic accent, не фон по умолчанию;
- ограниченный glow, без визуального шума;
- крупные уверенные заголовки, ясная hierarchy, спокойная плотность;
- 390×844 reference canvas, но responsive layout для реальных устройств;
- safe areas, keyboard avoidance, scroll recovery;
- minimum touch target `48×48 pt`;
- статус никогда не передаётся только цветом;
- каждая RU/KK projection и её shared shell используют один целевой locale; пользовательские `scope`, `occurrence`, `audit`, `mutation`, `last-write-wins`, `privacy`, `waitlist`, `timezone` и непереведённые English section labels запрещены;
- разрешены только утверждённые proper/platform terms, например Belcanto, Face ID, 2FA, iPhone и App Store;
- localization распространяется на Bottom Navigation и весь shared shell, а не только content; multiline row/helper copy должна расти или переноситься без clipping;
- primary content должно поддерживать platform-native large text и reflow без clipping;
- на iOS применяй Dynamic Type к content. Для компактной навигации используй системный tab bar либо Large Content Viewer для каждого custom item с полными локализованными label и symbol; Large Content Viewer не заменяет Dynamic Type для основного content и не должен имитироваться tooltip/modal;
- на Android проверяй весь custom shell при системном `fontScale = 2.0`: не clamp’и font scale, разрешай рост/reflow/scroll и сохраняй полные accessibility labels;
- screen reader order/labels/actions, contrast и Reduce Transparency/Motion должны оставаться осмысленными;
- dark/light semantic aliases должны идти через общие tokens, даже если production UI dark-first.

Page `35` — единственный канонический extension layer для новых UI patterns. Переиспользуй:

- Premium Context Hero;
- Growth Signal;
- Evidence Card;
- Agenda Entry;
- Explainable Insight;
- Lesson Recap;
- Bottom Navigation;
- Date Chip;
- RSVP Control.

Не создавай вторую screen-local design system. Общие primitives/tokens/components расширяй централизованно в `apps/mobile/src/ui/`.

Обязательно:

- обнови устаревший `figmaVisualSource` в `apps/mobile/src/ui/tokens.ts`: ссылки `4:4` и `4:5` больше не являются implementation source;
- замени PNG-based tab navigation в `apps/mobile/src/ui/lessonComponents.tsx` на нативную, доступную, role-aware navigation из Pages `20/35`;
- не удаляй premium imagery механически: отдели декоративные raster assets от системной icon/navigation layer;
- не хардкодь значения, которыми уже владеют semantic tokens;
- используй стабильные component APIs и variants вместо копирования JSX между screens.

Code Connect недоступен для текущего Figma seat. Это не блокер: Page `37`, board `HOF-16`, содержит ручной Figma-to-code mapping contract.

## 4. Роли, permissions и приватность

Роли:

- **Student** — только собственные обучение, расписание, practice, progress, RSVP и разрешённое community.
- **Teacher** — назначенные ученики/группы, attendance, lesson journal, reviews, event roster и делегированные действия.
- **Administrator** — расписание, серии, помещения, ученики, группы, события, requests, конфликты, operational notifications и audit; не подменяет педагогическую оценку.
- **Owner** — обзор школы, aggregate analytics, команда, роли, policies, governance и exports; student-level learning content по умолчанию не показывать.
- **Moderator** — permission set/capability, а не пятая глобальная роль.

Правила:

- permissions проверяются сервером для каждого query/mutation; скрытый UI не является защитой;
- capability выводится из tenant-scoped grants/policies, а не только из строки `Role`;
- role switch перестраивает navigation/routes, сбрасывает role-scoped cache и не переносит запрещённые данные;
- открытие Community не является role switch: сохраняй текущую Student/Teacher/Administrator/Owner роль, её permissions, routes, role-scoped cache и Bottom Navigation shell; меняется только разрешённая Community projection;
- Teacher Community на Page 28 использует существующий Teacher principal, Teacher permissions и Teacher Bottom Navigation при входе, refresh, deep link и возврате. Не добавляй `Community` в role enum, RoleGrant, session claims или отдельный role-scoped cache;
- tenant isolation обязательна на Store/Service/SQL/HTTP уровнях;
- Student-safe projections не раскрывают peers основной группы, staff notes, capacity или чужие identifiers;
- Owner analytics aggregate-first; малые выборки показывают `insufficient data`, а не deanonymized details;
- запрети удаление/понижение последнего Owner;
- нет публичной регистрации;
- нет unrestricted Student↔Student direct messages;
- invitation, consent, session, moderation и permission mutations должны иметь audit history.

## 5. Обязательная доменная модель

Следуй Ubiquitous Language и существующему ownership map. Названия ниже описывают необходимые понятия, а не разрешают создать вторые определения.

### Identity, account и governance

- Tenant/School;
- User/Identity;
- Membership, RoleGrant, PermissionGrant;
- Invitation и закрытая activation;
- Session/RefreshToken family, revoke и security event;
- verified contacts, 2FA, biometrics capability;
- Consent, PolicyVersion, Acceptance;
- privacy settings;
- data export request;
- account deletion request;
- audit event, outbox event, idempotency record.

### Основные занятия

Раздели:

- `CoreLessonSeries`;
- `CoreLessonOccurrence`;
- individual/group format;
- enrollment/assignment;
- attendance;
- room/teacher allocation;
- reschedule/cancellation request;
- immutable schedule history.

Инварианты:

- individual core lesson — ровно один ученик;
- group core lesson — capacity не больше 3;
- target roster группы может быть 2–3, но временно неполная группа не превращается автоматически в individual;
- server и DB запрещают roster >3;
- Student projection никогда не показывает состав группы или её staff-only capacity;
- lifecycle минимум: scheduled, started, completed, missed, cancelled, rescheduled;
- history и audit не перезаписываются при переносе.

Текущий generic `Lesson` допускает до 100 участников и только `scheduled`; его нельзя просто переиспользовать без корректной миграции модели и существующих данных.

Для Figma-aligned seed, screenshot и visual-test data сохраняй точное role mapping:

- `Коркем` — ведущий педагог;
- `Айгерим` — педагог групп по вторникам; fixture metadata: `Вокал · группы по вторникам`, день `Вт`;
- `Шугыла` — замещающий педагог;
- event specialist остаётся generic role `Специалист Belcanto`;
- administrator/support остаётся generic role `Координатор Belcanto`.

Не выполняй глобальную замену placeholder-имён: `Алиса` использовалась в разных ролях. Не назначай Айгерим event specialist и не переназначай вторничные группы. Это reference data для handoff/tests, а не hardcoded production UI; runtime-данные приходят из persistence/API.

### Неосновные занятия-события

События — отдельная сущность и не расходуют баланс основных уроков:

- `EventSeries`;
- `EventOccurrence`;
- teacher/host, room, duration;
- configurable event capacity, независимая от лимита `3`;
- weekly recurrence;
- occurrence override;
- RSVP;
- WaitlistEntry;
- timed SpotOffer;
- cancellation и participant notification.

Категории, которые должны поддерживаться данными, а не hardcoded UI:

- мастер-класс;
- актёрское мастерство;
- йога;
- караоке-пати;
- будущие администраторские категории.

Инварианты:

- RSVP относится к конкретному occurrence;
- confirmed RSVP добавляет событие в `My Schedule`;
- cancel RSVP удаляет его только из личного расписания, но не из каталога и не из серии;
- последнее место резервируется атомарно;
- full/waitlist/spot-offer race обрабатывается транзакционно;
- offer имеет server-side expiry и idempotent confirm;
- recurrence edit имеет ровно два scope: `this_occurrence` и `this_and_following`;
- `this_and_following` разделяет серию на границе изменения;
- прошлые occurrences, exceptions, attendance, RSVP и audit сохраняются;
- Student schedule визуально объединяет core lessons и подтверждённые events, но backend не смешивает сущности.

### Progress, skills и achievements

Progress заполняет преподаватель, а ученик видит понятное развитие в профиле.

Обязательные 5 областей:

1. Голос;
2. Интонация;
3. Музыкальность;
4. Сцена;
5. Самостоятельность.

Под ними могут быть skills/descriptors, например «Дыхание и опора».

Инварианты:

- evidence-backed, explainable и versioned;
- anchored descriptors вместо неясной цифры;
- baseline, periodic review, pulse/history, goals, evidence и achievements;
- каждый вывод ведёт к конкретному lesson journal, recording, performance или teacher note;
- нет общего opaque score, ranking учеников или скрытой формулы;
- published review исправляется новой revision с reason, author и timestamp, а не destructive overwrite;
- Student видит «что изменилось», «почему», evidence и следующий шаг;
- Teacher видит workflow review/correction;
- Owner видит только aggregate quality/freshness/coverage с small-sample guard.

### Lesson journal и единый педагогический цикл

Один teacher journal flow должен собирать:

- attendance;
- lesson goal и summary;
- private staff note отдельно от student-visible recap;
- evidence;
- progress updates;
- repertoire status;
- homework;
- achievements;
- next lesson focus.

Один publish action должен согласованно создать/обновить student recap, evidence, progress, repertoire, homework и achievements. Не допускай частично опубликованного состояния. Используй PostgreSQL transaction для authoritative state + append-only audit/outbox и idempotent projection/delivery processing.

Correction после publish — новая revision с reason. История видна согласно permissions.

### Practice, homework и repertoire

- Homework assignment, steps, due/expiry policy;
- practice session;
- recorder/upload lifecycle;
- submission versions;
- teacher feedback;
- retry/revision;
- accepted evidence;
- repertoire item/journey/status/history.

Offline draft разрешён для teacher journal и practice recording metadata. Upload обязан иметь resumable/retry/failure states. Sensitive media URLs не попадают в audit/outbox; используй production storage adapter и короткоживущие signed access URLs.

### Community и safety

- feed, announcement, event-linked content;
- post/thread/comment;
- role/permission-aware audience;
- scoped chat/conversation;
- draft/upload/offline recovery;
- report, block, restricted access;
- moderation queue/action/reason/audit;
- consent checks для media.

Не добавляй unrestricted student discovery или Student↔Student DM. Moderator — delegated capability. Block/report/deleted/no-access states должны безопасно восстанавливаться через deep link.

### Notifications и activity

- in-app activity;
- categories/preferences;
- push permission rationale;
- quiet hours;
- per-channel preference;
- privacy-safe preview;
- OS denied;
- deep link и fallback;
- delivery attempt/retry/dead-letter observability.

Одна запись в outbox без worker/consumer, retry policy и delivery status не считается реализованным notification flow.

## 6. Auth, security и data rules

- Invite-only activation; public signup отсутствует во всех route/API projections.
- Не храни raw invitation/session/reset tokens, secrets или sensitive media URLs в audit, logs, outbox или idempotency records.
- Сохрани rotating refresh tokens, replay detection и session revoke semantics.
- Native secrets — Secure Store; web access token — memory-only.
- Не вводи обход permissions через client-supplied role/tenant.
- Все mutations имеют stable idempotency key и payload fingerprint.
- Для versioned entities используй optimistic expected version; last-write-wins запрещён для schedule, RSVP capacity, progress, journal corrections, roles и policies.
- Authoritative mutation атомарно пишет domain state + append-only audit + outbox.
- PostgreSQL migrations должны быть additive, checksum-protected, concurrency-safe и сохранять существующие данные; destructive migration требует отдельного явного плана и authority.
- Ошибки API имеют стабильные typed codes; mobile decoder не принимает неизвестную форму молча.
- Time storage — UTC; API boundary — RFC3339; отображение и recurrence zone — IANA `Asia/Almaty`.
- Не полагайся на локальный device time для expiry, last seat, session или offer.
- Pagination/filtering/search должны быть server-backed для растущих списков.
- Не логируй PII/learning content в analytics.

## 7. Offline, sync и recovery

Offline-first не означает «всё можно менять офлайн».

Разрешённый offline draft:

- teacher lesson journal draft;
- practice/recording draft;
- community composer draft, если privacy/audience уже разрешены.

Требуют сети:

- RSVP/waitlist/spot offer;
- schedule/reschedule mutation;
- journal publish;
- progress publish/correction;
- permission/security/policy mutation;
- destructive account operations.

Обязательно:

- cached/read states с явной freshness;
- retry queue только для безопасных idempotent commands;
- pending/synced/failed/conflict statuses;
- duplicate mutation suppression;
- optimistic version conflict;
- user-readable merge/recovery для journal;
- upload resume/restart;
- deep-link fallback при deleted/no-access/expired destination;
- никогда не выдавать stale data за актуальные.

## 8. Архитектурный контракт репозитория

Сохрани существующий стек и границы:

- Expo SDK 57;
- React Native 0.86.2;
- React 19.2.3;
- TypeScript 6 strict;
- Expo Router;
- Reanimated 4;
- Onest;
- Go 1.26.5, `net/http`, `pgx/v5`;
- PostgreSQL 18.4;
- manual OpenAPI contract.

Ключевые пути:

- `apps/mobile/app/_layout.tsx`
- `apps/mobile/app.config.ts`
- `apps/mobile/src/session/`
- `apps/mobile/src/session/machine.ts`
- `apps/mobile/src/session/provider.tsx`
- `apps/mobile/src/access/capabilities.ts`
- `apps/mobile/src/access/routeGuards.ts`
- `apps/mobile/src/accessibility/policy.ts`
- `apps/mobile/src/api/contracts.ts`
- `apps/mobile/src/api/routes.ts`
- `apps/mobile/src/api/client.ts`
- `apps/mobile/src/controllers/`
- `apps/mobile/src/controllers/idempotency.ts`
- `apps/mobile/src/ui/tokens.ts`
- `apps/mobile/src/ui/components.tsx`
- `apps/mobile/src/ui/lessonComponents.tsx`
- `apps/mobile/src/ui/primitiveContracts.ts`
- `apps/api/internal/core/core.go`
- `apps/api/internal/app/service.go`
- `apps/api/internal/app/store.go`
- `apps/api/internal/store/memory/memory.go`
- `apps/api/internal/store/postgres/`
- `apps/api/internal/httpapi/http.go`
- `apps/api/internal/httpapi/handlers.go`
- `apps/api/migrations/`
- `apps/api/migrations/embed.go`
- `contracts/http/belcanto-v1.openapi.yaml`

Route-файлы Expo Router должны оставаться тонкими adapters. Business/data orchestration — в controllers/use-cases; API parsing — в strict decoders; permission-driven presentation — в feature screens/components.

Не создавай монолитные 500–1000-line screen components. Группируй код по domain/feature boundaries, используй typed view models и стабильные component contracts.

Каждую способность реализуй полным vertical slice:

```text
safe migration + DB constraints
→ core/domain types and permission rules
→ Store interface
→ memory + PostgreSQL parity
→ Service validation, transactions, idempotency and versions
→ HTTP handlers + typed errors
→ OpenAPI
→ TypeScript contract + strict decoder
→ ApiClient + controller
→ permission-driven production UI
→ unit + integration + interaction/E2E tests
```

Не реализуй сначала «весь frontend», а backend позже. Не оставляй production screens на локальных fixture/mock arrays.

Новые зависимости добавляй только при доказанной необходимости, с зафиксированной версией и без дублирования уже существующей capability.

## 9. Обязательные 16 end-to-end flows

Каждый flow должен иметь start, success, back/cancel, forbidden и recovery path согласно Page 36.

Для routes, deep links, analytics, snapshots и traceability используй только канонические screen keys Pages `21–34`. Node IDs prototype-only replicas Page 36 допустимы только в prototype-audit metadata.

| Flow | Обязательный путь | Критические recovery |
|---|---|---|
| A | Invite → activation → password/2FA → session → role → role-aware Today | expired invite, verification failure, access pending |
| B | Account → Security → Sessions → re-auth → revoke → security event | current session, stale session list |
| C | Student Today → lesson → recap → next action | recap unavailable, cancelled/rescheduled lesson |
| D | Progress → domain/skill → evidence → goal | history/version, unavailable evidence |
| E | Homework → practice → record/upload → submit → feedback | offline draft, upload fail, retry, revision |
| F | Event catalog → occurrence RSVP → My Schedule → cancel | last-seat race, full, conflict |
| G | Waitlist → timed offer → confirm | expired offer, already taken |
| H | Teacher Today → attendance → journal → publish | offline draft, concurrent conflict, partial delivery recovery |
| I | Review → publish → correction | expected-version failure, revision reason |
| J | Core series → individual/group → assignment | group capacity 3, room/teacher conflict |
| K | Event series → weekly recurrence → edit | occurrence vs this-and-following, exceptions |
| L | Reschedule request → options/decision → notification | concurrent schedule update, rejected request |
| M | Active-role shell → permitted Community projection → report/block → moderation | restricted/deleted content, audit |
| N | Notification → exact deep link | no permission, deleted target, fallback |
| O | Owner insight → decision → tracked impact | insufficient sample, stale/partial data |
| P | Offline → reconnect → sync recovery | duplicate command, conflict, upload retry |

## 10. Analytics и observability

Реализуй privacy-safe telemetry для:

- activation funnel;
- lesson/journal completion;
- recap viewed;
- homework started/submitted/retried;
- evidence/progress review viewed;
- event catalog → RSVP/waitlist/offer/cancel;
- reschedule request lifecycle;
- notification delivery/open/deep-link recovery;
- offline queue/sync/conflict;
- moderation lifecycle;
- owner insight freshness/decision.

Требования:

- stable event names/version;
- tenant/user correlation только через безопасные internal identifiers;
- без raw lesson notes, recordings, messages или contact data;
- request/command correlation IDs;
- operational metrics для outbox lag, notification failure, upload failure, conflict rate;
- stale/partial/insufficient data явно отражаются и в UI, и в telemetry.

## 11. Открытые P0 decisions

Не хардкодь и не выдавай за решённые:

1. точную late-cancellation policy и последствия;
2. guardian consent policy для несовершеннолетних и community media;
3. legal retention periods и export/deletion SLA.

Сделай модель расширяемой:

- policy/version/config entities;
- permission/consent gates;
- explicit safe disabled/restricted states;
- audit;
- future effective dates.

До решения не запускай необратимое автоматическое удаление и не заявляй legal compliance. Эти три вопроса не должны блокировать реализацию остальных slices.

Пока хотя бы одно из этих решений открыто, зависимые функции должны оставаться в явно безопасном `disabled/restricted` состоянии, а общий статус не может быть `release-ready`, `production-ready` или `GO`. Допустимый итог до решений: `IMPLEMENTATION COMPLETE FOR UNBLOCKED SCOPE — RELEASE BLOCKED BY P0`, с точным перечнем зависимых capabilities, экранов, API и тестов. HOF-21/G0 проходит только после закрытия всех трёх P0 решений.

## 12. Порядок исполнения

Работай вертикальными фазами; после каждой фазы запускай targeted tests и обновляй traceability matrix.

### Continuity и checkpoints

- Работа рассчитана на несколько сессий; не пытайся скрыть незавершённость ради одного финального ответа.
- После обязательного `artifact-value-gate` создай или продолжай machine-readable non-normative ledger `.claude/implementation/production-v2-checkpoint.json`. Он фиксирует только состояние реализации, а не вводит продуктовые или доменные решения. Минимальные поля: `schemaVersion`, `baselineSha`, `branch`, `updatedAt`, `phase`, `slices[]` со статусами `completed/in_progress/blocked/not_started`, canonical Figma screen keys, migrations, API, mobile paths, tests, exact commands/results, blockers, `lastGreenCodeSha` и `nextSafeAction`.
- После каждого зелёного vertical-slice checkpoint сначала обновляй task plan, проверяй slice и коммить код/контракты/тесты. Затем возьми полученный полный code commit SHA, обнови ledger с этим `lastGreenCodeSha`, сделай отдельный metadata-only checkpoint commit и push обоих fast-forward commits в `agent/production-v2`. Не пытайся записать SHA коммита внутрь самого этого же коммита.
- При возобновлении сессии сначала сверяй remote branch, последние commits и worktree, затем читай committed ledger и проверяй, что указанный `lastGreenCodeSha` существует, является предком текущей ветки и соответствует записанным test gates; продолжай с последнего подтверждённого checkpoint, не начинай Phase 0 заново. Если ledger противоречит Git/test evidence, доверяй evidence и исправь ledger до дальнейшей реализации.
- Если приближается лимит контекста или runtime, закончи текущую атомарную операцию, не оставляй незавершённую migration/contract pair, зафиксируй точный статус `completed / in progress / blocked / not started`, команды и следующий безопасный шаг.
- Не объявляй checkpoint зелёным, если соответствующий migration/API/mobile/test vertical slice неполон или обязательная проверка не прошла.

### Phase 0 — audit и foundation

- прочитать authority/doc contract;
- изучить Pages 19–37;
- выполнить исходный corpus validator и сохранить baseline findings: на момент передачи ожидается `errors: 0`, `99 warnings`, из них `77 known/accepted` и `2 skipped/partial`; exit code 0 не означает, что corpus не имеет известных предупреждений;
- сопоставить существующий код с 242 states;
- зафиксировать migration/data compatibility plan;
- определить shared API errors, idempotency, versioning, audit/outbox, pagination, media/notification adapter boundaries;
- обновить implementation task plan.

### Phase 1 — identity, account, roles, privacy

- invite-only auth;
- session/security/2FA/biometrics capability;
- account/profile/contacts;
- role switch/capabilities;
- privacy/consents/export/deletion request;
- Pages 31–32 relevant deep links and notifications.

### Phase 2 — schedule, core lessons и events

- core series/occurrences, groups, rooms, teachers;
- event series/occurrences, recurrence exceptions;
- RSVP, waitlist, spot offers;
- reschedule requests;
- Student/Teacher/Admin projections;
- Page 24 + scheduling/event flows F/G/J/K/L.

### Phase 3 — Student learning loop

- Today/Home;
- progress/skills/evidence/goals/achievements;
- homework/practice/recordings/feedback;
- repertoire;
- teachers/school;
- flows C/D/E.

### Phase 4 — Teacher learning operations

- Today/context;
- attendance;
- journal draft/publish/correction;
- students/groups;
- periodic review;
- event roster;
- flows H/I.

### Phase 5 — Community, notifications и safety

- feed/thread/composer;
- chat boundaries;
- report/block/moderation;
- notification preferences/delivery/deep links;
- flows M/N.

### Phase 6 — Administrator и Owner

- operational calendars, conflicts, requests, people/groups/rooms;
- series lifecycle and audit;
- owner health/learning quality/retention/utilization/workload/revenue;
- roles/policies/overrides/data quality/export;
- insufficient/stale/partial states;
- flow O.

### Phase 7 — reliability, accessibility, localization, motion

- Page 33 states;
- offline/cache/sync/upload recovery;
- выполнить полный accessibility/localization contract из §3 для RU/KK, pseudolocale, shared shell, platform-correct large text, screen reader, contrast и Reduce Transparency/Motion;
- security/privacy QA;
- analytics/observability;
- flow P.

### Phase 8 — E2E, visual QA и release readiness

- все 16 flows;
- iOS/Android representative visual comparison with Pages 21–35;
- deep links;
- migration/replay/concurrency tests;
- CI-equivalent gates;
- финальная traceability matrix без красных обязательных строк.

Не объявляй фазу завершённой, если слой ниже UI остался mock/stub.

## 13. Тестовый контракт

Добавь и пройди:

- migration schema, constraints, upgrade compatibility и безопасное rollback/down поведение, где поддерживается;
- core/service unit tests;
- permission и tenant boundary tests;
- Teacher → Community tests для entry, refresh, deep link и back: Teacher principal, permissions, cache scope и Bottom Navigation сохраняются; глобальная роль Community не создаётся;
- idempotent replay с тем же payload и rejection при другом payload;
- expected-version/stale write;
- transaction atomicity;
- memory/PostgreSQL parity;
- PostgreSQL integration/concurrency tests для last seat, group capacity, recurrence split, journal publish и last Owner;
- HTTP role/tenant/error/idempotency tests;
- OpenAPI lint и response/schema parity;
- strict mobile decoder tests;
- seed/snapshot tests для reference role mapping: Коркем = ведущий педагог, Айгерим = группы по вторникам, Шугыла = замещающий педагог; specialist/support остаются generic Belcanto roles;
- ApiClient/controller/session/access tests;
- UI interaction, loading/empty/error/offline/forbidden/conflict tests;
- accessibility labels/order/actions;
- RU/KK и pseudolocale expansion checks по всему content + shared shell, включая Bottom Navigation и multiline rows;
- iOS Dynamic Type checks на accessibility content-size categories и Large Content Viewer interaction для compact custom navigation;
- Android checks при системном `fontScale = 2.0` по всему custom shell без clipping, overlap или font-scale clamping;
- all A–P flow tests;
- visual QA iOS/Android.

Не уменьшай baseline существующих mobile tests: на момент передачи `16 suites / 101 tests`.

## 14. Обязательные команды

```bash
test "$(corepack pnpm --version)" = "11.7.0"
corepack pnpm install --frozen-lockfile

python3 .claude/scripts/validate.py
corepack pnpm contract:lint
corepack pnpm mobile:check
corepack pnpm --filter @belcanto/mobile exec expo-doctor

make api-format
make api-tidy
make api-vet
make api-build
: "${TEST_DATABASE_URL:?set TEST_DATABASE_URL to a verified disposable PostgreSQL 18.4 database}"
APP_ENV=test DATABASE_URL="$TEST_DATABASE_URL" make migrate
make api-test
make api-race

git diff --check
git status --short
```

Для integration tests используй disposable PostgreSQL `18.4`. До migration выведи только sanitised host/port/database target без credentials и докажи, что target disposable. Не наследуй существующий `DATABASE_URL`; если disposable target подтвердить нельзя, migration не запускай. Для CI-equivalent инструментов используй версии и action commits, закреплённые в текущем `.github/workflows/ci.yml`, а не `latest`. Затем выполни CI-equivalent проверки из `.github/workflows/ci.yml`, включая:

- `govulncheck`;
- production Docker build;
- Trivy scan.

Если конкретный runtime/simulator/credential реально недоступен, не подменяй проверку утверждением «должно работать». Зафиксируй точную недоступную команду, причину и оставшийся verification step. Все остальные проверки выполни. Для `expo-doctor` проверь не только exit code, но и сам вывод: строка `Error:` означает провал.

## 15. Definition of Done для незаблокированного scope и release gate

Незаблокированный implementation scope достигает checkpoint-complete только когда одновременно выполнено:

- все обязательные домены, не зависящие от открытого P0, имеют реальный persistence/API path;
- нет public signup;
- server-side permissions и tenant isolation закрыты тестами;
- core lessons и events — разные сущности;
- individual = 1, group capacity ≤3;
- Student-safe projection не раскрывает группу;
- RSVP/last seat/waitlist/offer атомарны;
- recurrence split сохраняет прошлое/exceptions/RSVP/audit;
- progress evidence-backed/versioned, без overall score/ranking;
- journal publish не создаёт partial learning state;
- published data исправляется revisions;
- offline/recovery не теряет и не дублирует commands;
- notification и media flows имеют production adapter/worker/retry boundaries;
- каждый из 242 semantic states покрыт route/state mapping; P0-зависимые состояния явно помечены `BLOCKED` и не считаются завершёнными;
- все незаблокированные ветви flows A–P достижимы и имеют recovery; P0-зависимые ветви перечислены отдельно и не входят в completed count;
- Page 35 patterns используются как единая UI system;
- minimum target 48 pt и полный accessibility/localization contract из §3 проверены, включая Bottom Navigation и multiline copy;
- representative layouts проверены как минимум на ширинах 320/360/390/430;
- в незаблокированном scope нет placeholder, TODO, static-only или permanently feature-flagged обязательного flow;
- feature flags могут управлять rollout, но не маскировать незавершённую незаблокированную capability;
- OpenAPI, mobile, Go, PostgreSQL, security и CI gates пройдены;
- все незаблокированные строки traceability matrix зелёные; строки, зависящие от трёх P0 decisions из раздела 11, явно имеют статус `BLOCKED` и безопасное restricted/disabled поведение;

### Release gate

Статус `PRODUCTION READY — GO` разрешён только когда дополнительно:

- все три P0 decisions из раздела 11 закрыты человеком и записаны по authority/change protocol;
- все ранее `BLOCKED` capabilities, states, API paths и tests реализованы и стали зелёными;
- все 242 semantic states и все 16 flows A–P полностью достижимы, включая recovery, без P0-disabled заменителей;
- traceability matrix полностью зелёная без обязательных `BLOCKED` строк;
- HOF-21 gates G0–G8 имеют подтверждённый `PASS`.

До выполнения release gate допустим checkpoint `IMPLEMENTATION COMPLETE FOR UNBLOCKED SCOPE — RELEASE BLOCKED BY P0`, но он не означает готовность production release.

## 16. Формат работы и финального отчёта

Сначала выведи:

1. подтверждённую исходную branch/commit/worktree;
2. прочитанные authority documents и найденные конфликты статусов;
3. краткую coverage/gap matrix;
4. фазовый implementation plan с test gates.

Затем **без остановки после плана** выполняй фазы 0–8. Не спрашивай про детали, которые уже определены Figma/HOF contracts. Эскалируй настоящий P0 blocker, а также любой случай, когда DES-008/DES-009 требует решения роли, обязательный authority source недоступен или конфликт запрещено разрешать агенту. Во всех случаях продолжай независимые slices.

В финале сообщи:

- что реализовано по доменам и flows;
- какие migrations/API/routes/components/tests добавлены;
- какие старые реализации мигрированы или удалены;
- результаты каждой обязательной команды;
- visual/accessibility QA matrix;
- сравнение финального corpus validator с исходным baseline без новых errors/warnings/known findings;
- открытые P0 decisions;
- итоговый статус строго одним из четырёх: `PRODUCTION READY — GO`; `IMPLEMENTATION COMPLETE FOR UNBLOCKED SCOPE — RELEASE BLOCKED BY P0`; `IN PROGRESS — CHECKPOINT <phase/slice>`; `BLOCKED — <exact blocker>`;
- последний запушенный checkpoint branch и полный commit SHA;
- остаточные риски без общих формулировок;
- точный `git status --short`.

Не называй работу завершённой после одного красивого экрана, после frontend-only pass или после unit tests без PostgreSQL/concurrency/E2E/visual verification.

---

## Reference standards для инженерной проверки

Это внешние reference standards, а не источник бизнес-фактов школы:

- Apple Human Interface Guidelines — Tab bars: https://developer.apple.com/design/human-interface-guidelines/tab-bars
- Material 3 — Navigation bar: https://m3.material.io/components/navigation-bar/guidelines
- WCAG 2.2 — Target Size Minimum: https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum
- WCAG — Contrast Minimum: https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum
- Android Offline-first data layer: https://developer.android.com/topic/architecture/data-layer/offline-first
- Google Calendar recurring events: https://developers.google.com/workspace/calendar/api/guides/recurringevents
- Android pseudolocales: https://developer.android.com/guide/topics/resources/pseudolocales
- Закон Республики Казахстан «О персональных данных и их защите»: https://adilet.zan.kz/rus/docs/Z1300000094
