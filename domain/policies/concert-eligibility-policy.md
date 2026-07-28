---
Status: Approved
Version: 1.0.0
Last Updated: 2026-07-27

Policy Id: CONCERT_ELIGIBILITY_POLICY

Policy Type:
  - Eligibility Policy
  - Validation Policy
  - Reaction Policy
  - Scheduling Policy
  - Recommendation Policy
  - Escalation Policy

Owners:
  - Product Owner
  - Education Lead
  - Concert Coordinator
  - Technical Lead

Related Aggregates:
  - Concert
  - ConcertParticipation
  - Performance
  - PerformanceSlot
  - Student
  - Teacher
  - Song
  - Repertoire
  - Progress
  - Assessment
  - Rehearsal
  - Consent
  - MediaAsset

Observed Events:
  - ConcertCreated
  - ConcertPublished
  - ConcertRequirementsChanged
  - ConcertRegistrationOpened
  - ConcertRegistrationClosed
  - ConcertParticipationProposed
  - ConcertParticipationRequested
  - ConcertParticipationWithdrawn
  - SongMarkedPerformanceReady
  - SongReadinessChanged
  - SongVersionChanged
  - TeacherConcertApprovalGranted
  - TeacherConcertApprovalWithdrawn
  - RehearsalCompleted
  - RehearsalAssessmentPublished
  - PerformanceSlotAssigned
  - PerformanceSlotChanged
  - TechnicalAssetUploaded
  - TechnicalAssetApproved
  - StudentConsentGranted
  - StudentConsentWithdrawn
  - GuardianConsentGranted
  - GuardianConsentWithdrawn
  - StudentAvailabilityChanged
  - ConcertEligibilityReviewRequested
  - ConcertProgramPublished
  - ConcertCancelled

Produced Commands:
  - CreateConcertParticipation
  - EvaluateConcertEligibility
  - MarkConcertParticipationEligible
  - MarkConcertParticipationConditionallyEligible
  - RejectConcertParticipation
  - RequestTeacherConcertApproval
  - RequestStudentConsent
  - RequestGuardianConsent
  - RequestConcertRehearsal
  - RequestTechnicalAsset
  - RequestTechnicalAssetReview
  - RequestConcertEligibilityReview
  - AssignPerformanceSlot
  - RemovePerformanceSlot
  - SuspendConcertEligibility
  - RestoreConcertEligibility
  - WithdrawConcertParticipation
  - PublishConcertProgramEntry
  - RemoveConcertProgramEntry
  - NotifyConcertCoordinator

Related Documents:
  - 000-domain-policy-overview.md
  - song-readiness-policy.md
  - progress-update-policy.md
  - goal-completion-policy.md
  - achievement-award-policy.md
  - lesson-completion-policy.md
  - ../student.md
  - ../assessment.md
  - ../progress.md
  - ../lesson.md
---

# Concert Eligibility Policy

> Concert Eligibility Policy определяет, может ли конкретный Student исполнить конкретную Song в конкретной версии и формате на конкретном Concert.
>
> Политика не оценивает общую ценность ученика и не заменяет педагогическое решение Teacher. Она объединяет педагогическую готовность, согласие, организационные ограничения, техническую подготовку и требования мероприятия.

---

# Purpose

Даже если Song получила статус `Performance Ready`, этого недостаточно для автоматического включения в программу Concert.

Необходимо учитывать:

- требования конкретного Concert;
- формат выступления;
- актуальность Song Readiness;
- выбранную версию Song;
- подтверждение Teacher;
- согласие Student;
- согласие законного представителя, если требуется;
- наличие репетиции;
- длительность номера;
- технические ресурсы;
- ограничения программы;
- доступность Student;
- готовность других участников дуэта или ансамбля;
- отсутствие блокирующих проблем;
- изменения, произошедшие после первоначального допуска.

Политика создает единый, объяснимый и аудируемый процесс допуска.

---

# Core Principle

Допуск относится не к Student или Song отдельно, а к конкретной заявке на выступление.

```text
Student
  +
Song Version
  +
Performance Format
  +
Concert
  +
Current Evidence
  =
Concert Participation Eligibility
```

Пример:

```text
Student: 1482
Song: "Feeling Good"
Song Version: concert-edit-v3
Performance Format: Solo
Concert: Summer Showcase 2026
```

Допуск для этой комбинации не переносится автоматически:

- на другой Concert;
- на другую версию Song;
- на дуэт;
- на другую тональность;
- на выступление после длительного перерыва;
- на другой состав участников.

---

## Eligibility Is Not Participation

Необходимо различать:

- Eligibility;
- Approval;
- Program Placement;
- Participation;
- Performance Completion.

```text
Proposed
   |
   v
Under Review
   |
   v
Eligible
   |
   v
Approved
   |
   v
Scheduled
   |
   v
Confirmed
   |
   v
Performed
```

Student может быть Eligible, но не попасть в программу из-за:

- ограниченного количества слотов;
- общей длительности Concert;
- баланса программы;
- технических ограничений;
- отмены Student;
- решения организатора;
- изменения формата мероприятия.

---

# Concert Participation

Рекомендуемая структура:

```text
ConcertParticipation
├── ParticipationId
├── ConcertId
├── StudentId
├── PerformanceType
├── SongId
├── SongVersionId
├── RepertoireEntryId
├── ParticipantIds
├── ResponsibleTeacherId
├── EligibilityStatus
├── ProgramStatus
├── PerformanceSlotId
├── Duration
├── ReadinessReference
├── ApprovalReferences
├── ConsentReferences
├── RehearsalReferences
├── TechnicalRequirementReferences
├── BlockingIssues
├── Conditions
├── EligibilityDecision
├── PolicyId
├── PolicyVersion
└── Version
```

---

# Performance Types

Допустимые базовые типы:

- Solo;
- Duet;
- Ensemble;
- Group Performance;
- Teacher and Student;
- Instrumental Collaboration;
- Spoken Performance;
- Experimental Format.

Каждый Concert может ограничивать доступные типы.

---

# Eligibility Status

## Proposed

Участие предложено, но проверка еще не начата.

## Under Review

Политика или уполномоченный Actor оценивают заявку.

## Insufficient Data

Для решения отсутствуют обязательные сведения.

## Conditionally Eligible

Основные требования выполнены, но остаются условия.

Примеры:

- пройти обязательную репетицию;
- загрузить утвержденный минус;
- получить согласие;
- подтвердить сокращенную версию;
- устранить технический блокер.

## Eligible

Все обязательные требования допуска выполнены.

Это еще не означает включение в программу.

## Not Eligible

Есть подтвержденное основание, препятствующее участию в текущем виде.

Решение должно содержать причину и следующий возможный шаг.

## Suspended

Ранее выданный допуск временно приостановлен из-за изменения условий.

## Withdrawn

Student, Teacher или уполномоченный сотрудник отозвали участие.

## Expired

Допуск более не считается актуальным.

Например:

- Concert перенесен на значительно более позднюю дату;
- Song Readiness устарела;
- изменилась версия Song;
- изменился состав участников.

