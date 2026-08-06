package app

import (
	"context"
	"slices"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.3 student repertoire (aggregate StudentSong). The stage vocabulary
// is the design's explicit journey; a new song starts at its first step
// unless the Teacher says otherwise.

type AddStudentSongInput struct {
	StudentID      string
	Title          string
	Artist         string
	Stage          string
	StageNote      string
	IdempotencyKey string
}

func validateSongStage(stage string) (string, error) {
	if !slices.Contains(core.SongStages, stage) {
		return "", core.E(core.CodeInvalidInput, "stage must be one of the journey steps", nil)
	}
	return stage, nil
}

func (s *Service) AddStudentSong(ctx context.Context, principal core.Principal, input AddStudentSongInput) (core.StudentSong, error) {
	studentID, err := security.ValidateIdentifier("studentId", input.StudentID, 128)
	if err != nil {
		return core.StudentSong{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	title, err := security.ValidateText("title", input.Title, 1, 200)
	if err != nil {
		return core.StudentSong{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	artist := ""
	if input.Artist != "" {
		artist, err = security.ValidateText("artist", input.Artist, 1, 200)
		if err != nil {
			return core.StudentSong{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	stage := core.SongStages[0]
	if input.Stage != "" {
		stage, err = validateSongStage(input.Stage)
		if err != nil {
			return core.StudentSong{}, err
		}
	}
	stageNote := ""
	if input.StageNote != "" {
		stageNote, err = security.ValidateText("stageNote", input.StageNote, 1, 500)
		if err != nil {
			return core.StudentSong{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.StudentSong{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	songID, err := security.NewID("song")
	if err != nil {
		return core.StudentSong{}, core.E(core.CodeInternal, "could not create the song id", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		StudentID, Title, Artist, Stage, StageNote string
	}{studentID, title, artist, stage, stageNote})
	if err != nil {
		return core.StudentSong{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	song, err := s.store.AddStudentSong(ctx, core.AddStudentSongCommand{
		Principal: principal, SongID: songID, StudentID: studentID,
		Title: title, Artist: artist, Stage: stage, StageNote: stageNote,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.StudentSong{}, normalizeStoreError("add student song", err)
	}
	return song, nil
}

type ChangeSongStageInput struct {
	SongID          string
	Stage           string
	StageNote       string
	ExpectedVersion int
	IdempotencyKey  string
}

func (s *Service) ChangeSongStage(ctx context.Context, principal core.Principal, input ChangeSongStageInput) (core.StudentSong, error) {
	songID, err := security.ValidateIdentifier("songId", input.SongID, 128)
	if err != nil {
		return core.StudentSong{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	stage, err := validateSongStage(input.Stage)
	if err != nil {
		return core.StudentSong{}, err
	}
	stageNote := ""
	if input.StageNote != "" {
		stageNote, err = security.ValidateText("stageNote", input.StageNote, 1, 500)
		if err != nil {
			return core.StudentSong{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if input.ExpectedVersion <= 0 {
		return core.StudentSong{}, core.E(core.CodeInvalidInput, "expectedVersion must be positive", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.StudentSong{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		SongID, Stage, StageNote string
		ExpectedVersion          int
	}{songID, stage, stageNote, input.ExpectedVersion})
	if err != nil {
		return core.StudentSong{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	song, err := s.store.ChangeSongStage(ctx, core.ChangeSongStageCommand{
		Principal: principal, SongID: songID, Stage: stage, StageNote: stageNote,
		ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey:  idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.StudentSong{}, normalizeStoreError("change song stage", err)
	}
	return song, nil
}

func (s *Service) ListStudentSongs(ctx context.Context, principal core.Principal, studentID string) ([]core.StudentSong, error) {
	normalizedID, err := security.ValidateIdentifier("studentId", studentID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	songs, err := s.store.ListStudentSongs(ctx, principal, normalizedID)
	if err != nil {
		return nil, normalizeStoreError("list repertoire", err)
	}
	return songs, nil
}
