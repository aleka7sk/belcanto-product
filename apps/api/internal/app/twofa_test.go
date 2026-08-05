package app_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

func totpNow(t *testing.T, fixture *fixture, secret string) string {
	t.Helper()
	code, err := security.TOTPCode(secret, fixture.clock.Now())
	if err != nil {
		t.Fatalf("compute TOTP code: %v", err)
	}
	return code
}

func lastOutboxAggregate(t *testing.T, fixture *fixture, eventType string) string {
	t.Helper()
	records := fixture.store.OutboxRecords()
	for index := len(records) - 1; index >= 0; index-- {
		if records[index].EventType == eventType {
			return records[index].AggregateID
		}
	}
	t.Fatalf("outbox event %s not found", eventType)
	return ""
}

func TestTwofaEnrollmentSignInAndDisable(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	owner := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)

	if _, err := fixture.service.StartTwofaEnrollment(ctx, owner, "wrong-password"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("enrollment without re-auth = %v", err)
	}
	enrollment, err := fixture.service.StartTwofaEnrollment(ctx, owner, ownerPassword)
	if err != nil {
		t.Fatalf("start enrollment: %v", err)
	}
	if enrollment.Secret == "" || !strings.HasPrefix(enrollment.ProvisioningURI, "otpauth://totp/Belcanto:") {
		t.Fatalf("enrollment material = %#v", enrollment)
	}
	if _, err := fixture.service.ConfirmTwofaEnrollment(ctx, owner, "000000"); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("confirm with wrong code = %v", err)
	}
	recoveryCodes, err := fixture.service.ConfirmTwofaEnrollment(ctx, owner, totpNow(t, fixture, enrollment.Secret))
	if err != nil {
		t.Fatalf("confirm enrollment: %v", err)
	}
	if len(recoveryCodes) != 10 {
		t.Fatalf("recovery codes = %d, want 10", len(recoveryCodes))
	}
	status, err := fixture.service.TwofaStatus(ctx, owner)
	if err != nil || !status.Enabled || status.RecoveryCodesRemaining != 10 {
		t.Fatalf("twofa status = %#v (%v)", status, err)
	}
	if _, err := fixture.service.StartTwofaEnrollment(ctx, owner, ownerPassword); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("re-enrollment while enabled = %v", err)
	}

	outcome, err := fixture.service.SignIn(ctx, "+77000000001", ownerPassword,
		core.SessionClientInfo{DeviceLabel: "iPhone 17", Platform: "ios"})
	if err != nil {
		t.Fatalf("sign in with 2FA: %v", err)
	}
	if outcome.Tokens != nil || outcome.TwofaChallenge == "" || outcome.TwofaExpiresAt == nil {
		t.Fatalf("sign-in outcome with 2FA = %#v", outcome)
	}
	if _, err := fixture.service.SignInWithTwofa(ctx, outcome.TwofaChallenge, "999999"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("second factor with wrong code = %v", err)
	}
	tokens, err := fixture.service.SignInWithTwofa(ctx, outcome.TwofaChallenge, totpNow(t, fixture, enrollment.Secret))
	if err != nil {
		t.Fatalf("complete second factor: %v", err)
	}
	if _, err := fixture.service.SignInWithTwofa(ctx, outcome.TwofaChallenge, totpNow(t, fixture, enrollment.Secret)); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("challenge replay = %v", err)
	}
	principal, err := fixture.service.Authenticate(ctx, tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate after second factor: %v", err)
	}
	devices, err := fixture.service.ListSessions(ctx, principal)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	var found bool
	for _, device := range devices {
		if device.Current {
			found = true
			if device.DeviceLabel != "iPhone 17" || device.Platform != "ios" {
				t.Fatalf("challenge device metadata lost: %#v", device)
			}
		}
	}
	if !found {
		t.Fatal("current session missing after second factor")
	}

	recoveryOutcome, err := fixture.service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{})
	if err != nil || recoveryOutcome.TwofaChallenge == "" {
		t.Fatalf("second sign-in = %#v (%v)", recoveryOutcome, err)
	}
	if _, err := fixture.service.SignInWithTwofa(ctx, recoveryOutcome.TwofaChallenge, recoveryCodes[0]); err != nil {
		t.Fatalf("recovery code sign-in: %v", err)
	}
	status, err = fixture.service.TwofaStatus(ctx, owner)
	if err != nil || status.RecoveryCodesRemaining != 9 {
		t.Fatalf("recovery codes remaining = %#v (%v)", status, err)
	}
	reusedOutcome, err := fixture.service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{})
	if err != nil {
		t.Fatalf("third sign-in: %v", err)
	}
	if _, err := fixture.service.SignInWithTwofa(ctx, reusedOutcome.TwofaChallenge, recoveryCodes[0]); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("reused recovery code = %v", err)
	}

	if err := fixture.service.DisableTwofa(ctx, owner, ownerPassword, totpNow(t, fixture, enrollment.Secret)); err != nil {
		t.Fatalf("disable twofa: %v", err)
	}
	status, err = fixture.service.TwofaStatus(ctx, owner)
	if err != nil || status.Enabled || status.RecoveryCodesRemaining != 0 {
		t.Fatalf("status after disable = %#v (%v)", status, err)
	}
	plain, err := fixture.service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{})
	if err != nil || plain.Tokens == nil {
		t.Fatalf("sign-in after disable = %#v (%v)", plain, err)
	}
}