## Cancelled

Participation больше не рассматривается из-за отмены Concert или административного решения.

---

# Program Status

Eligibility Status и Program Status хранятся отдельно.

Допустимые Program Status:

- Not Considered;
- Candidate;
- Waitlisted;
- Selected;
- Scheduled;
- Confirmed;
- Removed;
- Performed;
- No Show.

---

# Decision Outcomes

Политика может возвращать:

## Eligible

Все требования допуска выполнены.

## Conditionally Eligible

Допуск возможен после выполнения конкретных условий.

## Not Eligible

В текущем состоянии участие невозможно.

## Insufficient Data

Недостаточно данных для решения.

## Teacher Review Required

Нужно педагогическое решение.

## Concert Coordinator Review Required

Нужно организационное решение.

## Technical Review Required

Необходимо проверить технический материал или формат.

## Consent Required

Отсутствует обязательное согласие.

## Rehearsal Required

Не выполнено требование по репетиции.

## Context Reassessment Required

Текущая Song Readiness не соответствует контексту Concert.

## Suspended

Ранее выданный допуск временно приостановлен.

## Already Eligible

Повторная оценка не изменила решение.

## Rejected

Запрос некорректен, неавторизован или нарушает доменные ограничения.

---

# Eligibility Dimensions

Concert Eligibility состоит из нескольких независимых измерений.

## Pedagogical Eligibility

Готов ли Student исполнить произведение на необходимом уровне в контексте Concert.

## Consent Eligibility

Согласен ли Student участвовать.

## Organizational Eligibility

Соответствует ли заявка формату и ограничениям мероприятия.

## Technical Eligibility

Доступны ли необходимые материалы и оборудование.

## Participant Eligibility

Готовы ли все участники дуэта, ансамбля или группы.

## Schedule Eligibility

Может ли Student присутствовать в необходимое время.

## Safety Eligibility

Отсутствуют ли известные блокирующие обстоятельства, делающие участие небезопасным или неразумным.

Политика не ставит медицинские диагнозы.

---

# Preconditions

Перед оценкой должны существовать:

- Concert.
- Concert Participation.
- Student.
- Responsible Teacher или утвержденный педагогический ответственный.
- Song, если выступление музыкальное.
- Идентифицируемая Song Version.
- Performance Type.
- Требования Concert.
- Дата и формат мероприятия.
- Policy Version.
- Actor или Trigger с правом инициировать проверку.

---

# Concert Requirements

Concert должен содержать версионируемый набор требований.

```text
ConcertRequirements
├── ConcertId
├── Version
├── AllowedPerformanceTypes
├── MinimumReadinessStage
├── ReadinessValidityPeriod
├── TeacherApprovalRequired
├── RehearsalRequired
├── MinimumRehearsalCount
├── ConsentRequired
├── GuardianConsentRequired
├── MaximumPerformanceDuration
├── TechnicalSubmissionDeadline
├── RegistrationDeadline
├── ParticipantLimit
├── SongRestrictions
├── ContentRestrictions
├── EquipmentConstraints
├── AgeRestrictions
├── AccessibilityConstraints
├── ProgramRules
└── ValidFrom
```

Изменение Concert Requirements должно запускать повторную проверку затронутых Participation.

---

# Required Evidence

В зависимости от Concert могут использоваться:

- Song Readiness State;
- Teacher Concert Approval;
- Rehearsal Assessment;
- full run-through confirmation;
- Student Consent;
- Guardian Consent;
- technical asset approval;
- availability confirmation;
- ensemble participant approvals;
- current Song Version;
- performance duration;
- Concert Requirements Version;
- blocking issue resolution;
- Student Self Confirmation.

---

# Decision Rules

## CE-001: Eligibility is concert-specific

Допуск к одному Concert не переносится на другой.

Reason Code: `CONCERT_SPECIFIC_ELIGIBILITY_REQUIRED`

---

## CE-002: Participation record is required

Решение должно относиться к конкретной Concert Participation.

Reason Code: `CONCERT_PARTICIPATION_REQUIRED`

---

## CE-003: Concert must be active

Нельзя выдавать новый допуск для Concert в состоянии:

- Cancelled;
- Completed;
- Archived;
- Registration Closed, если исключения не разрешены.

Reason Code: `CONCERT_NOT_ACCEPTING_PARTICIPATION`

---

## CE-004: Concert requirements must be versioned

При оценке сохраняется версия требований.

Reason Code: `CONCERT_REQUIREMENTS_VERSION_REQUIRED`

---

## CE-005: Performance type must be allowed

Performance Type должен поддерживаться Concert.

Reason Code: `PERFORMANCE_TYPE_NOT_ALLOWED`

---

## CE-006: Song version must be explicit

Для музыкального выступления должна быть определена конкретная версия Song.

Недостаточно указать только название произведения.

Reason Code: `CONCERT_SONG_VERSION_REQUIRED`

---

## CE-007: Performance readiness is required when configured

Если Concert требует Performance Ready, связанная Song Readiness должна:

- относиться к тому же Student;
- относиться к той же Song Version;
- учитывать релевантный performance context;
- быть подтвержденной Teacher;
- быть актуальной.

Reason Code: `PERFORMANCE_READINESS_REQUIRED`

---

## CE-008: Performance Ready does not automatically grant eligibility

Даже подтвержденная Song Readiness является только одним из входов.

Reason Code: `SONG_READY_NOT_CONCERT_ELIGIBLE`

---

## CE-009: Readiness context must match the concert

Готовность для Regular Lesson недостаточна, если Concert требует публичного сценического контекста.

Reason Code: `READINESS_CONTEXT_MISMATCH`

---

## CE-010: Readiness must be recent enough

Concert Requirements могут задавать период актуальности.

Например:

`ReadinessValidityPeriod: 30 days`

Если подтверждение старше допустимого периода:

Decision: Context Reassessment Required

Reason Code: `SONG_READINESS_EXPIRED`

---

## CE-011: Material changes invalidate previous readiness linkage

Повторная оценка требуется при изменении:

- тональности;
- аранжировки;
- продолжительности;
- структуры;
- состава участников;
- минусовки;
- формата исполнения;
- сценической постановки.

Reason Code: `PERFORMANCE_VERSION_CHANGED`

---

## CE-012: Teacher approval may be mandatory

Если Concert Requirements требуют Teacher approval, без него Eligibility не выдается.

Decision: Teacher Review Required

Reason Code: `TEACHER_CONCERT_APPROVAL_REQUIRED`

---

## CE-013: Teacher approval must be scoped

Подтверждение должно относиться к:

- конкретному Student;
- конкретной Song Version;
- конкретному Concert;
- конкретному Performance Type.

Общее утверждение «песня готова» недостаточно.

---

## CE-014: Teacher approval may expire

Approval может потерять актуальность после:

- значительного переноса Concert;
- изменения Song Version;
- смены состава;
- выявления новых блокеров;
- отзыва Assessment;
- длительного перерыва.

---

## CE-015: AI cannot grant pedagogical eligibility

AI может подготовить предложение или обнаружить отсутствующие требования.

