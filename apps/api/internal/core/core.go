package core

import (
	"errors"
	"time"
)

type Role string

const (
	RoleOwner         Role = "Owner"
	RoleAdministrator Role = "Administrator"
	RoleTeacher       Role = "Teacher"
	RoleStudent       Role = "Student"
)

const StudentOnboardingManagerV1 = "StudentOnboardingManager.v1"

const (
	PermissionStudentsCreate            = "students.create"
	PermissionStudentOnboardingRead     = "student_onboarding.read"
	PermissionStudentInvitationsIssue   = "student_invitations.issue"
	PermissionStudentInvitationsReissue = "student_invitations.reissue"
	PermissionStudentInvitationsRevoke  = "student_invitations.revoke"
	PermissionStudentOnboardingDelegate = "student_onboarding.delegate"
)

func StudentOnboardingManagerV1PermissionSet() []string {
	return []string{
		PermissionStudentsCreate,
		PermissionStudentOnboardingRead,
	}
}

func OwnerStudentOnboardingPermissionSet() []string {
	return []string{
		PermissionStudentsCreate,
		PermissionStudentOnboardingRead,
		PermissionStudentInvitationsIssue,
		PermissionStudentInvitationsReissue,
		PermissionStudentInvitationsRevoke,
		PermissionStudentOnboardingDelegate,
	}
}

type ErrorCode string

const (
	CodeInvalidInput      ErrorCode = "INVALID_INPUT"
	CodeUnauthenticated   ErrorCode = "UNAUTHENTICATED"
	CodeForbidden         ErrorCode = "FORBIDDEN"
	CodeNotFound          ErrorCode = "NOT_FOUND"
	CodeConflict          ErrorCode = "CONFLICT"
	CodeInvalidState      ErrorCode = "INVALID_STATE"
	CodeInvalidActivation ErrorCode = "ACTIVATION_INVALID"
	CodeRateLimited       ErrorCode = "RATE_LIMITED"
	CodeUnavailable       ErrorCode = "UNAVAILABLE"
	CodeInternal          ErrorCode = "INTERNAL"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *AppError) Unwrap() error { return e.Err }

func E(code ErrorCode, message string, err error) error {
	return &AppError{Code: code, Message: message, Err: err}
}

func IsCode(err error, code ErrorCode) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == code
}

type Principal struct {
	AccountID string
	TenantID  string
	SessionID string
	Roles     []Role
}

func (p Principal) HasRole(role Role) bool {
	for _, candidate := range p.Roles {
		if candidate == role {
			return true
		}
	}
	return false
}

type BootstrapOwnerCommand struct {
	TenantID     string
	TenantName   string
	FullName     string
	Phone        string
	InvitationID string
	AccountID    string
	PersonID     string
	MembershipID string
	RoleGrantID  string
	Operator     string
	Reason       string
	TokenDigest  []byte
	Now          time.Time
	ExpiresAt    time.Time
}

type BootstrapStaffCommand struct {
	TenantID       string
	OwnerAccountID string
	FullName       string
	Phone          string
	Role           Role
	InvitationID   string
	AccountID      string
	PersonID       string
	MembershipID   string
	RoleGrantID    string
	Operator       string
	Reason         string
	TokenDigest    []byte
	Now            time.Time
	ExpiresAt      time.Time
}

type ReissueBootstrapInvitationCommand struct {
	TenantID     string
	AccountID    string
	InvitationID string
	Operator     string
	Reason       string
	TokenDigest  []byte
	Now          time.Time
	ExpiresAt    time.Time
}

