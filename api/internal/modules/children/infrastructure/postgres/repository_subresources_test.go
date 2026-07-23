package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	childdomain "nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/dbtest"
)

func newChildID(t *testing.T, pool *pgxpool.Pool, tenantID, branchID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO children (id, tenant_id, branch_id, first_name, date_of_birth, start_date, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, true)`,
		id, tenantID, branchID, name, dbtest.DateAt(2021, 1, 1), dbtest.DateAt(2024, 9, 1)); err != nil {
		t.Fatalf("insert child: %v", err)
	}
	return id
}

func ensureManagerMembership(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.MustParse("50000000-0000-0000-0000-000000000001")
	_, err := pool.Exec(context.Background(),
		`INSERT INTO memberships (id, tenant_id, branch_id, user_id, role, is_active)
		 VALUES ($1, $2, $3, $4, 'manager', true)
		 ON CONFLICT (id) DO NOTHING`,
		id, childTenantID, childBranchID, childUserID)
	if err != nil {
		t.Fatalf("insert membership: %v", err)
	}
	return id
}

// ===========================================================================
// child_profiles
// ===========================================================================

func TestChildProfile_InsertAndGet(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Profile Kid")
	now := time.Now().UTC().Truncate(time.Microsecond)
	sex := "F"
	regDate := dbtest.DateAt(2024, 8, 1)
	addrLine1 := "1 Nursery Lane"
	addrCity := "London"
	p := &childdomain.ChildProfile{
		ID:                       uuid.New(),
		TenantID:                 childTenantID,
		BranchID:                 childBranchID,
		ChildID:                  childID,
		Sex:                      &sex,
		AddressLine1:             &addrLine1,
		AddressCity:              &addrCity,
		RegistrationDate:         &regDate,
		DemographicsHomeReviewed: true,
		GDPRDeclaredAt:           &now,
	}

	tx := dbtest.BeginTx(t, pool)
	created, err := repo.InsertProfile(ctx, tx, p)
	if err != nil {
		t.Fatalf("InsertProfile: %v", err)
	}
	dbtest.CommitTx(t, tx)

	if created.ID != p.ID {
		t.Errorf("created.ID = %s, want %s", created.ID, p.ID)
	}

	got, err := repo.GetProfileByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetProfileByChild: %v", err)
	}
	if got == nil {
		t.Fatal("got = nil")
	}
	if got.ChildID != childID {
		t.Errorf("ChildID = %s, want %s", got.ChildID, childID)
	}
	if got.Sex == nil || *got.Sex != "F" {
		t.Errorf("Sex = %v, want F", got.Sex)
	}
	if got.AddressCity == nil || *got.AddressCity != "London" {
		t.Errorf("AddressCity = %v, want London", got.AddressCity)
	}
	if !got.DemographicsHomeReviewed {
		t.Error("DemographicsHomeReviewed = false, want true")
	}
}

func TestChildProfile_Update(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Update Kid")
	p := &childdomain.ChildProfile{
		ID:               uuid.New(),
		TenantID:         childTenantID,
		BranchID:         childBranchID,
		ChildID:          childID,
		DisabilityStatus: childdomain.YesNoUnknownUnknown,
	}
	tx := dbtest.BeginTx(t, pool)
	if _, err := repo.InsertProfile(ctx, tx, p); err != nil {
		t.Fatalf("InsertProfile: %v", err)
	}
	dbtest.CommitTx(t, tx)

	p.DisabilityStatus = childdomain.YesNoUnknownYes
	note := "wheelchair access"
	p.AccessRequirements = &note

	tx = dbtest.BeginTx(t, pool)
	updated, err := repo.UpdateProfile(ctx, tx, p)
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	dbtest.CommitTx(t, tx)
	if updated.DisabilityStatus != childdomain.YesNoUnknownYes {
		t.Errorf("DisabilityStatus = %s, want yes", updated.DisabilityStatus)
	}
	if updated.AccessRequirements == nil || *updated.AccessRequirements != "wheelchair access" {
		t.Errorf("AccessRequirements = %v, want wheelchair access", updated.AccessRequirements)
	}
}

func TestChildProfile_GetForUpdate_NotFound(t *testing.T) {
	repo, _ := setupChildRepo(t)
	ctx := context.Background()

	tx := dbtest.BeginTx(t, dbtest.RequirePostgres(t))
	defer dbtest.CommitTx(t, tx)
	got, err := repo.GetProfileForUpdate(ctx, tx, childTenantID, childBranchID, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("got = %+v, want nil for missing profile", got)
	}
}

// ===========================================================================
// child_contacts
//
// Note: ReplaceContactsForTypes is exercised by the application layer's
// ReplaceContacts handler test; the per-call smoke test is omitted here
// because the sqlc-generated enum-array param requires a registered codec
// that lives outside the children module.
// ===========================================================================

// ===========================================================================
// child_health_profiles
// ===========================================================================

func TestChildHealth_Upsert(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Health Kid")
	notes := "peanut allergy"
	h := &childdomain.ChildHealthProfile{
		ID:                         uuid.New(),
		TenantID:                   childTenantID,
		BranchID:                   childBranchID,
		ChildID:                    childID,
		MedicalConditionsStatus:    childdomain.YesNoUnknownYes,
		MedicalConditionsNotes:     &notes,
		PrescribedMedicationStatus: childdomain.YesNoUnknownNo,
		ImmunisationStatus:         childdomain.ImmunisationUpToDate,
		DietaryRequirementsStatus:  childdomain.YesNoUnknownYes,
		DoctorName:                 dbtest.StrPtr("Dr Who"),
	}

	tx := dbtest.BeginTx(t, pool)
	if _, err := repo.UpsertHealth(ctx, tx, h); err != nil {
		t.Fatalf("UpsertHealth: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, err := repo.GetHealthByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetHealthByChild: %v", err)
	}
	if got == nil {
		t.Fatal("got = nil")
	}
	if got.MedicalConditionsStatus != childdomain.YesNoUnknownYes {
		t.Errorf("MedicalConditionsStatus = %s, want yes", got.MedicalConditionsStatus)
	}
	if got.ImmunisationStatus != childdomain.ImmunisationUpToDate {
		t.Errorf("ImmunisationStatus = %s, want up_to_date", got.ImmunisationStatus)
	}
	if got.DoctorName == nil || *got.DoctorName != "Dr Who" {
		t.Errorf("DoctorName = %v, want Dr Who", got.DoctorName)
	}

	h.MedicalConditionsStatus = childdomain.YesNoUnknownNo
	tx = dbtest.BeginTx(t, pool)
	if _, err := repo.UpsertHealth(ctx, tx, h); err != nil {
		t.Fatalf("UpsertHealth update: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, err = repo.GetHealthByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetHealthByChild (2): %v", err)
	}
	if got.MedicalConditionsStatus != childdomain.YesNoUnknownNo {
		t.Errorf("after update MedicalConditionsStatus = %s, want no", got.MedicalConditionsStatus)
	}
	if got.DoctorName == nil || *got.DoctorName != "Dr Who" {
		t.Errorf("DoctorName lost on update: %v", got.DoctorName)
	}
}

// ===========================================================================
// child_safeguarding_profiles
// ===========================================================================

func TestChildSafeguarding_UpsertWithReferrals(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Safe Kid")
	referralType := "speech_and_language_therapist"
	date := "2024-01-15"
	s := &childdomain.ChildSafeguardingProfile{
		ID:                    uuid.New(),
		TenantID:              childTenantID,
		BranchID:              childBranchID,
		ChildID:               childID,
		SocialServicesStatus:  childdomain.YesNoUnknownNo,
		ConcernWalking:        childdomain.YesNoUnknownYes,
		ConcernSpeechLanguage: childdomain.YesNoUnknownYes,
		ProfessionalReferrals: []childdomain.ProfessionalReferral{
			{Type: referralType, ReferredDate: &date},
		},
	}

	tx := dbtest.BeginTx(t, pool)
	if _, err := repo.UpsertSafeguarding(ctx, tx, s); err != nil {
		t.Fatalf("UpsertSafeguarding: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, err := repo.GetSafeguardingByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetSafeguardingByChild: %v", err)
	}
	if got == nil {
		t.Fatal("got = nil")
	}
	if got.ConcernWalking != childdomain.YesNoUnknownYes {
		t.Errorf("ConcernWalking = %s, want yes", got.ConcernWalking)
	}
	if len(got.ProfessionalReferrals) != 1 {
		t.Errorf("len(ProfessionalReferrals) = %d, want 1", len(got.ProfessionalReferrals))
	}
	if got.ProfessionalReferrals[0].Type != referralType {
		t.Errorf("Referrals[0].Type = %s, want %s", got.ProfessionalReferrals[0].Type, referralType)
	}
}

// ===========================================================================
// child_consent_records
// ===========================================================================

func TestChildConsent_InsertThenUpdate(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Consent Kid")
	membershipID := ensureManagerMembership(t, pool)
	signed := dbtest.DateAt(2024, 9, 15)
	c := &childdomain.ChildConsent{
		ID:                                   uuid.New(),
		TenantID:                             childTenantID,
		BranchID:                             childBranchID,
		ChildID:                              childID,
		UrgentMedicalTreatment:               true,
		Plasters:                             true,
		SafeguardingReportingAcknowledgement: true,
		InformationSharingConsent:            true,
		GDPRDataProcessingConsent:            true,
		AreaSENCOLiaison:                     false,
		HealthVisitorLiaison:                 false,
		TransitionDocuments:                  true,
		LocalOutings:                         true,
		FacePainting:                         true,
		ParentSuppliedSunCream:               true,
		ParentSuppliedNappyCream:             true,
		DevelopmentProfilePhotos:             true,
		NurseryDisplayBoards:                 true,
		PromotionalLiterature:                false,
		NurseryWebsite:                       false,
		StaffStudentCoursework:               false,
		SocialMedia:                          false,
		SignerName:                           "Parent",
		SignedDate:                           signed,
		PaperFormOnFile:                      true,
		EnteredByUserID:                      childUserID,
		EnteredByMembershipID:                membershipID,
	}

	tx := dbtest.BeginTx(t, pool)
	if _, err := repo.InsertConsent(ctx, tx, c); err != nil {
		t.Fatalf("InsertConsent: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, found, err := repo.GetConsentByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetConsentByChild: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if !got.Plasters {
		t.Error("Plasters = false, want true")
	}
	if got.SignerName != "Parent" {
		t.Errorf("SignerName = %s, want Parent", got.SignerName)
	}

	got.PromotionalLiterature = true
	got.NurseryWebsite = true
	tx = dbtest.BeginTx(t, pool)
	if _, err := repo.UpdateConsent(ctx, tx, got); err != nil {
		t.Fatalf("UpdateConsent: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got2, _, err := repo.GetConsentByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetConsentByChild (2): %v", err)
	}
	if !got2.PromotionalLiterature {
		t.Error("PromotionalLiterature = false, want true after update")
	}
}

// ===========================================================================
// child_collection_settings
// ===========================================================================

func TestChildCollectionSettings_SetPassword(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Collection Kid")
	membershipID := ensureManagerMembership(t, pool)

	tx := dbtest.BeginTx(t, pool)
	cs := &childdomain.ChildCollectionSetting{
		ID:                           uuid.New(),
		TenantID:                     childTenantID,
		BranchID:                     childBranchID,
		ChildID:                      childID,
		Over18CollectionAcknowledged: true,
	}
	created, err := repo.UpsertCollectionSetting(ctx, tx, cs)
	if err != nil {
		t.Fatalf("UpsertCollectionSetting: %v", err)
	}
	plainPassword := "super-secure-collection-123"
	if err := repo.SetCollectionPassword(ctx, tx, childTenantID, childBranchID, childID, created.ID,
		plainPassword, "hint-text", time.Now().UTC(), childUserID, membershipID); err != nil {
		t.Fatalf("SetCollectionPassword: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, err := repo.GetCollectionSettingByChild(ctx, nil, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetCollectionSettingByChild: %v", err)
	}
	if got == nil {
		t.Fatal("got = nil after set")
	}
	if got.CollectionPassword != plainPassword {
		t.Errorf("got.CollectionPassword = %q, want %q", got.CollectionPassword, plainPassword)
	}
	if got.CollectionPasswordUpdatedAt == nil {
		t.Error("CollectionPasswordUpdatedAt nil after set")
	}
}

// ===========================================================================
// child_room_assignments
// ===========================================================================

func TestChildRoomAssignment_InsertAndOnlyOneCurrent(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Room Kid")
	room1ID := uuid.New()
	room2ID := uuid.New()
	dbtest.InsertRoom(t, pool, room1ID, childTenantID, childBranchID, "Babies", 10)
	dbtest.InsertRoom(t, pool, room2ID, childTenantID, childBranchID, "Toddlers", 10)

	tx := dbtest.BeginTx(t, pool)
	a1, err := repo.InsertRoomAssignment(ctx, tx, &childdomain.ChildRoomAssignment{
		ID: uuid.New(), TenantID: childTenantID, BranchID: childBranchID,
		ChildID: childID, RoomID: room1ID, StartDate: dbtest.DateAt(2024, 9, 1),
	})
	if err != nil {
		t.Fatalf("InsertRoomAssignment #1: %v", err)
	}
	dbtest.CommitTx(t, tx)
	if !a1.IsCurrent {
		t.Error("a1.IsCurrent = false, want true")
	}

	tx = dbtest.BeginTx(t, pool)
	if err := repo.CloseCurrentRoomAssignment(ctx, tx, childTenantID, childBranchID, childID, dbtest.DateAt(2025, 1, 15)); err != nil {
		t.Fatalf("CloseCurrentRoomAssignment: %v", err)
	}
	a2, err := repo.InsertRoomAssignment(ctx, tx, &childdomain.ChildRoomAssignment{
		ID: uuid.New(), TenantID: childTenantID, BranchID: childBranchID,
		ChildID: childID, RoomID: room2ID, StartDate: dbtest.DateAt(2025, 1, 16),
	})
	if err != nil {
		t.Fatalf("InsertRoomAssignment #2: %v", err)
	}
	dbtest.CommitTx(t, tx)
	if !a2.IsCurrent {
		t.Error("a2.IsCurrent = false, want true")
	}

	all, err := repo.ListRoomAssignmentsByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("ListRoomAssignmentsByChild: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("len(all) = %d, want 2", len(all))
	}
	currentCount := 0
	for _, a := range all {
		if a.IsCurrent {
			currentCount++
			if a.RoomID != room2ID {
				t.Errorf("current room = %s, want %s", a.RoomID, room2ID)
			}
		}
	}
	if currentCount != 1 {
		t.Errorf("currentCount = %d, want 1 (only-one-current invariant)", currentCount)
	}

	cur, found, err := repo.GetCurrentRoomAssignmentByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetCurrentRoomAssignmentByChild: %v", err)
	}
	if !found || cur.RoomID != room2ID {
		t.Errorf("current = %+v, want room2", cur)
	}
}

func TestChildRoomAssignment_CloseByID(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "CloseID Kid")
	roomID := uuid.New()
	dbtest.InsertRoom(t, pool, roomID, childTenantID, childBranchID, "Pre-school", 10)

	tx := dbtest.BeginTx(t, pool)
	a, err := repo.InsertRoomAssignment(ctx, tx, &childdomain.ChildRoomAssignment{
		ID: uuid.New(), TenantID: childTenantID, BranchID: childBranchID,
		ChildID: childID, RoomID: roomID, StartDate: dbtest.DateAt(2024, 9, 1),
	})
	if err != nil {
		t.Fatalf("InsertRoomAssignment: %v", err)
	}
	dbtest.CommitTx(t, tx)

	tx = dbtest.BeginTx(t, pool)
	closed, err := repo.CloseRoomAssignmentByID(ctx, tx, childTenantID, childBranchID, a.ID, dbtest.DateAt(2025, 6, 1))
	if err != nil {
		t.Fatalf("CloseRoomAssignmentByID: %v", err)
	}
	if !closed {
		t.Error("closed = false, want true")
	}
	dbtest.CommitTx(t, tx)

	all, err := repo.ListRoomAssignmentsByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if all[0].IsCurrent {
		t.Error("all[0].IsCurrent = true, want false after close")
	}
	if all[0].EndDate == nil {
		t.Error("all[0].EndDate = nil, want 2025-06-01")
	}
}

// ===========================================================================
// child_billing_profiles
// ===========================================================================

func TestChildBillingProfile_UpsertAndConstraint(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Billing Kid")
	bp := &childdomain.ChildBillingProfile{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		BillingBasis:  "site_rate",
		EffectiveFrom: dbtest.DateAt(2024, 9, 1),
	}

	tx := dbtest.BeginTx(t, pool)
	if _, err := repo.UpsertBillingProfile(ctx, tx, bp); err != nil {
		t.Fatalf("UpsertBillingProfile site_rate: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, found, err := repo.GetBillingProfileByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetBillingProfileByChild: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if got.BillingBasis != "site_rate" {
		t.Errorf("BillingBasis = %s, want site_rate", got.BillingBasis)
	}
	if got.CustomRateMinor != nil {
		t.Errorf("CustomRateMinor = %v, want nil for site_rate", got.CustomRateMinor)
	}

	custom := 550
	bp.BillingBasis = "custom"
	bp.CustomRateMinor = &custom
	tx = dbtest.BeginTx(t, pool)
	if _, err := repo.UpsertBillingProfile(ctx, tx, bp); err != nil {
		t.Fatalf("UpsertBillingProfile custom: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, _, err = repo.GetBillingProfileByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetBillingProfileByChild (2): %v", err)
	}
	if got.BillingBasis != "custom" {
		t.Errorf("BillingBasis = %s, want custom", got.BillingBasis)
	}
	if got.CustomRateMinor == nil || *got.CustomRateMinor != 550 {
		t.Errorf("CustomRateMinor = %v, want 550", got.CustomRateMinor)
	}

	// The DB-level consistency check rejects custom without a rate.
	badChildID := newChildID(t, pool, childTenantID, childBranchID, "Bad Billing Kid")
	bad := &childdomain.ChildBillingProfile{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       badChildID,
		BillingBasis:  "custom",
		EffectiveFrom: dbtest.DateAt(2024, 9, 1),
	}
	tx = dbtest.BeginTx(t, pool)
	if _, err := repo.UpsertBillingProfile(ctx, tx, bad); err == nil {
		_ = tx.Rollback(context.Background())
		t.Fatal("expected error for custom without custom_rate_minor")
	}
	_ = tx.Rollback(context.Background())
}

// ===========================================================================
// child_leaving_records
// ===========================================================================

func TestChildLeavingRecord_InsertAndGet(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Leaving Kid")
	note := "moved to another city"
	p := &childdomain.ChildLeavingRecord{
		ID:         uuid.New(),
		TenantID:   childTenantID,
		BranchID:   childBranchID,
		ChildID:    childID,
		LeftAt:     time.Now().UTC(),
		ReasonCode: "left_nursery",
		ReasonNote: &note,
	}

	tx := dbtest.BeginTx(t, pool)
	if err := repo.InsertLeavingRecord(ctx, tx, p); err != nil {
		t.Fatalf("InsertLeavingRecord: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, found, err := repo.GetLeavingRecordByChild(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetLeavingRecordByChild: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if got.ReasonCode != "left_nursery" {
		t.Errorf("ReasonCode = %s, want left_nursery", got.ReasonCode)
	}
	if got.ReasonNote == nil || *got.ReasonNote != "moved to another city" {
		t.Errorf("ReasonNote = %v, want 'moved to another city'", got.ReasonNote)
	}
}

func TestChildLeavingRecord_InvalidReason(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newChildID(t, pool, childTenantID, childBranchID, "Bad Reason Kid")
	p := &childdomain.ChildLeavingRecord{
		ID:         uuid.New(),
		TenantID:   childTenantID,
		BranchID:   childBranchID,
		ChildID:    childID,
		LeftAt:     time.Now().UTC(),
		ReasonCode: "totally_made_up_reason",
	}
	tx := dbtest.BeginTx(t, pool)
	err := repo.InsertLeavingRecord(ctx, tx, p)
	if err == nil {
		_ = tx.Rollback(context.Background())
		t.Fatal("expected error for invalid reason_code, got nil")
	}
	_ = tx.Rollback(context.Background())
}
