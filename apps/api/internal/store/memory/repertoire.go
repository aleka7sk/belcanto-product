package memory

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 student repertoire — parity with PostgreSQL. Stage history is
// append-only; the stage is a named journey step, never a score.

type songStageChange struct {
	FromStage string
	ToStage   string
	Note      string
	ChangedBy string
	ChangedAt time.Time
}

type studentSong struct {
	ID         string
	TenantID   string
	StudentID  string
	Title      string
	Artist     string
	Stage      string
	StageNote  string
	AssignedBy string
	History    []songStageChange
	Version    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (s *Store) repertoireMarkerAuthority(actorID, tenantID, studentID string, now time.Time) error {
	if s.students[studentID] == nil || s.students[studentID].TenantID != tenantID {
		return core.E(core.CodeNotFound, "Student not found", nil)
	}
	if assignment := s.assignmentAt(studentID, now); assignment != nil &&
		assignment.TeacherAccountID == actorID {
		return nil
	}
	if actor := s.activeAccount(actorID, tenantID); actor != nil &&
		actor.Roles[core.RoleAdministrator] != "" {
		return nil
	}
	return core.E(core.CodeForbidden, "repertoire is led by the Student's Teacher or an Administrator", nil)
}

func (s *Store) songView(stored *studentSong) core.StudentSong {
	view := core.StudentSong{
		ID: stored.ID, StudentID: stored.StudentID, Title: stored.Title,
		Artist: stored.Artist, Stage: stored.Stage, StageNote: stored.StageNote,
		AssignedBy: core.TeacherSummary{AccountID: stored.AssignedBy},
		History:    make([]core.SongStageChange, 0, len(stored.History)),
		Version:    stored.Version, CreatedAt: stored.CreatedAt, UpdatedAt: stored.UpdatedAt,
	}
	if account := s.accounts[stored.AssignedBy]; account != nil {
		view.AssignedBy.FullName = account.FullName
	}
	// Append order is the sequence order; newest first, like PostgreSQL's
	// seq DESC.
	for index := len(stored.History) - 1; index >= 0; index-- {
		change := stored.History[index]
		view.History = append(view.History, core.SongStageChange{
			FromStage: change.FromStage, ToStage: change.ToStage,
			Note: change.Note, ChangedAt: change.ChangedAt,
		})
	}
	return view
}

func (s *Store) AddStudentSong(_ context.Context, command core.AddStudentSongCommand) (core.StudentSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if err := s.repertoireMarkerAuthority(principal.AccountID, principal.TenantID, command.StudentID, command.Now); err != nil {
		return core.StudentSong{}, err
	}
	if response, ok, err := s.replay("add_student_song", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.StudentSong{}, err
		}
		var result core.StudentSong
		if err := json.Unmarshal(response, &result); err != nil {
			return core.StudentSong{}, core.E(core.CodeInternal, "decode idempotent song result", err)
		}
		return result, nil
	}
	stored := &studentSong{
		ID: command.SongID, TenantID: principal.TenantID, StudentID: command.StudentID,
		Title: command.Title, Artist: command.Artist,
		Stage: command.Stage, StageNote: command.StageNote,
		AssignedBy: principal.AccountID, Version: 1,
		CreatedAt: command.Now, UpdatedAt: command.Now,
		History: []songStageChange{{
			ToStage: command.Stage, Note: command.StageNote,
			ChangedBy: principal.AccountID, ChangedAt: command.Now,
		}},
	}
	s.songs[command.SongID] = stored
	result := s.songView(stored)
	if err := s.completeIdempotency("add_student_song", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.StudentSong{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "SongAdded",
		"student_song", stored.ID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) ChangeSongStage(_ context.Context, command core.ChangeSongStageCommand) (core.StudentSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if response, ok, err := s.replay("change_song_stage", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.StudentSong{}, err
		}
		var result core.StudentSong
		if err := json.Unmarshal(response, &result); err != nil {
			return core.StudentSong{}, core.E(core.CodeInternal, "decode idempotent song result", err)
		}
		return result, nil
	}
	stored := s.songs[command.SongID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.StudentSong{}, core.E(core.CodeNotFound, "song not found", nil)
	}
	if err := s.repertoireMarkerAuthority(principal.AccountID, principal.TenantID, stored.StudentID, command.Now); err != nil {
		return core.StudentSong{}, err
	}
	if command.ExpectedVersion != stored.Version {
		return core.StudentSong{}, core.E(core.CodeConflict, "the song changed; reload and retry", nil)
	}
	stored.History = append(stored.History, songStageChange{
		FromStage: stored.Stage, ToStage: command.Stage, Note: command.StageNote,
		ChangedBy: principal.AccountID, ChangedAt: command.Now,
	})
	stored.Stage = command.Stage
	stored.StageNote = command.StageNote
	stored.Version++
	stored.UpdatedAt = command.Now
	result := s.songView(stored)
	if err := s.completeIdempotency("change_song_stage", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.StudentSong{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "SongStageChanged",
		"student_song", stored.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "SongStageChanged", stored.ID, command.Now)
	return result, nil
}

func (s *Store) ListStudentSongs(_ context.Context, principal core.Principal, studentID string) ([]core.StudentSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	manager, isSelf := s.journalViewerScope(principal, studentID)
	isTeacher := false
	if assignment := s.assignmentAt(studentID, timeNowFallback()); assignment != nil &&
		assignment.TeacherAccountID == principal.AccountID {
		isTeacher = true
	}
	for _, stored := range s.journals {
		if stored.TenantID == principal.TenantID && stored.StudentID == studentID &&
			stored.TeacherAccountID == principal.AccountID {
			isTeacher = true
		}
	}
	if !manager && !isSelf && !isTeacher {
		return nil, core.E(core.CodeForbidden, "repertoire is visible to the Student and assigned staff", nil)
	}
	result := []core.StudentSong{}
	for _, stored := range s.songs {
		if stored.TenantID != principal.TenantID || stored.StudentID != studentID {
			continue
		}
		result = append(result, s.songView(stored))
	}
	sort.Slice(result, func(left, right int) bool {
		if !result[left].UpdatedAt.Equal(result[right].UpdatedAt) {
			return result[left].UpdatedAt.After(result[right].UpdatedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}
