package memory

import (
	"context"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.6 operations summary — parity with PostgreSQL.

func (s *Store) OperationsSummary(_ context.Context, principal core.Principal, now time.Time) (core.OperationsSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	manager := actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
	if !manager {
		return core.OperationsSummary{}, core.E(core.CodeForbidden, "the operations summary is a school view", nil)
	}
	zone, err := time.LoadLocation("Asia/Almaty")
	if err != nil {
		return core.OperationsSummary{}, core.E(core.CodeInternal, "school time zone is unavailable", err)
	}
	local := now.In(zone)
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, zone)
	dayEnd := dayStart.AddDate(0, 0, 1)
	weekEnd := dayStart.AddDate(0, 0, 7)

	var summary core.OperationsSummary
	attendanceByOccurrence := map[string]bool{}
	for _, record := range s.attendance {
		if record.TenantID == principal.TenantID {
			attendanceByOccurrence[record.OccurrenceID] = true
		}
	}
	for _, stored := range s.lessons {
		if stored.TenantID != principal.TenantID || stored.Status != core.LessonScheduled {
			continue
		}
		if !stored.StartsAt.Before(dayStart) && stored.StartsAt.Before(dayEnd) {
			summary.LessonsToday++
		}
		end := stored.StartsAt.Add(time.Duration(stored.DurationMinutes) * time.Minute)
		if end.Before(now) && !attendanceByOccurrence[stored.ID] {
			summary.PastLessonsNoAttendance++
		}
	}
	for _, request := range s.rescheduleRequests {
		if request.TenantID == principal.TenantID && request.Status == "pending" {
			summary.PendingReschedules++
		}
	}
	for _, report := range s.communityReports {
		if report.TenantID == principal.TenantID && report.Status == "new" {
			summary.NewCommunityReports++
		}
	}
	for _, journal := range s.journals {
		if journal.TenantID == principal.TenantID && journal.Status == "draft" {
			summary.DraftJournals++
		}
	}
	for _, studentRecord := range s.students {
		if studentRecord.TenantID == principal.TenantID && s.activeStudent(studentRecord) {
			summary.ActiveStudents++
		}
	}
	for _, series := range s.lessonSeries {
		if series.TenantID == principal.TenantID && series.Status == "active" {
			summary.ActiveSeries++
		}
	}
	for _, occurrence := range s.eventOccurrences {
		if occurrence.TenantID == principal.TenantID && occurrence.Status == "scheduled" &&
			!occurrence.StartsAt.Before(dayStart) && occurrence.StartsAt.Before(weekEnd) {
			summary.UpcomingEventsWeek++
		}
	}
	for _, request := range s.deletions {
		if request.TenantID == principal.TenantID &&
			(request.Status == "requested" || request.Status == "pending_review") {
			summary.PendingDeletionRequests++
		}
	}
	return summary, nil
}