AI не может:

- подтвердить Teacher approval;
- присвоить Eligible;
- отменить решение Teacher;
- признать Student неготовым окончательно.

Reason Code: `AI_CANNOT_APPROVE_CONCERT_ELIGIBILITY`

---

## CE-016: Student consent is required

Student не должен включаться в Concert без согласия, если только отдельная утвержденная модель не описывает иной процесс предварительного предложения.

Reason Code: `STUDENT_CONSENT_REQUIRED`

---

## CE-017: Consent must be informed

Student должен понимать:

- дату;
- формат;
- площадку;
- предполагаемую Song;
- публичность;
- возможную запись;
- примерное время участия;
- основные требования.

Согласие без достаточной информации может потребовать повторного подтверждения.

---

## CE-018: Consent can be withdrawn

Student может отозвать согласие согласно правилам Concert.

Отзыв не должен:

- ухудшать Progress;
- аннулировать Achievement;
- создавать наказание;
- автоматически снижать оценку Teacher.

Reason Code: `STUDENT_CONSENT_WITHDRAWN`

---

## CE-019: Guardian consent may be required

Для несовершеннолетнего или другого регулируемого случая необходимо согласие законного представителя.

Reason Code: `GUARDIAN_CONSENT_REQUIRED`

---

## CE-020: Media consent is separate

Согласие на участие не означает согласие на:

- фото;
- видео;
- прямую трансляцию;
- публикацию;
- маркетинговое использование.

Reason Code: `MEDIA_CONSENT_NOT_GRANTED`

Отсутствие Media Consent не всегда блокирует само выступление. Это зависит от формата Concert.

---

## CE-021: Rehearsal may be mandatory

Если Concert требует обязательную репетицию, Eligibility может быть только условной до ее завершения.

Reason Code: `CONCERT_REHEARSAL_REQUIRED`

---

## CE-022: Rehearsal must match the performance format

Сольная репетиция не подтверждает готовность дуэта.

Репетиция без микрофона может быть недостаточной, если Concert требует проверку сценического оборудования.

Reason Code: `REHEARSAL_CONTEXT_MISMATCH`

---

## CE-023: Rehearsal completion alone is not enough

Сам факт присутствия на репетиции не подтверждает успешную готовность.

При необходимости требуется Rehearsal Assessment.

---

## CE-024: Failed rehearsal triggers review, not automatic humiliation

При существенных проблемах результатом может быть:

- дополнительная репетиция;
- изменение Song Version;
- сокращение номера;
- смена формата;
- перенос участия;
- добровольный отказ;
- Not Eligible.

Объяснение должно быть уважительным.

---

## CE-025: Performance duration must fit concert constraints

Продолжительность должна быть известна и не превышать установленный лимит.

Reason Code: `PERFORMANCE_DURATION_EXCEEDED`

---

## CE-026: Duration must include applicable transitions

В зависимости от формата учитываются:

- музыкальное вступление;
- выход;
- техническая пауза;
- представление;
- смена участников;
- завершение.

Конкретная модель задается Concert Requirements.

---

## CE-027: Technical assets may be mandatory

Могут потребоваться:

- backing track;
- sheet music;
- lyrics;
- click track;
- stems;
- instrument requirements;
- microphone requirements;
- playback notes;
- lighting cues.

Reason Code: `TECHNICAL_ASSET_REQUIRED`

---

## CE-028: Technical asset must be approved

Загрузка файла не означает его готовность.

Проверяются:

- формат;
- версия;
- длительность;
- качество;
- отсутствие повреждения;
- соответствие заявке;
- правильное начало и окончание;
- понятное имя;
- доступность для площадки.

Reason Code: `TECHNICAL_ASSET_NOT_APPROVED`

---

## CE-029: Technical changes require reassessment

После утверждения минуса его замена должна:

- создать новую версию;
- отменить старую активную ссылку;
- инициировать техническую проверку;
- при существенных изменениях инициировать Teacher Review.

---

## CE-030: Student availability must be confirmed

Student должен быть доступен:

- в дату Concert;
- в необходимое время;
- на обязательной репетиции;
- в требуемом месте;
- в рамках допустимых ограничений.

Reason Code: `STUDENT_AVAILABILITY_NOT_CONFIRMED`

---

## CE-031: Schedule conflicts must be explicit

Если участие конфликтует с другим обязательством, политика не должна молча выбрать одно.

Decision: Concert Coordinator Review Required

Reason Code: `STUDENT_SCHEDULE_CONFLICT`

---

## CE-032: Eligibility does not guarantee a slot

Даже Eligible Participation может:

- попасть в Waitlist;
- не войти в итоговую программу;
- быть заменена другим номером;
- перейти на другой Concert.

Reason Code: `ELIGIBLE_WITHOUT_PROGRAM_SLOT`

---

## CE-033: Performance slot assignment requires eligibility or approved exception

Обычно Slot назначается только Eligible или Conditionally Eligible Participation.

Исключение должно быть:

- авторизовано;
- объяснено;
- аудировано;
- ограничено по времени.

---

## CE-034: Slot changes may require revalidation

Изменение времени может повлиять на:

- доступность Student;
- готовность участников;
- технические переходы;
- возрастные ограничения;
- транспорт;
- продолжительность программы.

---

## CE-035: Duet eligibility requires all participants

Для дуэта должны быть проверены:

- оба Student;
- общая Song Version;
- роли и партии;
- совместная репетиция;
- согласие каждого;
- доступность каждого;
- Teacher approval;
- техническая готовность.

Reason Code: `DUET_PARTICIPANT_NOT_ELIGIBLE`

---

## CE-036: Ensemble eligibility requires composition identity

Должен быть определен актуальный состав.

Изменение участника может требовать:

- новой репетиции;
- изменения партий;
- повторного approval;
- новой оценки продолжительности;
- технического пересмотра.

---

## CE-037: Individual readiness does not prove ensemble readiness

Даже если каждый участник готов отдельно, совместная готовность должна быть подтверждена.

Reason Code: `ENSEMBLE_READINESS_NOT_CONFIRMED`

---

## CE-038: Group participation cannot hide individual consent

Для каждого участника должно быть подтверждено требуемое согласие.

---

## CE-039: Content restrictions must be respected

Concert может иметь требования к:

- тексту;
- возрастной уместности;
- ненормативной лексике;
- тематике;
- длительности;
- языку;
- формату мероприятия.

Reason Code: `PERFORMANCE_CONTENT_NOT_ALLOWED`

---

## CE-040: Content review must not become arbitrary censorship

Ограничение должно ссылаться на:

- опубликованное правило Concert;
- юридическое требование;
- возрастной формат;
- согласованный бренд-контекст;
- безопасность мероприятия.

Решение должно быть объяснимым.

---

## CE-041: Student health information must be handled carefully

Политика не ставит диагноз и не требует раскрытия подробной медицинской информации.

Student или Teacher могут указать:

`ParticipationAvailability: Temporarily Unavailable`

без раскрытия диагноза.

---

## CE-042: Known immediate safety concern may suspend eligibility