type BootstrapInvitationResult struct {
	InvitationID string    `json:"invitationId"`
	AccountID    string    `json:"accountId"`
	Kind         string    `json:"kind"`
	Status       string    `json:"status"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type ActivationPreview struct {
	InvitationID string    `json:"invitationId"`
	Kind         string    `json:"kind"`
	DisplayName  string    `json:"displayName"`
	MaskedPhone  string    `json:"maskedPhone"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type ActivationCompleteCommand struct {
	TokenDigest        []byte
	Phone              string
	PasswordHash       string
	IdempotencyKey     string
	PayloadFingerprint []byte
	Now                time.Time
}

type ActivationValidationCommand struct {
	TokenDigest        []byte
	Phone              string
	IdempotencyKey     string
	PayloadFingerprint []byte
	Now                time.Time
}

type CredentialRecord struct {
	AccountID    string
	TenantID     string
	Phone        string
	PasswordHash string
	Status       string
	Roles        []Role
}

type SessionMaterial struct {
	SessionID        string
	FamilyID         string
	AccessDigest     []byte
	RefreshDigest    []byte
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	CreatedAt        time.Time
}

type SessionTokens struct {
	AccessToken      string    `json:"accessToken"`
	RefreshToken     string    `json:"refreshToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
}

type GrantDelegationCommand struct {
	ID                 string
	TenantID           string
	OwnerAccountID     string
	AdministratorID    string
	Bundle             string
	Reason             string
	ExpiresAt          *time.Time
	IdempotencyKey     string
	PayloadFingerprint []byte
	Now                time.Time
}

type DelegationResult struct {
	ID              string     `json:"id"`
	AdministratorID string     `json:"administratorAccountId"`
	Bundle          string     `json:"bundle"`
	Status          string     `json:"status"`
	GrantedAt       time.Time  `json:"grantedAt"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
}

type RevokeDelegationCommand struct {
	TenantID           string
	OwnerAccountID     string
	DelegationID       string
	Reason             string
	IdempotencyKey     string
	PayloadFingerprint []byte
	Now                time.Time
}

type CreateStudentCommand struct {
	TenantID            string
	ActorAccountID      string
	PersonID            string
	MembershipID        string
	StudentID           string
	AccountID           string
	RoleGrantID         string
	TeacherAssignmentID string
	FullName            string
	Phone               string
	EnrollmentReference string
	TeacherAccountID    string
	Locale              string
	Timezone            string
	AdultConfirmed      bool
	IdempotencyKey      string
	PayloadFingerprint  []byte
	Now                 time.Time
}

type StudentResult struct {
	StudentID       string `json:"studentId"`
	AccountID       string `json:"accountId"`
	OnboardingState string `json:"onboardingState"`
}

type PublishFirstMinuteCommand struct {
	TenantID           string
	ActorAccountID     string
	StudentID          string
	RevisionID         string
	WhatWorked         string
	CurrentFocus       string
	NextStep           string
	ExpectedVersion    int64
	IdempotencyKey     string
	PayloadFingerprint []byte
	Now                time.Time
}

type FirstMinute struct {
	StudentID    string    `json:"studentId"`
	Revision     int64     `json:"revision"`
	WhatWorked   string    `json:"whatWorked"`
	CurrentFocus string    `json:"currentFocus"`
	NextStep     string    `json:"nextStep"`
	PublishedAt  time.Time `json:"publishedAt"`
}

type InvitationMode string

const (
	InvitationIssue   InvitationMode = "issue"
	InvitationReissue InvitationMode = "reissue"
)

type IssueInvitationCommand struct {
	TenantID           string
	ActorAccountID     string
	StudentID          string
	InvitationID       string
	TokenDigest        []byte
	Mode               InvitationMode
	ExpiresAt          time.Time
	IdempotencyKey     string
	PayloadFingerprint []byte
	Now                time.Time
}

type InvitationResult struct {
	InvitationID string    `json:"invitationId"`
	StudentID    string    `json:"studentId"`
	Status       string    `json:"status"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type RevokeInvitationCommand struct {
	TenantID           string
	ActorAccountID     string
	InvitationID       string
	IdempotencyKey     string
	PayloadFingerprint []byte
	Now                time.Time
}

type BootstrapView struct {
	AccountID      string       `json:"accountId"`
	Roles          []Role       `json:"roles"`
	AccessProfiles []string     `json:"accessProfiles"`
	Permissions    []string     `json:"permissions"`
	StudentID      string       `json:"studentId,omitempty"`
	FullName       string       `json:"fullName,omitempty"`
	FirstMinute    *FirstMinute `json:"firstMinute,omitempty"`
}

type StaffMember struct {
	AccountID                     string     `json:"accountId"`
	FullName                      string     `json:"fullName"`
	Roles                         []Role     `json:"roles"`
	AccessProfiles                []string   `json:"accessProfiles"`
	OnboardingDelegationID        string     `json:"onboardingDelegationId,omitempty"`
	OnboardingDelegationExpiresAt *time.Time `json:"onboardingDelegationExpiresAt,omitempty"`
}

type OnboardingState string

const (
	OnboardingAwaitingFirstMinute OnboardingState = "awaiting_first_minute"
	OnboardingReadyToInvite       OnboardingState = "ready_to_invite"
	OnboardingInvited             OnboardingState = "invited"
	OnboardingActivated           OnboardingState = "activated"
)

type StudentOnboardingItem struct {
	StudentID           string          `json:"studentId"`
	FullName            string          `json:"fullName"`
	EnrollmentReference string          `json:"enrollmentReference"`
	TeacherAccountID    string          `json:"teacherAccountId"`
	StudentVersion      int64           `json:"studentVersion"`
	OnboardingState     OnboardingState `json:"onboardingState"`
	InvitationID        string          `json:"invitationId,omitempty"`
	InvitationExpiresAt *time.Time      `json:"invitationExpiresAt,omitempty"`
}
