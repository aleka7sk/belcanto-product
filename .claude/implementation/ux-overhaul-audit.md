# UX Overhaul — аудит и целевая архитектура

> Рабочий артефакт реализации (non-normative), сессия mobile-ux-overhaul.
> Не корпус: продуктовые решения здесь не живут. Ворота ценности пройдены:
> артефакт делает однозначными решения миграции scaffold/токенов/навигации.
> База: ветка `agent/production-v2`, HEAD `50f5130`, гейт зелёный
> (mobile:check 40 suites / 214 tests, eslint 0 warnings, tsc strict OK).
> Источник визуальной правды: Figma `yXE7a9vAyWdbU9iLnjFmXf`, Pages 19–37.
> Методика: seed-факты промпта перепроверены по коду построчно; три
> параллельные разведки (scaffold/нав; тач-цели/a11y; токены/списки/refresh)
> + прямое чтение ядра и Figma-метаданных страниц 21/22/28.

Формат finding: `ID · файл:строка · P0/P1/P2 · что не так · как воспроизвести · чему противоречит`.

---

## 1. Навигация (ось 1) — P0

Контракт Figma: **каждый** production-фрейм (Pages 21–32) содержит
`Fixed Bottom Navigation Host` — 366×68, боковой инсет 12, y=768,
**вне** «scroll content» (768pt). Нав не скроллится никогда.

- **NAV-01** · `src/ui/screens/StudentHomeScreen.tsx:200` · P0 · `RoleBottomNav` — последний ребёнок контента внутри ручного `ScrollView` (:94). Repro: войти учеником → дом; таб-бар уезжает со скроллом и живёт ПОД кнопкой «Выйти». Противоречит STU-HOME-01 (316:2, host 316:34).
- **NAV-02** · `src/ui/screens/ScheduleScreen.tsx:266` · P0 · нав внутри `PremiumScrollScreen`. Repro: Расписание → пролистать вниз. Противоречит STU-SCHEDULE-01 (Page 24, fixed host).
- **NAV-03** · `src/ui/screens/LessonDetailScreen.tsx:139` · P0 · то же на детали урока. Противоречит Page 24 host-контракту.
- **NAV-04** · `src/navigation/tabs.ts:50,57,64` + `src/ui/screens/ScheduleScreen.tsx:148-158` · P0 · таб `schedule` есть у всех четырёх ролей, но экран гардит `roles.includes("Student")`: Teacher/Admin/Owner получают «Раздел недоступен» в `PremiumScrollScreen` **без таб-бара** — тупик из первого ряда навигации. Repro: войти педагогом → таб «Расписание».
- **NAV-05** · `src/ui/screens/StaffWorkspaceScreen.tsx:120` · P0 · цель табов Teacher `today` и Admin `people` — экран без нава (единственный выход — «Выйти» или жест назад). Repro: админ → таб «Люди».
- **NAV-06** · `src/ui/screens/AccessDelegationScreen.tsx:195` (и гард :97) · P0 · цель таба Owner `team` — экран без нава. Repro: владелец → таб «Команда».
- **NAV-07** · `src/ui/screens/account/shared.tsx:51` · P0 · таб `today` Teacher ведёт на `/(protected)` (StaffWorkspace), а `TeacherTodayScreen` (кокпит Page 26, роут `/teacher`) **недостижим ни одним табом** — единственная ссылка `StaffWorkspaceScreen.tsx:155`.
- **NAV-08** · P0 · экраны без нава, достижимые из навигации: `StaffLessonsScreen` (7 входящих ссылок), `CreateLessonScreen`, `CreateStudentScreen`, `OnboardingDetailScreen`, `TeacherChangeScreen` — все на `PremiumScrollScreen` (слота нет). Auth-флоу (Welcome/SignIn/Activation/Recovery/SessionRecovery) без нава — корректно.
- **NAV-09** · `src/ui/screens/account/shared.tsx:41` + гарды · P1 · подсветка active-таба врёт: (а) `AccountNav` default `profile` — у Admin/Owner таба `profile` нет → на 17 экранах account/activity/progress не подсвечен ни один таб (должен `more`); (б) гарды с `active="today"` у ролей без таба `today`: `OperationsScreen.tsx:34`, `OwnerOverviewScreen.tsx:43`, `TeacherStudentsScreen.tsx:56,92` (основной экран педагога подсвечивает «Сегодня» вместо «Ученики»!), `AssessmentDetail:69,129`, `AssessmentCompose:95,179`, `TeacherLesson:212,228`.
- **NAV-10** · `src/ui/screens/teacher/TeacherStudentsScreen.tsx:44-46`, `src/ui/screens/OwnerOverviewScreen.tsx:34-36` · P1 · сегмент читается из параметров только в `useState`-инициализаторе: повторный тап по табу `review`/`analytics` на уже открытом экране сегмент НЕ переключает; подсветка табов `review`/`analytics` не совпадает с активным сегментом.
- **NAV-11** · `src/ui/screens/account/shared.tsx:50-74` + копии в 3 экранах · P1 · переключение табов делает `router.push` — стек растёт неограниченно; поведение «таб = переключение контекста» (LinkedIn/YouTube) требует `replace`.
- **NAV-12** · `src/ui/screens/account/shared.tsx:18-32,48` · P2 · `BUILT_TABS` = все 13 ключей → `isTabEnabled` всегда true, стиль disabled недостижим; плюс три экранные копии onSelectTab с разными подмножествами (StudentHome отключает `community` — уже построенный раздел).
- **NAV-13** · `src/ui/roleNavigation.tsx:80` + `src/ui/patterns/accountPatterns.tsx:53` · P1 · высота нава 68 захардкожена дважды, токена нет.
- **NAV-14** · `src/ui/screens/ScheduleScreen.tsx:287` · P1 · `role="Student"` захардкожен в наве — прошедший гард сотрудник увидел бы студенческие табы.