Если уполномоченный Actor указывает на непосредственный риск, Eligibility может быть временно приостановлена до Review.

Reason Code: `PARTICIPATION_SAFETY_REVIEW_REQUIRED`

Это не является медицинским заключением.

---

## CE-043: Absence from critical rehearsal may suspend eligibility

Если обязательная репетиция пропущена и исключение не подтверждено:

Decision: Suspended

Reason Code: `MANDATORY_REHEARSAL_MISSED`

---

## CE-044: Withdrawal is not failure

Статус Withdrawn не должен интерпретироваться как:

- снижение Skill;
- потеря Progress;
- нарушение дисциплины;
- негативный Achievement.

---

## CE-045: Organizer removal requires reason

Удаление Eligible Participation из программы должно иметь организационную причину.

Например:

- программа переполнена;
- Concert сокращен;
- техническая невозможность;
- изменение формата;
- отмена блока программы.

---

## CE-046: Removal from program does not revoke pedagogical readiness

Song может остаться Performance Ready, даже если номер исключен из Concert.

---

## CE-047: Requirement changes trigger reevaluation

При изменении Concert Requirements все затронутые Participation должны быть повторно оценены.

Reason Code: `CONCERT_REQUIREMENTS_CHANGED`

---

## CE-048: Concert postponement may expire eligibility

Если Concert существенно перенесен, должны быть повторно проверены:

- Readiness;
- Teacher approval;
- согласие;
- доступность;
- технические материалы;
- состав участников.

---

## CE-049: Concert cancellation closes participation

При отмене Concert:

- Participation получает Cancelled;
- Slot снимается;
- Program Entry удаляется;
- Student и Staff уведомляются;
- Song Readiness сохраняется;
- согласие не переносится автоматически на новое мероприятие.

---

## CE-050: Eligibility decisions are versioned

Каждое значимое решение создает новую версию.

Старое решение сохраняется.

---

## CE-051: Historical decisions must remain reproducible

Для каждой версии сохраняются:

- Concert Requirements Version;
- Song Version;
- Readiness Reference;
- Approval References;
- Consent References;
- Rehearsal References;
- Technical References;
- Policy Version;
- Decision;
- Reason Codes.

---

## CE-052: Duplicate triggers must be idempotent

Повторная обработка одного события не создает:

- вторую Participation;
- второе решение;
- повторный Slot;
- повторное уведомление;
- дублирующий Program Entry.

Reason Code: `CONCERT_ELIGIBILITY_TRIGGER_ALREADY_PROCESSED`

---

## CE-053: Concurrent updates require reevaluation

Если одновременно изменились Readiness, Consent и Slot, политика должна повторно оценить актуальное состояние.

Reason Code: `CONCERT_ELIGIBILITY_VERSION_CONFLICT`

---

## CE-054: Eligibility cannot be calculated exclusively on the client

Web или mobile client может показать предварительную проверку.

Финальное доменное решение принимает backend.

---

## CE-055: Unauthorized actors cannot override eligibility

Override требует отдельного полномочия и Audit.

Reason Code: `CONCERT_ELIGIBILITY_OVERRIDE_NOT_AUTHORIZED`

---

## CE-056: Override cannot invent missing consent

Ни Teacher, ни Administrator, ни Owner не могут подменить обязательное согласие Student или Guardian.

---

## CE-057: Override cannot fabricate pedagogical approval

Организационный Actor не может самостоятельно подтвердить педагогическую готовность.

---

## CE-058: Conditional eligibility must contain explicit conditions

Недопустимо:

> Conditionally Eligible

без списка условий.

Допустимо:

> Conditions:
> - Complete stage rehearsal by 2026-08-10
> - Upload approved backing track
> - Confirm participation after final schedule

Reason Code: `ELIGIBILITY_CONDITIONS_REQUIRED`

---

## CE-059: Conditions must be verifiable

Условие должно иметь:

- тип;
- ответственного;
- срок, если применимо;
- критерий выполнения;
- статус;
- Evidence Reference.

---

## CE-060: Expired conditions invalidate conditional eligibility

Если обязательное условие не выполнено к сроку, Participation должна быть повторно оценена.

---

## CE-061: Not Eligible requires next-step guidance

Решение должно по возможности указывать:

- что препятствует участию;
- можно ли исправить;
- кто отвечает;
- до какого срока;
- возможен ли другой Concert;
- возможна ли другая Song или версия.

---

## CE-062: Student-visible explanation must be respectful

Недопустимо:

> Вы недостаточно хорошо поете для концерта.

Допустимо:

> Для этого концерта требуется еще одна полная репетиция с микрофоном. После нее преподаватель повторно подтвердит готовность номера.

---

## CE-063: Eligibility must not compare students

Допуск не должен зависеть от рейтинга Student среди других.

Отбор программы может учитывать художественный баланс, но это отдельное организационное решение.

Reason Code: `STUDENT_RANKING_NOT_ELIGIBILITY_CRITERION`

---

## CE-064: Payment status is not pedagogical eligibility

Финансовые отношения не должны подменять педагогический допуск.

Если отдельные бизнес-правила требуют урегулирования оплаты мероприятия, они должны существовать вне Concert Eligibility Policy.

---

## CE-065: CRM status is not eligibility evidence

Статус клиента в CRM не определяет готовность к Concert.

---

## CE-066: Public program publication requires approved data

Перед публикацией должны быть проверены:

- имя или сценическое имя;
- Song title;
- авторство и необходимые атрибуты;
- Performance Type;
- Visibility;
- consent;
- порядок;
- длительность.

---

## CE-067: Private data must not leak into program

Program Entry не должен содержать:

- внутренние Assessment;
- Blocking Issues;
- медицинские сведения;
- private Teacher notes;
- Reason Codes;
- историю отказов.

---

## CE-068: Recording status is separate from performance status

Student может:

- участвовать без публикации записи;
- разрешить внутреннюю запись, но не публикацию;
- разрешить фото, но не видео;
- позже отозвать разрешение на будущую публикацию в рамках действующих правил.

---

## CE-069: Readiness withdrawal triggers eligibility review

Если Teacher отзывает Performance Ready, связанная Participation должна быть приостановлена или пересмотрена.

Reason Code: `SONG_READINESS_WITHDRAWN`

---

## CE-070: Teacher approval withdrawal triggers suspension

Decision: Suspended

Reason Code: `TEACHER_APPROVAL_WITHDRAWN`

---

## CE-071: Consent withdrawal triggers participation withdrawal

Если обязательное согласие отозвано:

Eligibility Status: Withdrawn или Eligibility Status: Suspended

в зависимости от контекста и возможности повторного согласия.

---

## CE-072: Technical rejection may preserve pedagogical eligibility

Если минус поврежден, Student может оставаться педагогически готовым.

Общий результат:

- Decision: Conditionally Eligible
- Condition: Replace technical asset

---

## CE-073: Performance slot is not evidence of eligibility

Ошибочно назначенный Slot не должен создавать допуск.

---

## CE-074: Published program is not immutable

