package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.6 operations summary (Pages 29/30). Every number is derived from
// stored rows — organisational signals for the school, never a rating
// of people. Owner and Administrator only.

func (s *Store) OperationsSummary(ctx context.Context, principal core.Principal, now time.Time) (core.OperationsSummary, error) {
	var summary core.OperationsSummary
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		manager, err := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if err != nil {
			return err
		}
		if !manager {
			return core.E(core.CodeForbidden, "the operations summary is a school view", nil)
		}
		// Almaty civil day for «сегодня»; the school runs on Asia/Almaty.
		zone, err := time.LoadLocation("Asia/Almaty")
		if err != nil {
			return core.E(core.CodeInternal, "school time zone is unavailable", err)
		}
		local := now.In(zone)
		dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, zone)
		dayEnd := dayStart.AddDate(0, 0, 1)
		weekEnd := dayStart.AddDate(0, 0, 7)
		return tx.QueryRow(ctx, `
			SELECT
				(SELECT count(*) FROM core_lesson_occurrences
				 WHERE tenant_id = $1 AND status = 'scheduled'
				   AND starts_at >= $2 AND starts_at < $3),
				(SELECT count(*) FROM lesson_reschedule_requests
				 WHERE tenant_id = $1 AND status = 'pending'),
				(SELECT count(*) FROM community_reports
				 WHERE tenant_id = $1 AND status = 'new'),
				(SELECT count(*) FROM lesson_journals
				 WHERE tenant_id = $1 AND status = 'draft'),
				(SELECT count(*) FROM core_lesson_occurrences o
				 WHERE o.tenant_id = $1 AND o.status = 'scheduled'
				   AND o.starts_at + (o.duration_minutes || ' minutes')::interval < $4
				   AND NOT EXISTS (
					SELECT 1 FROM core_lesson_attendance a
					WHERE a.tenant_id = o.tenant_id AND a.occurrence_id = o.id
				 )),
				(SELECT count(*) FROM students
				 WHERE tenant_id = $1 AND status = 'active'),
				(SELECT count(*) FROM core_lesson_series
				 WHERE tenant_id = $1 AND status = 'active'),
				(SELECT count(*) FROM event_occurrences
				 WHERE tenant_id = $1 AND status = 'scheduled'
				   AND starts_at >= $2 AND starts_at < $5),
				(SELECT count(*) FROM account_deletion_requests
				 WHERE tenant_id = $1 AND status IN ('requested', 'pending_review'))
		`, principal.TenantID, dayStart, dayEnd, now, weekEnd).Scan(
			&summary.LessonsToday,
			&summary.PendingReschedules,
			&summary.NewCommunityReports,
			&summary.DraftJournals,
			&summary.PastLessonsNoAttendance,
			&summary.ActiveStudents,
			&summary.ActiveSeries,
			&summary.UpcomingEventsWeek,
			&summary.PendingDeletionRequests,
		)
	})
	if err != nil {
		return core.OperationsSummary{}, err
	}
	return summary, nil
}