## 2. Scaffold-инвентарь (ось 2)

Три несовместимых каркаса:

| Каркас | Экранов | Нав | Safe-area | PTR |
|---|---|---|---|---|
| `AccountScreenShell` (accountPatterns.tsx:43) | 40 | слот, фиксирован | да | **невозможен** (нет scrollProps) |
| `PremiumScrollScreen` (components.tsx:48) | 12 | **нет слота** | да | через scrollProps (2 экрана) |
| Ручной ScrollView | 2 (StudentHome, Welcome) | inline | вручную | нет |

- **SC-01** · `src/ui/components.tsx:65` · P0 · `minHeight: height` (полный вьюпорт) на контенте — короткие экраны растягиваются, in-scroll нав оказывается за нижней кромкой.
- **SC-02** · `src/ui/patterns/accountPatterns.tsx:43-51` · P1 · шелл не пробрасывает `scrollProps`/`refreshControl` → pull-to-refresh физически невозможен на 40 экранах.
- **SC-03** · `src/ui/screens/StudentHomeScreen.tsx:93-97` · P1 · ручная скролл-обвязка, дублирующая PremiumScrollScreen.
- **SC-04** · P2 · `SafeAreaView` нигде; инсеты вручную в 4 файлах тремя разными формулами (`insets.top+s6` / `insets.top+lg` / `insets.top+sm`).

## 3. Тач-цели (ось 3) — P0

48 задекларирован **трижды** (`accessibility/policy.ts:1` — 0 применений; `sizes.touchMin` — 1 применение; `metrics.minimumTarget` — 7 применений в legacy) и не применён к нарушителям:

- **TT-01** · `src/ui/patterns/communityPatterns.tsx:314` · P0 · «Пожаловаться» под комментарием: `minHeight: space.s6` = **24pt** — половина минимума; единственный вход в модерацию для участника.
- **TT-02** · `src/ui/patterns/practicePatterns.tsx:162` · P0 · `TaskRow` (checkbox) `minHeight: 32`; строки вплотную.
- **TT-03** · `src/ui/patterns/journalPatterns.tsx:363` · P0 · `AreaChip` `height: 34` фиксированный, без hitSlop — **16 точек вызова** по приложению.
- **TT-04** · `src/ui/patterns/eventPatterns.tsx:267` · P0 · `FilterChip` `height: 34` фиксированный; плюс статический style-массив — нет press-фидбэка.
- **TT-05** · `OwnerOverviewScreen.tsx:219` / `teacher/TeacherStudentsScreen.tsx:206` / `community/CommunityScreen.tsx:271` · P0 · три копии сегментов `minHeight: 38`.
- **TT-06** · `src/ui/components.tsx:560` · P0 · `PrimaryButton` `height: 48` фиксированный (не растёт с Dynamic Type; ломается с ~×2.2); `:534` `revealAction` `height: 48` + `top: -9`.
- **TT-07** · `src/ui/patterns/dateChip.tsx:95` · P1 · `DateChip` `height: 64` фикс. при контенте 54 — ломается уже при ×1.2; `ScheduleScreen.tsx:297` `gap: 3` между целями.
- **TT-08** · `src/ui/patterns/accountPatterns.tsx:502,512` · P2 · BlockAction 52 литералом; `:475` iconTile 44 — случайный пол высоты SettingsRow.
- **TT-09** · Dynamic Type: фиксированные height с текстом — `lessonComponents.tsx:225` datePill 48/42w, `:254` selectionMark 22, `CreateStudentScreen.tsx:373` checkbox 24 (ломается ×1.15), аватары с инициалами 32/48/72. Нигде нет `maxFontSizeMultiplier` — рост неограничен, контейнеры фиксированы.

## 4. Токены (ось 4) — P1

- **TK-01** · `src/ui/components.tsx:34` · P1 · shared-модуль legacy-only (7/7 легаси-групп); `InlineNotice` (43 потребителя), `PremiumTextField` (19+), `uiStyles` (10) протаскивают legacy-цвета во **все** современные экраны.
- **TK-02** · `src/ui/lessonComponents.tsx:12` · P1 · legacy-only; `DateStrip` — мёртвый код (0 потребителей).
- **TK-03** · 13 legacy-only экранов (AccessDelegation, CreateLesson, CreateStudent, LessonDetail, OnboardingDetail, Schedule, SessionRecovery, SignIn, StaffLessons, StaffWorkspace, StudentHome, TeacherChange, Welcome) + `ActivationScreen.tsx:33` смешанный (`spacing`+`space`, `typeScale`+`typeStyles` в одном файле).
- **TK-04** · `src/ui/patterns/accountPatterns.tsx:16` · P2 · единственный pattern с легаси-хвостом (`gradients`).
- **TK-05** · P2 · мёртвые modern-токены: `motion`, `elevation`, `semanticByMode.light`, `palette` — вне tokens.ts используются только контрактным тестом.
- Шкалы: legacy `spacing.field=18`, `spacing.section=30` не имеют modern-эквивалента; Figma-фреймы используют gutter 16 (s4) и секционные интервалы 20–24 (s5/s6) → легаси-значения умирают при миграции, новые значения НЕ нужны (корпус не затрагивается).

## 5. Плотность и IA (ось 5) — P1

- **IA-01** · `src/ui/screens/community/PostDetailScreen.tsx` · P1 · один скролл сливает три Figma-экрана: тред COM-POST-01 (347:151), жалобу COM-SAFE-02 (348:683 — отдельный экран с заголовком «Пожаловаться · Что произошло?»), блокировку COM-SAFE-03 (348:757 — отдельный экран «Безопасность»); 3 `ScreenHeading` + 7 `BlockAction` на одной странице. Repro: открыть пост → «Пожаловаться» → форма раскрывается посреди треда.
- **IA-02** · `src/ui/screens/journal/StudentProgressScreen.tsx:222-228` + `GrowthSections.tsx` (584 строки) · P1 · один скролл сливает раскадровку из 12 экранов Page 22 (STU-GROWTH-01..12): обзор + цель (04) + достижения (08) + инлайн-формы преподавателя (создать/завершить/переосмыслить цель, наградить/отозвать, каталог) = ~24 интерактива и ~10 полей.
- **IA-03** · `src/ui/screens/practice/HomeworkDetailScreen.tsx` (501 строка) · P1 · путь ученика (STU-PRACTICE-02..09) + ревью педагога + отмена в одном скролле; ~15 действий, 6 полей.
- **IA-04** · `src/ui/screens/community/CommunityScreen.tsx:175-181` · P2 · «таб» События — на деле кнопка навигации в каталог, замаскированная под таб.