Program Entry может быть изменен, но каждое изменение должно быть версионировано и аудировано.

---

## CE-075: Final eligibility check may be required

Concert может задавать обязательную повторную проверку:

- за неделю;
- за день;
- после генеральной репетиции;
- непосредственно перед публикацией программы.

---

# Eligibility Evaluation Matrix

| Dimension | Required Evidence | Possible Failure |
| --- | --- | --- |
| Pedagogical | Song Readiness, Teacher Approval | Readiness missing or stale |
| Consent | Student Consent, Guardian Consent | Consent absent or withdrawn |
| Rehearsal | Rehearsal Completion, Assessment | Required rehearsal missing |
| Technical | Approved backing track and requirements | Asset missing or invalid |
| Schedule | Availability confirmation | Conflict or absence |
| Format | Allowed performance type | Unsupported format |
| Duration | Confirmed runtime | Duration exceeds limit |
| Participants | All participants eligible | One or more participants blocked |
| Program | Coordinator decision | No available slot |
| Publication | Visibility and media rules | Consent or metadata missing |

---

# Conditional Eligibility Model

```text
EligibilityCondition
├── ConditionId
├── ParticipationId
├── ConditionType
├── Description
├── ResponsibleActorId
├── RequiredEvidenceType
├── DueAt
├── Status
├── SatisfiedBy
├── SatisfiedAt
└── FailureConsequence
```

Допустимые статусы:

- Pending;
- Satisfied;
- Waived;
- Failed;
- Expired;
- Cancelled.

---

# Typical Conditions

- Teacher approval pending;
- rehearsal required;
- backing track required;
- technical approval required;
- duration reduction required;
- Student confirmation pending;
- guardian consent pending;
- duet partner confirmation pending;
- arrangement update pending;
- blocking section review pending;
- final availability confirmation pending.

---

# Evaluation Flow

```text
Concert Participation proposed
        |
        v
Validate Concert and Participation
        |
        v
Load Concert Requirements Version
        |
        v
Resolve Student, Song Version and Performance Type
        |
        v
Evaluate pedagogical readiness
        |
        +--> Teacher Review Required
        |
        +--> Context Reassessment Required
        |
        v
Evaluate consent
        |
        +--> Consent Required
        |
        v
Evaluate rehearsal requirements
        |
        +--> Rehearsal Required
        |
        v
Evaluate technical requirements
        |
        +--> Technical Review Required
        |
        v
Evaluate duration, format and participants
        |
        +--> Not Eligible
        |
        +--> Conditionally Eligible
        |
        v
Mark Eligible
        |
        v
Concert Coordinator program review
        |
        +--> Candidate
        |
        +--> Waitlisted
        |
        +--> Selected
        |
        v
Assign Slot
        |
        v
Final eligibility validation
        |
        v
Publish Program Entry
```

---

# Commands Produced

## CreateConcertParticipation

Создает заявку на участие.

## EvaluateConcertEligibility

Запускает оценку на текущих данных.

## MarkConcertParticipationEligible

Фиксирует подтвержденный допуск.

## MarkConcertParticipationConditionallyEligible

Фиксирует допуск с проверяемыми условиями.

## RejectConcertParticipation

Фиксирует Not Eligible с Reason Codes и следующим шагом.

## RequestTeacherConcertApproval

Создает педагогический запрос.

## RequestStudentConsent

Запрашивает согласие Student.

## RequestGuardianConsent

Запрашивает согласие законного представителя.

## RequestConcertRehearsal

Создает требование пройти репетицию.

## RequestTechnicalAsset

Запрашивает отсутствующий материал.

## RequestTechnicalAssetReview

Создает задачу технической проверки.

## RequestConcertEligibilityReview

Создает ручной Review сложного случая.

## AssignPerformanceSlot

Назначает время и место в программе.

## RemovePerformanceSlot

Удаляет Slot с обязательной причиной.

## SuspendConcertEligibility

Временно приостанавливает допуск.

## RestoreConcertEligibility

Восстанавливает допуск после повторной оценки.

## WithdrawConcertParticipation

Фиксирует добровольный или организационный отказ.

## PublishConcertProgramEntry

Публикует безопасное представление номера.

## RemoveConcertProgramEntry

Снимает номер из публичной программы.

## NotifyConcertCoordinator

Сообщает о блокере или необходимости решения.

---

# Domain Events

```text
ConcertParticipationCreated
ConcertEligibilityEvaluationStarted
ConcertParticipationMarkedEligible
ConcertParticipationMarkedConditionallyEligible
ConcertParticipationRejected
ConcertEligibilitySuspended
ConcertEligibilityRestored
ConcertEligibilityExpired
ConcertTeacherApprovalRequested
ConcertTeacherApprovalGranted
ConcertTeacherApprovalWithdrawn
ConcertConsentRequested
ConcertConsentGranted
ConcertConsentWithdrawn
ConcertRehearsalRequired
ConcertRehearsalConfirmed
ConcertTechnicalAssetRequested
ConcertTechnicalAssetApproved
ConcertTechnicalAssetRejected
PerformanceSlotAssigned
PerformanceSlotChanged
PerformanceSlotRemoved
ConcertParticipationWaitlisted
ConcertParticipationSelected
ConcertParticipationWithdrawn
ConcertProgramEntryPublished
ConcertProgramEntryRemoved
ConcertEligibilityReevaluationRequested
```

## ConcertParticipationMarkedEligible Event

Событие должно содержать:

- ParticipationId;
- ConcertId;
- StudentId;
- PerformanceType;
- SongId;
- SongVersionId;
- ParticipantIds;
- ConcertRequirementsVersion;
- ReadinessReference;
- TeacherApprovalReference;
- ConsentReferences;
- RehearsalReferences;
- TechnicalApprovalReferences;
- EligibilityDecisionId;
- PolicyId;
- PolicyVersion;
- EvaluatedAt;
- CorrelationId;
- CausationId.

Событие не должно содержать полный текст закрытых Assessment.

---

# Eligibility Decision

```text
ConcertEligibilityDecision
├── DecisionId
├── ParticipationId
├── Outcome
├── DimensionResults
├── Conditions
├── BlockingIssues
├── ReasonCodes
├── EvidenceReferences
├── EvaluatedBy
├── HumanApprovals
├── PolicyId
├── PolicyVersion
├── RequirementsVersion
├── EvaluatedAt
└── Version
```

---

# Human Review

Human Review обязателен при:

- противоречивых Teacher Assessments;
- устаревшей Song Readiness;
- значительном изменении Song Version;
- смене участников;
- пропуске обязательной репетиции;
- спорном техническом ограничении;
- запросе на override;
- возможном Safety concern;
- несоответствии контекстов;
- ручном удалении из уже опубликованной программы;
- конфликте между Student и Teacher;
- AI-generated recommendation;
- нестандартном Performance Type.

## Review Result

Reviewer может:

- подтвердить Eligibility;
- создать Conditions;
- отклонить;
- запросить дополнительную репетицию;
- запросить новую Song Version;
- изменить Performance Type;
- сократить номер;
- заменить Song;
- отложить до следующего Concert;
- временно приостановить;
- передать решение другому Actor;
- разрешить документированное исключение.