func TestChallengeAttemptsExhaust(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	owner := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)
	enrollment, err := fixture.service.StartTwofaEnrollment(ctx, owner, ownerPassword)
	if err != nil {
		t.Fatalf("start enrollment: %v", err)
	}
	if _, err := fixture.service.ConfirmTwofaEnrollment(ctx, owner, totpNow(t, fixture, enrollment.Secret)); err != nil {
		t.Fatalf("confirm enrollment: %v", err)
	}
	outcome, err := fixture.service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{})
	if err != nil {
		t.Fatalf("sign in: %v", err)
	}
	for attempt := 0; attempt < 5; attempt++ {
		if _, err := fixture.service.SignInWithTwofa(ctx, outcome.TwofaChallenge, "000000"); !core.IsCode(err, core.CodeUnauthenticated) {
			t.Fatalf("attempt %d = %v", attempt, err)
		}
	}
	if _, err := fixture.service.SignInWithTwofa(ctx, outcome.TwofaChallenge, totpNow(t, fixture, enrollment.Secret)); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("burned challenge with valid code = %v", err)
	}
	page, err := fixture.service.ListSecurityEvents(ctx, owner, "", 50)
	if err != nil {
		t.Fatalf("security events: %v", err)
	}
	var burned bool
	for _, event := range page.Events {
		if event.Action == "TwofaChallengeFailed" && event.Decision == "deny" {
			burned = true
		}
	}
	if !burned {
		t.Fatal("exhausted challenge must appear in the security feed")
	}
}