## 6. Списки и перфоманс (ось 6) — P1

- **LST-01** · P1 · `FlatList`/`SectionList` — **0** во всём приложении; 89 `.map()` в JSX. Кандидаты на виртуализацию (переменная длина): ActivityScreen:256,260 (готовые секции today/earlier), CommunityScreen:228, EventsCatalogScreen:112, ModerationScreen:109, RepertoireScreen:271 (O(n·m): вложенные map), TeacherStudentsScreen:140,170 (ростер дважды), StaffLessonsScreen:150,183 (окно −14/+90 дней), PracticeHubScreen:120, StudentProgressScreen:169,193, SecurityActivityScreen:117, StaffWorkspaceScreen:243, PostDetailScreen:211.
- **LST-02** · P1 · pull-to-refresh: 2 экрана из 54 (StaffLessons, StaffWorkspace — оба через escape-hatch PremiumScrollScreen).
- **LST-03** · P0 · **retry потерян на всех 35 modern-экранах**: `InlineNotice` со статичным заголовком «Повторить» — не кнопка (`components.tsx:390-412` — View без onPress); ни один `.reload()` не привязан к ошибке. В 11 legacy-экранах retry есть (SecondaryButton «Повторить»). Repro: выключить API → открыть Активность → «Повторить» не нажимается.
- **LST-04** · P1 · loading-состояние почти нигде в pattern A (`.loading` читается 1 раз на 35 экранов); `PracticeHubScreen.tsx:58` конфлатирует `value===null` (loading) с ошибкой; хуки стреляют запросами до гардов (`PracticeHubScreen.tsx:41` vs `:45`, `OperationsScreen.tsx:30` vs `:32`).
- **LST-05** · `community/CommunityScreen.tsx:191-197` · P1 · ошибка загрузки policies не показывается — гейт правил сообщества молча пропускается при транспортной ошибке (COM-SAFE-04 рассчитан на «политики нет», а не «не смогли узнать»).

## 7. Доступность (ось 7) — P1/P2

- **A11Y-01** · три сегмент-контрола (см. TT-05) · P1 · контейнер без `tablist`, дети `button` без `accessibilityState` — выбранность только цветом. Единственный правильный tablist — RoleBottomNav.
- **A11Y-02** · `community/CommunityScreen.tsx:210-216` · P1 · тоггл «Только объявления» — role=button без state (должен switch + checked).
- **A11Y-03** · `events/RescheduleRequestScreen.tsx:180,188,205` · P1 · StatusRow используется как радио-выбор без checked/radiogroup.
- **A11Y-04** · `eventPatterns.tsx:90 vs :99` · P2 · RsvpControl: busy объявлен только в неинтерактивной ветке.
- **A11Y-05** · `communityPatterns.tsx:147` · P2 · пустой accessibilityLabel при отсутствии title/body.
- **A11Y-06** · `lessonComponents.tsx:145` · P2 · SelectableRow: тернарник-пустышка в state.
- **A11Y-07** · `CreateStudentScreen.tsx:233` · P2 · выбор педагога без radiogroup (AccessDelegation:208 — с ним).
- **A11Y-08** · `accountPatterns.tsx:304` (AccountBanner role=text без accessible), `lessonComponents.tsx:20` (декоративный аватар в фокус-порядке) · P2.

## 8. Консистентность (ось 8) — P2

- **CON-01** · три системы заголовка экрана: `ScreenHeading` (semantic) / ручные brand+eyebrow+title (legacy) / `uiStyles.screenTitle`.
- **CON-02** · три стиля empty/error: PremiumCard-текст / StatusCard / голый Text.
- **CON-03** · текстовый chevron «›» (lessonComponents:72, StaffWorkspace:272) vs `ChevronIcon`.
- **CON-04** · `TextAction` ширина < 48pt при коротких метках (paddingHorizontal 8).

---

# Target — целевая архитектура (Фаза B)

Направление владельца (§5 промпта) — уточнённые решения:

## T1. Единый scaffold `Screen` — `src/ui/screen.tsx`