Каждое решение должно быть объяснимым.

---

# Override Model

Override допустим только для конкретных организационных или технических правил, которые Concert Requirements явно разрешают переопределять.

Нельзя override:

- обязательное согласие;
- отсутствие Student;
- чужую идентичность;
- педагогическое подтверждение без полномочий;
- требования безопасности;
- обязательное юридическое ограничение.

```text
EligibilityOverride
├── OverrideId
├── ParticipationId
├── RuleId
├── PreviousOutcome
├── NewOutcome
├── Reason
├── AuthorizedBy
├── AuthorizedAt
├── ExpiresAt
└── AuditReference
```

---

# Program Selection

Concert Eligibility Policy не определяет художественный состав программы полностью.

После получения Eligibility Concert Coordinator может учитывать:

- общую длительность;
- разнообразие форматов;
- технические переходы;
- порядок номеров;
- количество выступлений одного Student;
- нагрузку на участников;
- темп мероприятия;
- возможности площадки;
- доступное количество слотов.

Но Coordinator не должен утверждать, что Student педагогически не готов, если решение относится только к ограничению программы.

Правильный результат:

```text
Eligibility: Eligible
Program Status: Waitlisted
Reason: Program capacity limit
```

---

# Final Check

Перед окончательной публикацией или Concert может выполняться Final Eligibility Check.

Проверяются:

- Concert не отменен;
- Participation не Withdrawn;
- согласие активно;
- Song Version не изменилась;
- Teacher Approval активно;
- обязательные Conditions выполнены;
- технический материал утвержден;
- Slot актуален;
- участники доступны;
- не возникли новые Blocking Issues.

---

# Failure Modes

## Concert not found

- Decision: Rejected
- Reason Code: CONCERT_NOT_FOUND

## Participation not found

- Decision: Rejected
- Reason Code: CONCERT_PARTICIPATION_NOT_FOUND

## Student mismatch

- Decision: Rejected
- Reason Code: CONCERT_PARTICIPATION_STUDENT_MISMATCH

Security Audit обязателен.

## Song version missing

- Decision: Insufficient Data
- Reason Code: CONCERT_SONG_VERSION_REQUIRED

## Song readiness stale

- Decision: Context Reassessment Required
- Reason Code: SONG_READINESS_EXPIRED

## Teacher approval missing

- Decision: Teacher Review Required
- Reason Code: TEACHER_CONCERT_APPROVAL_REQUIRED

## Student consent missing

- Decision: Consent Required
- Reason Code: STUDENT_CONSENT_REQUIRED

## Guardian consent missing

- Decision: Consent Required
- Reason Code: GUARDIAN_CONSENT_REQUIRED

## Rehearsal missing

- Decision: Conditionally Eligible
- Condition: Complete mandatory rehearsal
- Reason Code: CONCERT_REHEARSAL_REQUIRED

## Backing track missing

- Decision: Conditionally Eligible
- Reason Code: TECHNICAL_ASSET_REQUIRED

## Technical asset rejected

- Decision: Technical Review Required
- Reason Code: TECHNICAL_ASSET_NOT_APPROVED

## Performance too long

- Decision: Not Eligible or Conditionally Eligible (если допускается сокращение)
- Reason Code: PERFORMANCE_DURATION_EXCEEDED

## Duet partner not ready

- Decision: Conditionally Eligible or Not Eligible
- Reason Code: DUET_PARTICIPANT_NOT_ELIGIBLE

## Schedule conflict

- Decision: Concert Coordinator Review Required
- Reason Code: STUDENT_SCHEDULE_CONFLICT

## Concert requirements changed

- Decision: Suspended (до повторной оценки)
- Reason Code: CONCERT_REQUIREMENTS_CHANGED

## Teacher approval withdrawn

- Decision: Suspended
- Reason Code: TEACHER_APPROVAL_WITHDRAWN

## Consent withdrawn

- Decision: Withdrawn
- Reason Code: STUDENT_CONSENT_WITHDRAWN

## Duplicate trigger

- Decision: Already Eligible or current state without new effects
- Reason Code: CONCERT_ELIGIBILITY_TRIGGER_ALREADY_PROCESSED

## Version conflict

- Decision: Deferred
- Reason Code: CONCERT_ELIGIBILITY_VERSION_CONFLICT

Политика повторно оценивается на актуальном состоянии.

---

# Explainability Examples

## Eligible

> Преподаватель подтвердил готовность концертной версии песни. Обязательная репетиция завершена, минусовка проверена, а участие подтверждено вами.

## Conditionally Eligible

> Номер может быть включен в программу после финальной репетиции с микрофоном и загрузки утвержденной минусовки до 10 августа.

## Readiness Expired

> Предыдущее подтверждение готовности было сделано до длительного перерыва. Для участия требуется короткая повторная проверка с преподавателем.

## Duration Exceeded

> Текущая версия длится 5 минут 20 секунд, а лимит номера на этом концерте — 4 минуты. Можно подготовить сокращенную концертную версию.

## Duet Condition

> Ваша часть номера готова. Для окончательного допуска необходима совместная репетиция с партнером.

## Waitlisted

> Номер соответствует требованиям концерта, но все доступные слоты уже заняты. Заявка сохранена в листе ожидания.

## Withdrawn

> Участие отменено по вашему запросу. Это не влияет на прогресс и готовность песни.

---

# Examples

## Example 1: Solo performance fully ready

Дано:

- Song имеет актуальный Performance Ready;
- Teacher подтвердил конкретный Concert;
- Student согласен;
- репетиция пройдена;
- минус утвержден;
- длительность соответствует лимиту.

Результат:

- Decision: Eligible
- Command: MarkConcertParticipationEligible
- Reason Code: ALL_CONCERT_REQUIREMENTS_SATISFIED

## Example 2: Song ready, no consent

Дано:

- педагогические требования выполнены;
- Student еще не подтвердил участие.

Результат:

- Decision: Consent Required
- Reason Code: STUDENT_CONSENT_REQUIRED

## Example 3: Song ready for lesson only

Дано:

- Readiness Context: Regular Lesson;
- Concert Context: Public Stage;
- сценической репетиции не было.

Результат:

- Decision: Context Reassessment Required
- Reason Code: READINESS_CONTEXT_MISMATCH

## Example 4: Missing backing track

Дано:

- Student и Teacher подтвердили участие;
- Song готова;
- обязательный backing track не загружен.

Результат:

- Decision: Conditionally Eligible
- Condition: Upload and approve backing track
- Reason Code: TECHNICAL_ASSET_REQUIRED

## Example 5: Performance duration too long

Дано:

- лимит Concert — 4 минуты;
- текущая версия — 5 минут 30 секунд;
- сокращенная версия разрешена.

Результат:

- Decision: Conditionally Eligible
- Condition: Prepare approved version within 4 minutes
- Reason Code: PERFORMANCE_DURATION_EXCEEDED

## Example 6: Duet partner not confirmed