func TestActivationMultiStepJourney(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	link, _, err := fixture.service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: "tenant_belcanto", OwnerAccountID: fixture.owner.AccountID,
		FullName: "Шугыла Замещающая", Phone: "+77000000777", Role: core.RoleTeacher,
		Operator: "test-operator", Reason: "multi-step activation journey",
	})
	if err != nil {
		t.Fatalf("bootstrap journey teacher: %v", err)
	}
	token := tokenFromLink(t, link)

	progress, err := fixture.service.ActivationProgress(ctx, token)
	if err != nil {
		t.Fatalf("initial progress: %v", err)
	}
	if progress.PasswordSet || progress.ContactVerified || progress.Completed || progress.DisplayName != "Шугыла Замещающая" {
		t.Fatalf("initial progress = %#v", progress)
	}

	if err := fixture.service.StartActivationContact(ctx, token, core.ContactEmail, "shugyla@example.kz"); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("contact before password = %v", err)
	}
	if err := fixture.service.SetActivationPassword(ctx, token, "+77000000777", "short"); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("weak activation password = %v", err)
	}
	if err := fixture.service.SetActivationPassword(ctx, token, "+77000000777", teacherPassword); err != nil {
		t.Fatalf("set activation password: %v", err)
	}
	if err := fixture.service.StartActivationContact(ctx, token, core.ContactEmail, "not-an-email"); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("invalid email = %v", err)
	}
	if err := fixture.service.StartActivationContact(ctx, token, core.ContactEmail, "Shugyla@Example.kz"); err != nil {
		t.Fatalf("start activation contact: %v", err)
	}
	verificationID := lastOutboxAggregate(t, fixture, "ContactVerificationRequested")
	if err := fixture.service.VerifyActivationContact(ctx, token, "000000"); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("wrong contact code = %v", err)
	}
	if err := fixture.service.VerifyActivationContact(ctx, token, fixture.codec.ContactVerificationCode(verificationID)); err != nil {
		t.Fatalf("verify activation contact: %v", err)
	}
	progress, err = fixture.service.ActivationProgress(ctx, token)
	if err != nil || !progress.PasswordSet || !progress.ContactVerified || progress.ContactKind != core.ContactEmail {
		t.Fatalf("progress after contact = %#v (%v)", progress, err)
	}
	if !strings.Contains(progress.ContactMasked, "@example.kz") || strings.Contains(progress.ContactMasked, "shugyla@") {
		t.Fatalf("contact mask leaks the address: %q", progress.ContactMasked)
	}

	start, err := fixture.service.StartActivationTwofa(ctx, token)
	if err != nil {
		t.Fatalf("start activation twofa: %v", err)
	}
	recovery, err := fixture.service.ConfirmActivationTwofa(ctx, token, totpNow(t, fixture, start.Secret))
	if err != nil || len(recovery) != 10 {
		t.Fatalf("confirm activation twofa = %d codes (%v)", len(recovery), err)
	}

	if err := fixture.service.FinishActivation(ctx, token, "+77000000777", "finish-journey"); err != nil {
		t.Fatalf("finish activation: %v", err)
	}
	if err := fixture.service.FinishActivation(ctx, token, "+77000000777", "finish-journey"); err != nil {
		t.Fatalf("idempotent finish replay: %v", err)
	}

	outcome, err := fixture.service.SignIn(ctx, "+77000000777", teacherPassword, core.SessionClientInfo{})
	if err != nil || outcome.TwofaChallenge == "" {
		t.Fatalf("post-activation sign-in = %#v (%v)", outcome, err)
	}
	tokens, err := fixture.service.SignInWithTwofa(ctx, outcome.TwofaChallenge, totpNow(t, fixture, start.Secret))
	if err != nil {
		t.Fatalf("post-activation second factor: %v", err)
	}
	principal, err := fixture.service.Authenticate(ctx, tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate journey teacher: %v", err)
	}
	contacts, err := fixture.service.ListVerifiedContacts(ctx, principal)
	if err != nil || len(contacts) != 1 || contacts[0].Value != "shugyla@example.kz" {
		t.Fatalf("verified contacts = %#v (%v)", contacts, err)
	}
}

func TestContactChangeRequiresReauthAndCode(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	owner := signInPrincipal(t, fixture.service, "+77000000001", ownerPassword)

	if err := fixture.service.StartContactChange(ctx, owner, "wrong-password", core.ContactEmail, "owner@belcanto.kz"); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("contact change without re-auth = %v", err)
	}
	if err := fixture.service.StartContactChange(ctx, owner, ownerPassword, core.ContactEmail, "owner@belcanto.kz"); err != nil {
		t.Fatalf("start contact change: %v", err)
	}
	verificationID := lastOutboxAggregate(t, fixture, "ContactVerificationRequested")
	for attempt := 0; attempt < 5; attempt++ {
		if _, err := fixture.service.ConfirmContactChange(ctx, owner, "111111"); !core.IsCode(err, core.CodeInvalidInput) {
			t.Fatalf("wrong code attempt %d = %v", attempt, err)
		}
	}
	if _, err := fixture.service.ConfirmContactChange(ctx, owner, fixture.codec.ContactVerificationCode(verificationID)); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("burned verification with valid code = %v", err)
	}
	if err := fixture.service.StartContactChange(ctx, owner, ownerPassword, core.ContactEmail, "owner@belcanto.kz"); err != nil {
		t.Fatalf("restart contact change: %v", err)
	}
	verificationID = lastOutboxAggregate(t, fixture, "ContactVerificationRequested")
	contact, err := fixture.service.ConfirmContactChange(ctx, owner, fixture.codec.ContactVerificationCode(verificationID))
	if err != nil || contact.Value != "owner@belcanto.kz" || contact.Kind != core.ContactEmail {
		t.Fatalf("confirmed contact = %#v (%v)", contact, err)
	}
	contacts, err := fixture.service.ListVerifiedContacts(ctx, owner)
	if err != nil || len(contacts) != 1 {
		t.Fatalf("verified contacts = %#v (%v)", contacts, err)
	}
	page, err := fixture.service.ListSecurityEvents(ctx, owner, "", 50)
	if err != nil {
		t.Fatalf("security events: %v", err)
	}
	seen := map[string]bool{}
	for _, event := range page.Events {
		seen[event.Action] = true
	}
	if !seen["ContactChangeStarted"] || !seen["ContactVerified"] {
		t.Fatalf("contact events missing from feed: %#v", seen)
	}
}