Эволюция `AccountScreenShell`; геометрия нав-хоста из Figma 310:20542
(host 366×68, боковой инсет 12, нижний зазор 8 = 844−768−68):

```
navigation (tokens.ts, проекция Figma-компонента Bottom Navigation · Production):
  height 68 · sideInset 12 (=space.s3) · bottomGap 8 (=space.s2) · maxWidth 366
```

API: `children · navigation? · refreshControl? · scrollProps? · keyboardAware? · gutter=space.s4 · contentGap=space.s3 · testID`.
Правила: нав всегда absolute-хост у нижней кромки (`bottom: insets.bottom + bottomGap`), контент резервирует `insets.bottom + height + bottomGap*2 + s5`; `paddingTop: insets.top + s6`; НИКАКОГО `minHeight: height`. `ScreenList` — та же рама с `FlatList` вместо ScrollView (header/empty/footer — render-слоты) для лент.
`AccountScreenShell` → делегат Screen (совместимость, затем переименование по экранам); `PremiumScrollScreen` → делегат Screen (`contentGap: 0`, свой gutter) до по-экранной миграции, затем удаление. Ручные обвязки (StudentHome, Welcome) мигрируют на Screen.

## T2. Единая роль-навигация `RoleNav` (эволюция AccountNav, shared.tsx)

Один источник маршрутов таб→роут для всех экранов (копии onSelectTab удаляются):
- Student: today→`/(protected)` · schedule→`/schedule` · practice→`/practice` · community→`/community` · profile→`/account`;
- Teacher: today→`/teacher` (кокпит, чинит NAV-07) · schedule→`/lessons` (чинит NAV-04) · students→`/teacher/students` · review→`…?segment=review` · community→`/community`;
- Administrator: operations→`/operations` · schedule→`/lessons` · people→`/(protected)?workspace=staff` · community→`/community` · more→`/account`;
- Owner: overview→`/overview` · analytics→`…?segment=analytics` · operations→`/operations` · team→`/access` · more→`/account`.
Переключение — `router.replace` (NAV-11). Active по умолчанию: `profile`→фолбэк `more` при отсутствии (NAV-09). Сегментные экраны читают сегмент из URL-параметров (`router.setParams`), а не из useState (NAV-10). Все без-нав экраны из NAV-05/06/08 получают `navigation={<RoleNav active=…/>}`.

## T3. Один Button API — `src/ui/button.tsx` (semantic-токены)

`Button`: kind `primary|secondary|text` × shape `block` (52, radius.lg — Figma «Action · …») | `pill` (48, radius.pill — auth-флоу); всегда `minHeight`; busy/disabled/tone danger; ширина text-кнопок ≥ touchMin. `BlockAction`/`PrimaryButton`/`SecondaryButton`/`TextAction` становятся тонкими делегатами (API сохраняется, реализация одна), затем имена сходятся к Button по мере миграции экранов.

## T4. Один `SegmentedControl` — `src/ui/segmentedControl.tsx`

`tablist`/`tab` роли + `selected` state; визуальная высота 38 (Figma Community tabs), hit-area 48 через hitSlop 5/5; заменяет три копии. «Таб-кнопка» (Events в Community) объявляется `button`, не tab.

## T5. Чипы и цели

Общий `Chip` (semantic): визуал 34 (Figma), `minHeight` + hitSlop 7/7 → 48; `AreaChip`/`FilterChip` — делегаты. TaskRow → minHeight touchMin. reportAction → minHeight touchMin. DateChip → minHeight 64. PrimaryButton/revealAction → minHeight. Единственный источник минимума — `sizes.touchMin` (policy.ts реэкспортирует его).

## T6. Токены — одна система

Порядок: (1) components.tsx + lessonComponents.tsx → semantic (разблокирует всё); (2) legacy-экраны по контурам (срезы 2–3); (3) удалить legacy-слой из tokens.ts + gradients-хвост accountPatterns (градиенты — в semantic-производные или устраняются по Figma). Соответствия: canvas→bgCanvas, surface→bgSurface(ink850)/bgRaised, raised→bgRaised, border→borderDefault, borderGlass→borderGlass, textPrimary→textPrimary, textSecondary→textSecondary, textMuted→textMuted, textAccent→textAccent, textGold→textGold, violet→bgAction/accentViolet, violetPressed→bgActionPressed, cyan→accentCyan, gold→accentGold, danger→feedbackDanger, success→feedbackSuccess(зелёный по Figma, НЕ cyan — фикс легаси-бага), spacing.field/section→s4/s5/s6/s8, radii→radius, typeScale→typeStyles (по факт. размерам), fonts→typeStyles.fontFamily. Новые значения НЕ вводятся.