Дано:

- первый Student готов;
- второй Student не дал согласие;
- совместная репетиция не проведена.

Результат:

- Decision: Conditionally Eligible
- Conditions:
  - Obtain second participant consent
  - Complete duet rehearsal
- Reason Codes:
  - DUET_PARTICIPANT_NOT_ELIGIBLE
  - CONCERT_REHEARSAL_REQUIRED

## Example 7: Old readiness after concert postponement

Дано:

- Concert перенесен на три месяца;
- Readiness была подтверждена до переноса;
- validity period — 30 дней.

Результат:

- Decision: Suspended
- Reason Code: SONG_READINESS_EXPIRED
- Command: RequestTeacherConcertApproval

## Example 8: Student withdraws

Дано:

- Participation уже была Eligible;
- Student отозвал согласие.

Результат:

- Decision: Withdrawn
- Commands:
  - WithdrawConcertParticipation
  - RemovePerformanceSlot
  - RemoveConcertProgramEntry
- Reason Code: STUDENT_CONSENT_WITHDRAWN

## Example 9: Concert is full

Дано:

- Participation полностью Eligible;
- доступных Slot нет.

Результат:

- Eligibility Status: Eligible
- Program Status: Waitlisted
- Reason Code: CONCERT_PROGRAM_CAPACITY_REACHED

## Example 10: Teacher withdraws approval

Дано:

- после репетиции появились значимые проблемы;
- Teacher отозвал approval.

Результат:

- Decision: Suspended
- Reason Code: TEACHER_APPROVAL_WITHDRAWN
- Command: RequestConcertEligibilityReview

---

# Student Presentation

Student должен видеть:

- Concert;
- Song и версию;
- формат выступления;
- текущий Eligibility Status;
- выполненные требования;
- оставшиеся условия;
- сроки;
- согласие;
- репетиции;
- технические материалы;
- Program Status;
- Slot после назначения;
- понятное объяснение решения.

Следует избегать:

- внутренних Reason Codes;
- private Teacher notes;
- сравнений с другими;
- формулировок, создающих стыд;
- обещания участия до назначения Slot;
- автоматической публикации без consent.

---

# Teacher Presentation

Teacher должен видеть:

- Song Readiness;
- Concert Requirements;
- Song Version;
- полный педагогический Evidence timeline;
- Rehearsal Results;
- Blocking Issues;
- Student Self Confirmation;
- текущий Eligibility Status;
- условия;
- необходимость Review;
- изменения после approval;
- AI recommendations;
- Program Status.

---

# Concert Coordinator Presentation

Coordinator должен видеть:

- Eligible и Conditionally Eligible заявки;
- Performance Type;
- длительность;
- количество участников;
- технические требования;
- доступность;
- выполненность условий;
- Program Status;
- Slot;
- публичные данные;
- блокеры, относящиеся к организации.

Coordinator не должен автоматически видеть закрытые педагогические заметки.

---

# Administrator Presentation

Administrator может видеть:

- статус заявки;
- дедлайны;
- согласия;
- репетиции;
- технические материалы;
- Slot;
- Program Status;
- задачи и блокеры.

Педагогическое основание отображается в минимально необходимом объеме.

---

# Owner Analytics

Owner может видеть агрегированные данные:

- количество Participation;
- Eligible / Conditional / Not Eligible;
- частоту причин блокировки;
- заполненность программы;
- количество Waitlisted;
- долю отмен;
- технические задержки;
- частоту повторных проверок;
- своевременность consent;
- количество изменений программы;
- нагрузку на Teachers и Coordinators.

Owner Analytics не должна:

- ранжировать учеников;
- подменять педагогические решения;
- делать вывод о качестве Teacher только по числу допусков;
- раскрывать sensitive notes.

---

# AI Assistance

AI может:

- проверять полноту заявки;
- сопоставлять Concert Requirements;
- находить отсутствующие документы;
- обнаруживать устаревшую Readiness;
- предлагать список Conditions;
- проверять длительность;
- искать конфликт расписания;
- структурировать технические требования;
- создавать Draft Explanation;
- обнаруживать изменения Song Version;
- предлагать порядок ручного Review.

AI не может:

- подтверждать педагогическую готовность;
- выдавать финальный Eligibility;
- подменять consent;
- принимать решение об исключении;
- ставить медицинские выводы;
- придумывать выполненные репетиции;
- публиковать Program Entry;
- назначать Slot без уполномоченного процесса;
- скрывать неопределенность.

AI proposal должен сохранять:

- model or mechanism;
- version;
- input references;
- proposed outcome;
- proposed conditions;
- confidence;
- timestamp;
- human confirmation status.

---

# Privacy

Concert Participation может содержать:

- данные о доступности;
- педагогические наблюдения;
- записи репетиций;
- сведения о несовершеннолетних;
- контактные и организационные данные;
- информацию о будущих перемещениях;
- согласия на публикацию.

Необходимо:

- применять минимальный доступ;
- разделять внутренние и публичные данные;
- не включать private notes в Program;
- ограничивать доступ к Media Assets;
- хранить consent history;
- аудитировать просмотр sensitive данных;
- не раскрывать причины отказа другим Student;
- не публиковать будущий Slot раньше разрешенного момента.

---

# Security

Необходимо защищать:

- подмену StudentId;
- изменение чужой Participation;
- фальшивое consent;
- подделку Teacher approval;
- замену технического файла после approval;
- повторное использование старого approval;
- неавторизованный override;
- публикацию закрытого Program;
- скачивание private Media Assets;
- массовый экспорт чувствительных данных.

---

# Audit Requirements

Для каждой оценки сохраняются:

- PolicyId;
- PolicyVersion;
- ConcertId;
- ConcertRequirementsVersion;
- ParticipationId;
- ParticipationVersion;
- StudentId;
- SongId;
- SongVersionId;
- PerformanceType;
- ParticipantIds;
- ReadinessReference;
- TeacherApprovalReference;
- ConsentReferences;
- RehearsalReferences;
- TechnicalReferences;
- availability state;
- ActorId;
- Decision;
- Dimension Results;
- Conditions;
- Blocking Issues;
- Reason Codes;
- AI metadata;
- EvaluatedAt;
- CorrelationId;
- CausationId.

Для ручного решения:

- ReviewerId;
- ReviewerRole;
- ReviewedAt;
- previous decision;
- final decision;
- explanation;
- override reference;
- expiration;
- resulting version.

Для Program changes:

- previous Slot;
- new Slot;
- previous Program Status;
- new Program Status;
- ChangedBy;
- ChangeReason;
- PublishedAt.

---

# Test Requirements

## Basic Eligibility Tests

- valid solo Participation becomes Eligible;
- missing Concert is rejected;
- inactive Concert cannot accept new Participation;
- missing Song Version produces Insufficient Data;
- unsupported Performance Type is rejected;
- duplicate evaluation is idempotent.

## Readiness Tests

- current Performance Ready is accepted;
- Lesson-only readiness is insufficient for public stage;
- stale readiness requires reassessment;
- changed key triggers review;
- changed arrangement triggers review;
- withdrawn readiness suspends eligibility;
- AI-only readiness is insufficient.

## Teacher Approval Tests

- required approval blocks eligibility when absent;
- authorized Teacher can approve;
- unrelated Teacher cannot approve;
- approval is scoped to Concert and Song Version;
- approval withdrawal suspends eligibility;
- old approval is not reused after material change.

## Consent Tests

- Student Consent is required when configured;
- Student can withdraw consent;
- Guardian Consent is required where configured;
- media consent is independent;
- missing media consent does not always block participation;
- consent cannot be overridden;
- consent history is preserved.

## Rehearsal Tests

- mandatory rehearsal creates condition;
- completed rehearsal satisfies condition;
- attendance without assessment may be insufficient;
- wrong rehearsal context is rejected;
- duet requires joint rehearsal;
- missed rehearsal suspends eligibility when required.

## Technical Tests

- missing backing track creates condition;
- uploaded file is not automatically approved;
- rejected asset blocks final eligibility;
- replacement asset creates new version;
- duplicate approval is idempotent;
- duration is calculated from approved version.

## Duration Tests

- performance within limit is accepted;
- performance exceeding limit is rejected or conditional;
- shortened version can restore eligibility;
- duration change triggers reevaluation;
- transition time is included when configured.

## Participant Tests

- duet requires both participants;
- one withdrawn participant suspends duet;
- ensemble composition is versioned;
- changing ensemble participant triggers review;
- individual readiness does not prove ensemble readiness;
- each participant requires consent.

## Schedule Tests

- available Student passes;
- conflict requires Coordinator Review;
- Slot assignment does not create eligibility;
- Slot change can trigger availability recheck;
- unavailable Student cannot be finalized;
- program waitlist remains separate from Eligibility.

## Program Tests

- Eligible Participation can be waitlisted;
- only authorized Coordinator selects Program;
- removed number keeps pedagogical readiness;
- published Program excludes private data;
- Program change is versioned;
- cancelled Concert removes Program Entries.

## Conditional Eligibility Tests

- conditional outcome has at least one condition;
- condition has responsible Actor;
- condition can be satisfied;
- expired condition triggers reevaluation;
- waived condition requires authorization;
- all conditions satisfied can produce Eligible.

## Suspension Tests

- requirements change suspends affected Participation;
- approval withdrawal suspends;
- readiness withdrawal suspends;
- technical failure may suspend;
- restored eligibility requires reevaluation;
- suspension preserves history.

## Permission Tests

- Student can request and withdraw participation;
- Student cannot self-approve eligibility;
- Teacher cannot invent consent;
- Coordinator cannot fabricate Teacher approval;
- Administrator cannot override protected rules;
- Owner cannot bypass consent;
- AI cannot finalize eligibility.

## Versioning Tests

- every decision creates a version;
- requirements version is stored;
- Song Version is stored;
- historical decision is reproducible;
- duplicate trigger does not duplicate effects;
- concurrent updates trigger retry;
- stale update does not overwrite new decision.

## Privacy Tests

- Student sees only their Participation;
- public Program excludes private notes;
- Coordinator sees only required pedagogical summary;
- Media Assets respect consent;
- Guardian data is protected;
- withdrawn consent is reflected;
- event payloads exclude sensitive content.

## Explainability Tests

- Eligible decision is explainable;
- Conditional decision lists conditions;
- Not Eligible decision includes next step;
- Waitlist is not described as pedagogical rejection;
- withdrawal is not framed as failure;
- stale readiness explanation is understandable;
- Student-facing text contains no internal Reason Codes.

## AI Tests

- AI proposal remains non-final;
- hallucinated approval is rejected;
- missing Evidence Reference is rejected;
- AI cannot generate consent;
- AI cannot publish Program;
- AI confidence is not domain confidence;
- human confirmation is recorded.

---

# Non-Goals

Concert Eligibility Policy не определяет:

- общую Song Readiness;
- Progress Calculation;
- Goal Completion;
- Achievement Award;
- финансовые платежи;
- продажу билетов;
- CRM-статусы;
- стоимость участия;
- оплату Teacher;
- договоры с площадкой;
- авторские лицензии;
- медицинскую диагностику;
- транспорт;
- размещение гостей;
- маркетинговую кампанию;
- монтаж видеозаписей;
- окончательный художественный порядок программы;
- правила начисления BelCoins.

---

# Open Questions

Необходимо определить:

- какие типы Concert существуют;
- обязательна ли Teacher approval всегда;
- какой срок актуальности Song Readiness;
- сколько репетиций требуется;
- нужна ли генеральная репетиция;
- какие performance contexts поддерживаются;
- кто является Responsible Teacher;
- кто принимает решение при нескольких Teachers;
- какие правила могут быть overridden;
- какие правила никогда нельзя override;
- как обрабатывать несовершеннолетних;
- нужна ли отдельная модель Guardian;
- какие согласия обязательны;
- может ли Student участвовать без видеосъемки;
- как учитывать закрытые Concert;
- как публиковать сценические имена;
- как хранить Song Version;
- как версионировать backing track;
- какие аудиоформаты допустимы;
- кто утверждает Technical Asset;
- как рассчитывается длительность;
- входит ли представление Student во время номера;
- как обрабатывать medley;
- как учитывать несколько Songs в одном номере;
- как моделировать дуэты и ансамбли;
- как обрабатывать замену участника;
- может ли Teacher участвовать вместе со Student;
- как обрабатывать live band;
- как учитывать оборудование площадки;
- как моделировать технический райдер;
- какие возрастные ограничения допустимы;
- как учитывать языковые и content restrictions;
- как обрабатывать опоздание;
- когда Participation становится No Show;
- влияет ли No Show на что-либо вне Concert;
- можно ли переносить заявку на следующий Concert;
- требуется ли новое consent при переносе;
- требуется ли новое approval при переносе;
- когда Eligibility считается Expired;
- нужна ли автоматическая Final Check;
- когда публикуется Slot;
- кто может изменять программу;
- как уведомлять об изменении Slot;
- нужна ли Waitlist;
- как выбирать из Waitlist;
- как хранить Program Versions;
- нужно ли ограничивать число номеров одного Student;
- как учитывать усталость при нескольких выступлениях;
- как обрабатывать участие в разных составах;
- как фиксировать успешное Performance;
- кто публикует ConcertPerformanceAssessed;
- как связывать выступление с Achievement;
- как учитывать отмену Concert;
- как хранить consent на фото и видео;
- как реализовать отзыв согласия после публикации;
- какие данные доступны Owner;
- какие данные видит Administrator;
- какие данные видит Concert Coordinator;
- нужен ли отдельный Concert domain document;
- нужен ли отдельный Performance aggregate;
- нужен ли отдельный Program aggregate;
- нужен ли отдельный Technical Asset lifecycle;
- нужна ли отдельная Consent Policy;
- нужна ли отдельная Concert Program Policy;
- нужна ли отдельная Performance Completion Policy.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определена модель педагогического, организационного, технического и согласительного допуска к Concert. |