## T7. Данные: retry/refresh обязательны

`useAccountResource` дополняется `refreshing` (reload при value≠null); стандарт экрана: ошибка = `ErrorNotice` (InlineNotice + кнопка «Повторить» → reload) — чинит LST-03; loading-плейсхолдер обязателен; PTR через `refreshControl` на каждом data-экране; гарды до вызова хука.

## T8. IA-декомпозиция по Figma-раскадровке

- `/community/[postId]` = только тред (COM-POST-01); `/community/report` (params targetType/targetId) = COM-SAFE-02; `/community/safety` (params accountId) = COM-SAFE-03; PostDetail ссылается на них.
- `/progress` = обзор (STU-GROWTH-01: signal + области + evidence + ссылки); `/progress/goal` = цель + формы преподавателя (04); `/progress/achievements` = награды + формы (08); assessments-секция остаётся списком со ссылками на существующий `/assessment/[id]`.
- `/practice/[homeworkId]` = статус/план/фидбэк; `/practice/[homeworkId]/submit` = форма сдачи + загрузки; ревью педагога → `/practice/[homeworkId]/review`.

## T9. FlatList

`ScreenList` для: активность (SectionList-семантика через секции header), лента сообщества, каталог событий, модерация, репертуар, ростер учеников, staff-занятия. Малые фиксированные наборы (чипы/сегменты/enum) остаются map.

---

# План исполнения

- **Срез 1 — Фундамент**: Screen + navigation-токены + RoleNav-маршруты + все NAV-01..08 + Button/Segmented/Chip + 48pt (TT-01..06) + миграция components/lessonComponents на semantic + ErrorNotice/refreshing. Гейты → коммит → ledger → push.
- **Срез 2 — Контур ученика**: Вход→Дом→Расписание→Урок→Практика→Прогресс (декомпозиция IA-02)→Сообщество (IA-01)→Активность→Профиль; токены, FlatList, PTR, сверка с Figma по фреймам.
- **Срез 3 — Контуры персонала**: Teacher* (NAV-09b подсветки), Operations, OwnerOverview, Series, StaffLessons/Workspace, Create*/TeacherChange/Onboarding.
- Каждый срез: mobile:check зелёный, тестов ≥214 и только рост, ledger-запись с закрытыми finding-ID.

## Статус findings (обновляется по срезам)

| Срез | Закрыто | Примечание |
|---|---|---|
| ux1 (1973ad3) | NAV-01..14, SC-01..03, TT-01..09, TK-01..02, A11Y-01/02 (частично), LST-03/05 (частично) | ErrorNotice/PTR-разводка по экранам, FlatList и IA-декомпозиция — в срезах ux2/ux3; SC-04 умирает вместе с миграцией экранов на Screen |
| ux2 | IA-01 (тред/жалоба/блокировка — три экрана: /community/[postId] + /community/report + /community/safety), IA-02 (/progress → обзор + /progress/goal + /progress/achievements), IA-03 (/practice/[homeworkId] → статус + /submit + /review), LST-01 (частично: ScreenList на активности и ленте сообщества), LST-02/03/04 (контур ученика: PTR+retry+guard-before-hook на Practice/Repertoire/Progress/PostDetail/Community/Activity/Schedule/Home; account-зона и события — механическая разводка), TK-03 (частично: Welcome/SignIn/SessionRecovery/Activation/StudentHome/Schedule/LessonDetail на semantic) | Реестр непроверенного: пофреймовая визуальная сверка (242 состояния) остаётся device/human-шагом из p8; Home/auth несут B.0-композицию до контентных решений Page 21/32; FlatList для Repertoire/EventsCatalog отложен (объёмы малы), для staff-лент — срез ux3 |
