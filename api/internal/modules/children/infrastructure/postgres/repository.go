package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type ChildRepository struct {
	pool *pgxpool.Pool
}

func NewChildRepository(pool *pgxpool.Pool) *ChildRepository {
	return &ChildRepository{pool: pool}
}

func (r *ChildRepository) List(ctx context.Context, tenantID, branchID uuid.UUID, filter domain.StatusFilter, limit, offset int) ([]domain.Child, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ChildrenList(ctx, sqlc.ChildrenListParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		Limit:        int32(limit),
		Offset:       int32(offset),
		StatusFilter: string(filter),
	})
	if err != nil {
		return nil, fmt.Errorf("query children: %w", err)
	}
	out := make([]domain.Child, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapChildRow(row))
	}
	return out, nil
}

func (r *ChildRepository) Count(ctx context.Context, tenantID, branchID uuid.UUID, filter domain.StatusFilter) (int, error) {
	q := sqlc.New(r.pool)
	n, err := q.ChildrenCount(ctx, sqlc.ChildrenCountParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		StatusFilter: string(filter),
	})
	if err != nil {
		return 0, fmt.Errorf("count children: %w", err)
	}
	return int(n), nil
}

func (r *ChildRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildrenGetByID(ctx, sqlc.ChildrenGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Child{}, false, nil
	}
	if err != nil {
		return domain.Child{}, false, fmt.Errorf("query child by id: %w", err)
	}
	return mapChildRow(row), true, nil
}

func (r *ChildRepository) Create(ctx context.Context, child domain.Child, notes string, tenantID, branchID uuid.UUID) error {
	q := sqlc.New(r.pool)
	return q.ChildrenCreate(ctx, sqlc.ChildrenCreateParams{
		ID:          uuidToPgtype(child.ID),
		TenantID:    uuidToPgtype(tenantID),
		BranchID:    uuidToPgtype(branchID),
		FirstName:   child.FirstName,
		Column5:     stringPtrToInterface(child.MiddleName),
		Column6:     stringPtrToInterface(child.LastName),
		DateOfBirth: timeToPgtypeDate(child.DateOfBirth),
		StartDate:   timeToPgtypeDate(child.StartDate),
		EndDate:     timeToPgtypeDatePtr(child.EndDate),
		Column10:    notes,
		IsActive:    child.IsActive,
	})
}

func (r *ChildRepository) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	params := sqlc.ChildrenUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	}

	if v, ok := fields["first_name"]; ok {
		params.SetFirstName = int32(1)
		params.FirstName = v.(string)
	}
	if v, ok := fields["middle_name"]; ok {
		params.SetMiddleName = int32(1)
		params.MiddleName = v.(string)
	}
	if v, ok := fields["last_name"]; ok {
		params.SetLastName = int32(1)
		params.LastName = v.(string)
	}
	if v, ok := fields["date_of_birth"]; ok {
		params.SetDateOfBirth = int32(1)
		params.DateOfBirth = timeToPgtypeDate(v.(time.Time))
	}
	if v, ok := fields["start_date"]; ok {
		params.SetStartDate = int32(1)
		params.StartDate = timeToPgtypeDate(v.(time.Time))
	}
	if v, ok := fields["end_date"]; ok {
		params.SetEndDate = int32(1)
		if t, ok := v.(time.Time); ok {
			params.EndDate = timeToPgtypeDate(t)
		}
	}
	if v, ok := fields["notes"]; ok {
		params.SetNotes = int32(1)
		params.Notes = v.(string)
	}

	q := sqlc.New(r.pool)
	ct, err := q.ChildrenUpdate(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("update child: %w", err)
	}
	return ct, nil
}

func (r *ChildRepository) MarkInactive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx)
	return q.ChildrenMarkInactive(ctx, sqlc.ChildrenMarkInactiveParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *ChildRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ChildrenGetByIDForUpdate(ctx, sqlc.ChildrenGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Child{}, false, nil
	}
	if err != nil {
		return domain.Child{}, false, fmt.Errorf("query child for update: %w", err)
	}
	return mapChildRow(row), true, nil
}

func (r *ChildRepository) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	q := sqlc.New(tx)
	exists, err := q.ChildrenExistsInScope(ctx, sqlc.ChildrenExistsInScopeParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if err != nil {
		return false, fmt.Errorf("check child exists in scope: %w", err)
	}
	return exists, nil
}

func (r *ChildRepository) ListAttendance(ctx context.Context, tenantID, branchID uuid.UUID, localDate time.Time) ([]domain.AttendanceChild, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ChildrenListAttendance(ctx, sqlc.ChildrenListAttendanceParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		LocalDate: timeToPgtypeDate(localDate),
	})
	if err != nil {
		return nil, fmt.Errorf("query attendance children: %w", err)
	}
	out := make([]domain.AttendanceChild, 0, len(rows))
	for _, row := range rows {
		var hasIncomplete bool
		if v, ok := row.HasIncompleteSession.(bool); ok {
			hasIncomplete = v
		}
		out = append(out, domain.AttendanceChild{
			ID:                   pgtypeUUIDToUUID(row.ID),
			FirstName:            row.FirstName,
			MiddleName:           pgtypeTextToStringPtr(row.MiddleName),
			LastName:             pgtypeTextToStringPtr(row.LastName),
			EnrollmentComplete:   row.EnrollmentComplete.Bool,
			AttendanceState:      row.AttendanceState,
			OpenSessionID:        pgtypeUUIDToUUIDPtr(row.OpenSessionID),
			CheckedInAt:          pgtypeTimestamptzToTimePtr(row.CheckedInAt),
			HasIncompleteSession: hasIncomplete,
			AbsenceMarkerID:      pgtypeUUIDToUUIDPtr(row.AbsenceMarkerID),
			AbsenceMarkedAt:      pgtypeTimestamptzToTimePtr(row.AbsenceMarkedAt),
		})
	}
	return out, nil
}

func (r *ChildRepository) GetForAttendanceCheck(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (domain.Child, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ChildrenGetByID(ctx, sqlc.ChildrenGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Child{}, false, nil
	}
	if err != nil {
		return domain.Child{}, false, fmt.Errorf("get child for attendance check: %w", err)
	}
	return mapChildRow(row), true, nil
}

func (r *ChildRepository) GetChildForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (domain.ChildCorrectionInfo, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ChildrenGetForCorrection(ctx, sqlc.ChildrenGetForCorrectionParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ChildCorrectionInfo{}, false, nil
	}
	if err != nil {
		return domain.ChildCorrectionInfo{}, false, fmt.Errorf("get child for correction: %w", err)
	}
	return domain.ChildCorrectionInfo{
		ID:        pgtypeUUIDToUUID(row.ID),
		StartDate: pgtypeDateToTime(row.StartDate),
		EndDate:   pgtypeDateToTimePtr(row.EndDate),
	}, true, nil
}

func mapChildRow(row interface{}) domain.Child {
	type fields struct {
		ID                      pgtype.UUID
		FirstName               string
		MiddleName              pgtype.Text
		LastName                pgtype.Text
		DateOfBirth             pgtype.Date
		StartDate               pgtype.Date
		EndDate                 pgtype.Date
		SiteCoreHourlyRateMinor pgtype.Int4
		Notes                   pgtype.Text
		IsActive                bool
		HasCurrentRoom          bool
		HasParentCarerContact   bool
		HasBookingPattern       bool
		CreatedAt               pgtype.Timestamptz
		UpdatedAt               pgtype.Timestamptz
	}
	var f fields
	switch v := row.(type) {
	case sqlc.ChildrenListRow:
		f = fields{
			ID: v.ID, FirstName: v.FirstName, MiddleName: v.MiddleName, LastName: v.LastName,
			DateOfBirth: v.DateOfBirth, StartDate: v.StartDate, EndDate: v.EndDate,
			SiteCoreHourlyRateMinor: v.SiteCoreHourlyRateMinor, Notes: v.Notes,
			IsActive: v.IsActive, HasCurrentRoom: v.HasCurrentRoom, HasParentCarerContact: v.HasParentCarerContact,
			HasBookingPattern: v.HasBookingPattern,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
	case sqlc.ChildrenGetByIDRow:
		f = fields{
			ID: v.ID, FirstName: v.FirstName, MiddleName: v.MiddleName, LastName: v.LastName,
			DateOfBirth: v.DateOfBirth, StartDate: v.StartDate, EndDate: v.EndDate,
			SiteCoreHourlyRateMinor: v.SiteCoreHourlyRateMinor, Notes: v.Notes,
			IsActive: v.IsActive, HasCurrentRoom: v.HasCurrentRoom, HasParentCarerContact: v.HasParentCarerContact,
			HasBookingPattern: v.HasBookingPattern,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
	case sqlc.ChildrenGetByIDForUpdateRow:
		f = fields{
			ID: v.ID, FirstName: v.FirstName, MiddleName: v.MiddleName, LastName: v.LastName,
			DateOfBirth: v.DateOfBirth, StartDate: v.StartDate, EndDate: v.EndDate,
			SiteCoreHourlyRateMinor: v.SiteCoreHourlyRateMinor, Notes: v.Notes,
			IsActive: v.IsActive, HasCurrentRoom: v.HasCurrentRoom, HasParentCarerContact: v.HasParentCarerContact,
			HasBookingPattern: v.HasBookingPattern,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
	default:
		return domain.Child{}
	}
	return domain.Child{
		ID:                      pgtypeUUIDToUUID(f.ID),
		FirstName:               f.FirstName,
		MiddleName:              pgtypeTextToStringPtr(f.MiddleName),
		LastName:                pgtypeTextToStringPtr(f.LastName),
		DateOfBirth:             pgtypeDateToTime(f.DateOfBirth),
		StartDate:               pgtypeDateToTime(f.StartDate),
		EndDate:                 pgtypeDateToTimePtr(f.EndDate),
		SiteCoreHourlyRateMinor: pgtypeInt4ToIntPtr(f.SiteCoreHourlyRateMinor),
		Notes:                   pgtypeTextToStringPtr(f.Notes),
		IsActive:                f.IsActive,
		HasCurrentRoom:          f.HasCurrentRoom,
		HasParentCarerContact:   f.HasParentCarerContact,
		HasBookingPattern:       f.HasBookingPattern,
		CreatedAt:               pgtypeTimestamptzToTime(f.CreatedAt),
		UpdatedAt:               pgtypeTimestamptzToTime(f.UpdatedAt),
	}
}

// --- Child Profile ---

func (r *ChildRepository) GetProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildProfile, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildProfileGetByChild(ctx, sqlc.ChildProfileGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get child profile: %w", err)
	}
	return mapChildProfileRow(row), nil
}

func (r *ChildRepository) GetProfileForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (*domain.ChildProfile, error) {
	q := sqlc.New(tx)
	row, err := q.ChildProfileGetForUpdate(ctx, sqlc.ChildProfileGetForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get child profile for update: %w", err)
	}
	return mapChildProfileRow(row), nil
}

func (r *ChildRepository) InsertProfile(ctx context.Context, tx pgx.Tx, p *domain.ChildProfile) (*domain.ChildProfile, error) {
	q := sqlc.New(tx)
	homeAddr, _ := json.Marshal(p.HomeAddress)
	if p.HomeAddress == nil {
		homeAddr = []byte("{}")
	}
	row, err := q.ChildProfileInsert(ctx, sqlc.ChildProfileInsertParams{
		ID:                          uuidToPgtype(p.ID),
		TenantID:                    uuidToPgtype(p.TenantID),
		BranchID:                    uuidToPgtype(p.BranchID),
		ChildID:                     uuidToPgtype(p.ChildID),
		Column5:                     stringPtrToText(p.Sex),
		Column6:                     stringPtrToText(p.Religion),
		Column7:                     stringPtrToText(p.EthnicOrigin),
		Column8:                     stringPtrToText(p.FirstLanguage),
		Column9:                     stringPtrToText(p.OtherLanguages),
		HomeAddress:                 homeAddr,
		Column11:                    stringPtrToText(p.HomePostcode),
		Column12:                    stringPtrToText(p.HomeTelephone),
		DisabilityStatus:            string(p.DisabilityStatus),
		Column14:                    stringPtrToText(p.DisabilityNotes),
		Column15:                    stringPtrToText(p.AccessRequirements),
		Column16:                    stringPtrToText(p.RoutineCareNotes),
		Column17:                    stringPtrToText(p.GDPRDeclaredByName),
		GdprDeclaredAt:              timestamptzPtrToPgtype(p.GDPRDeclaredAt),
		GdprDeclarationDate:         datePtrToPgtype(p.GDPRDeclarationDate),
		RegistrationDate:            datePtrToPgtype(p.RegistrationDate),
		DemographicsHomeReviewed:    p.DemographicsHomeReviewed,
		MedicalDietaryReviewed:      p.MedicalDietaryReviewed,
		HealthContactsReviewed:      p.HealthContactsReviewed,
		SocialDevelopmentReviewed:   p.SocialDevelopmentReviewed,
		ParentResponsibilityReviewed: p.ParentResponsibilityReviewed,
		EmergencyCollectionReviewed: p.EmergencyCollectionReviewed,
		RoutineCareReviewed:         p.RoutineCareReviewed,
	})
	if err != nil {
		return nil, fmt.Errorf("insert child profile: %w", err)
	}
	return mapChildProfileRow(row), nil
}

func (r *ChildRepository) UpdateProfile(ctx context.Context, tx pgx.Tx, p *domain.ChildProfile) (*domain.ChildProfile, error) {
	q := sqlc.New(tx)
	homeAddr, _ := json.Marshal(p.HomeAddress)
	if p.HomeAddress == nil {
		homeAddr = []byte("{}")
	}
	row, err := q.ChildProfileUpdate(ctx, sqlc.ChildProfileUpdateParams{
		TenantID:                    uuidToPgtype(p.TenantID),
		BranchID:                    uuidToPgtype(p.BranchID),
		ChildID:                     uuidToPgtype(p.ChildID),
		ID:                          uuidToPgtype(p.ID),
		Column5:                     stringPtrToText(p.Sex),
		Column6:                     stringPtrToText(p.Religion),
		Column7:                     stringPtrToText(p.EthnicOrigin),
		Column8:                     stringPtrToText(p.FirstLanguage),
		Column9:                     stringPtrToText(p.OtherLanguages),
		HomeAddress:                 homeAddr,
		Column11:                    stringPtrToText(p.HomePostcode),
		Column12:                    stringPtrToText(p.HomeTelephone),
		DisabilityStatus:            string(p.DisabilityStatus),
		Column14:                    stringPtrToText(p.DisabilityNotes),
		Column15:                    stringPtrToText(p.AccessRequirements),
		Column16:                    stringPtrToText(p.RoutineCareNotes),
		Column17:                    stringPtrToText(p.GDPRDeclaredByName),
		GdprDeclaredAt:              timestamptzPtrToPgtype(p.GDPRDeclaredAt),
		GdprDeclarationDate:         datePtrToPgtype(p.GDPRDeclarationDate),
		RegistrationDate:            datePtrToPgtype(p.RegistrationDate),
		DemographicsHomeReviewed:    p.DemographicsHomeReviewed,
		MedicalDietaryReviewed:      p.MedicalDietaryReviewed,
		HealthContactsReviewed:      p.HealthContactsReviewed,
		SocialDevelopmentReviewed:   p.SocialDevelopmentReviewed,
		ParentResponsibilityReviewed: p.ParentResponsibilityReviewed,
		EmergencyCollectionReviewed: p.EmergencyCollectionReviewed,
		RoutineCareReviewed:         p.RoutineCareReviewed,
	})
	if err != nil {
		return nil, fmt.Errorf("update child profile: %w", err)
	}
	return mapChildProfileRow(row), nil
}

func mapChildProfileRow(row sqlc.ChildProfile) *domain.ChildProfile {
	p := &domain.ChildProfile{
		ID:                          pgtypeUUIDToUUID(row.ID),
		TenantID:                    pgtypeUUIDToUUID(row.TenantID),
		BranchID:                    pgtypeUUIDToUUID(row.BranchID),
		ChildID:                     pgtypeUUIDToUUID(row.ChildID),
		Sex:                         pgtypeTextToPtr(row.Sex),
		Religion:                    pgtypeTextToPtr(row.Religion),
		EthnicOrigin:                pgtypeTextToPtr(row.EthnicOrigin),
		FirstLanguage:               pgtypeTextToPtr(row.FirstLanguage),
		OtherLanguages:              pgtypeTextToPtr(row.OtherLanguages),
		HomePostcode:                pgtypeTextToPtr(row.HomePostcode),
		HomeTelephone:               pgtypeTextToPtr(row.HomeTelephone),
		DisabilityStatus:            domain.YesNoUnknown(row.DisabilityStatus),
		DisabilityNotes:             pgtypeTextToPtr(row.DisabilityNotes),
		AccessRequirements:          pgtypeTextToPtr(row.AccessRequirements),
		RoutineCareNotes:            pgtypeTextToPtr(row.RoutineCareNotes),
		GDPRDeclaredByName:          pgtypeTextToPtr(row.GdprDeclaredByName),
		GDPRDeclaredAt:              pgtypeTimestamptzToTimePtr(row.GdprDeclaredAt),
		GDPRDeclarationDate:         pgtypeDateToTimePtr(row.GdprDeclarationDate),
		RegistrationDate:            pgtypeDateToTimePtr(row.RegistrationDate),
		DemographicsHomeReviewed:    row.DemographicsHomeReviewed,
		MedicalDietaryReviewed:      row.MedicalDietaryReviewed,
		HealthContactsReviewed:      row.HealthContactsReviewed,
		SocialDevelopmentReviewed:   row.SocialDevelopmentReviewed,
		ParentResponsibilityReviewed: row.ParentResponsibilityReviewed,
		EmergencyCollectionReviewed: row.EmergencyCollectionReviewed,
		RoutineCareReviewed:         row.RoutineCareReviewed,
		CreatedAt:                   pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:                   pgtypeTimestamptzToTime(row.UpdatedAt),
	}
	if row.HomeAddress != nil {
		_ = json.Unmarshal(row.HomeAddress, &p.HomeAddress)
	}
	if p.HomeAddress == nil {
		p.HomeAddress = map[string]any{}
	}
	return p
}

// --- Child Contacts ---

func (r *ChildRepository) ListContactsByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]domain.ChildContact, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ChildContactsListByChild(ctx, sqlc.ChildContactsListByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if err != nil {
		return nil, fmt.Errorf("list child contacts: %w", err)
	}
	out := make([]domain.ChildContact, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapChildContactRow(row))
	}
	return out, nil
}

func (r *ChildRepository) ReplaceContactsForTypes(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, contactTypes []domain.ContactType, entries []domain.ChildContact) error {
	q := sqlc.New(tx)
	types := make([]string, len(contactTypes))
	for i, t := range contactTypes {
		types[i] = string(t)
	}
	if err := q.ChildContactsDeleteByTypes(ctx, sqlc.ChildContactsDeleteByTypesParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		ChildID:   uuidToPgtype(childID),
		Column4:   types,
	}); err != nil {
		return fmt.Errorf("delete contacts: %w", err)
	}
	for _, e := range entries {
		addr, _ := json.Marshal(e.Address)
		if e.Address == nil {
			addr = []byte("{}")
		}
		workAddr, _ := json.Marshal(e.WorkAddress)
		if e.WorkAddress == nil {
			workAddr = []byte("{}")
		}
		if _, err := q.ChildContactsInsert(ctx, sqlc.ChildContactsInsertParams{
			ID:                        uuidToPgtype(e.ID),
			TenantID:                  uuidToPgtype(e.TenantID),
			BranchID:                  uuidToPgtype(e.BranchID),
			ChildID:                   uuidToPgtype(e.ChildID),
			ContactType:               sqlc.ChildContactType(string(e.ContactType)),
			SortOrder:                 int32(e.SortOrder),
			FullName:                  e.FullName,
			Column8:                   stringPtrToText(e.RelationshipToChild),
			Address:                   addr,
			Column10:                  stringPtrToText(e.Telephone),
			Column11:                  stringPtrToText(e.Email),
			WorkAddress:               workAddr,
			HasParentalResponsibility: boolPtrToPgtype(e.HasParentalResponsibility),
		}); err != nil {
			return fmt.Errorf("insert child contact: %w", err)
		}
	}
	return nil
}

func mapChildContactRow(row sqlc.ChildContact) domain.ChildContact {
	c := domain.ChildContact{
		ID:                        pgtypeUUIDToUUID(row.ID),
		TenantID:                  pgtypeUUIDToUUID(row.TenantID),
		BranchID:                  pgtypeUUIDToUUID(row.BranchID),
		ChildID:                   pgtypeUUIDToUUID(row.ChildID),
		ContactType:               domain.ContactType(row.ContactType),
		SortOrder:                 int(row.SortOrder),
		FullName:                  row.FullName,
		RelationshipToChild:       pgtypeTextToPtr(row.RelationshipToChild),
		Telephone:                 pgtypeTextToPtr(row.Telephone),
		Email:                     pgtypeTextToPtr(row.Email),
		HasParentalResponsibility: pgtypeBoolToPtr(row.HasParentalResponsibility),
		CreatedAt:                 pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:                 pgtypeTimestamptzToTime(row.UpdatedAt),
	}
	if row.Address != nil {
		_ = json.Unmarshal(row.Address, &c.Address)
	}
	if c.Address == nil {
		c.Address = map[string]any{}
	}
	if row.WorkAddress != nil {
		_ = json.Unmarshal(row.WorkAddress, &c.WorkAddress)
	}
	if c.WorkAddress == nil {
		c.WorkAddress = map[string]any{}
	}
	return c
}

// --- Child Health Profile ---

func (r *ChildRepository) GetHealthByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildHealthProfile, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildHealthProfileGetByChild(ctx, sqlc.ChildHealthProfileGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get child health profile: %w", err)
	}
	return mapHealthProfileRow(row), nil
}

func (r *ChildRepository) UpsertHealth(ctx context.Context, tx pgx.Tx, p *domain.ChildHealthProfile) (*domain.ChildHealthProfile, error) {
	q := sqlc.New(tx)
	row, err := q.ChildHealthProfileUpsert(ctx, sqlc.ChildHealthProfileUpsertParams{
		ID:                         uuidToPgtype(p.ID),
		TenantID:                   uuidToPgtype(p.TenantID),
		BranchID:                   uuidToPgtype(p.BranchID),
		ChildID:                    uuidToPgtype(p.ChildID),
		MedicalConditionsStatus:    string(p.MedicalConditionsStatus),
		Column6:                    stringPtrToText(p.MedicalConditionsNotes),
		PrescribedMedicationStatus: string(p.PrescribedMedicationStatus),
		Column8:                    stringPtrToText(p.MedicationNotes),
		ImmunisationStatus:         string(p.ImmunisationStatus),
		Column10:                   stringPtrToText(p.ImmunisationCountry),
		Column11:                   stringPtrToText(p.IllnessDiagnosisHistory),
		DietaryRequirementsStatus:  string(p.DietaryRequirementsStatus),
		Column13:                   stringPtrToText(p.DietaryRequirementsNotes),
		Column14:                   stringPtrToText(p.DietarySideEffects),
		Column15:                   stringPtrToText(p.DoctorName),
		Column16:                   stringPtrToText(p.DoctorAddress),
		Column17:                   stringPtrToText(p.DoctorPhone),
		Column18:                   stringPtrToText(p.HealthVisitorName),
		Column19:                   stringPtrToText(p.HealthVisitorAddress),
		Column20:                   stringPtrToText(p.HealthVisitorPhone),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert child health: %w", err)
	}
	return mapHealthProfileRow(row), nil
}

func mapHealthProfileRow(row sqlc.ChildHealthProfile) *domain.ChildHealthProfile {
	return &domain.ChildHealthProfile{
		ID:                         pgtypeUUIDToUUID(row.ID),
		TenantID:                   pgtypeUUIDToUUID(row.TenantID),
		BranchID:                   pgtypeUUIDToUUID(row.BranchID),
		ChildID:                    pgtypeUUIDToUUID(row.ChildID),
		MedicalConditionsStatus:    domain.YesNoUnknown(row.MedicalConditionsStatus),
		MedicalConditionsNotes:     pgtypeTextToPtr(row.MedicalConditionsNotes),
		PrescribedMedicationStatus: domain.YesNoUnknown(row.PrescribedMedicationStatus),
		MedicationNotes:            pgtypeTextToPtr(row.MedicationNotes),
		ImmunisationStatus:         domain.ImmunisationStatus(row.ImmunisationStatus),
		ImmunisationCountry:        pgtypeTextToPtr(row.ImmunisationCountry),
		IllnessDiagnosisHistory:    pgtypeTextToPtr(row.IllnessDiagnosisHistory),
		DietaryRequirementsStatus:  domain.YesNoUnknown(row.DietaryRequirementsStatus),
		DietaryRequirementsNotes:   pgtypeTextToPtr(row.DietaryRequirementsNotes),
		DietarySideEffects:         pgtypeTextToPtr(row.DietarySideEffects),
		DoctorName:                 pgtypeTextToPtr(row.DoctorName),
		DoctorAddress:              pgtypeTextToPtr(row.DoctorAddress),
		DoctorPhone:                pgtypeTextToPtr(row.DoctorPhone),
		HealthVisitorName:          pgtypeTextToPtr(row.HealthVisitorName),
		HealthVisitorAddress:       pgtypeTextToPtr(row.HealthVisitorAddress),
		HealthVisitorPhone:         pgtypeTextToPtr(row.HealthVisitorPhone),
		CreatedAt:                  pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:                  pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

// --- Child Safeguarding Profile ---

func (r *ChildRepository) GetSafeguardingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildSafeguardingProfile, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildSafeguardingProfileGetByChild(ctx, sqlc.ChildSafeguardingProfileGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get child safeguarding profile: %w", err)
	}
	return mapSafeguardingProfileRow(row), nil
}

func (r *ChildRepository) UpsertSafeguarding(ctx context.Context, tx pgx.Tx, p *domain.ChildSafeguardingProfile) (*domain.ChildSafeguardingProfile, error) {
	q := sqlc.New(tx)
	referrals, _ := json.Marshal(p.ProfessionalReferrals)
	if p.ProfessionalReferrals == nil {
		referrals = []byte("[]")
	}
	row, err := q.ChildSafeguardingProfileUpsert(ctx, sqlc.ChildSafeguardingProfileUpsertParams{
		ID:                          uuidToPgtype(p.ID),
		TenantID:                    uuidToPgtype(p.TenantID),
		BranchID:                    uuidToPgtype(p.BranchID),
		ChildID:                     uuidToPgtype(p.ChildID),
		SocialServicesStatus:        string(p.SocialServicesStatus),
		Column6:                     stringPtrToText(p.SocialServicesNotes),
		Column7:                     stringPtrToText(p.SocialWorkerName),
		Column8:                     stringPtrToText(p.SocialWorkerPhone),
		Column9:                     stringPtrToText(p.SocialWorkerEmail),
		ConcernWalking:              string(p.ConcernWalking),
		ConcernSpeechLanguage:       string(p.ConcernSpeechLanguage),
		ConcernHearing:              string(p.ConcernHearing),
		ConcernSight:                string(p.ConcernSight),
		ConcernEmotionalWellbeing:   string(p.ConcernEmotionalWellbeing),
		ConcernBehaviour:            string(p.ConcernBehaviour),
		ProfessionalReferrals:       referrals,
		Column17:                    stringPtrToText(p.RestrictedNotes),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert child safeguarding: %w", err)
	}
	return mapSafeguardingProfileRow(row), nil
}

func mapSafeguardingProfileRow(row sqlc.ChildSafeguardingProfile) *domain.ChildSafeguardingProfile {
	p := &domain.ChildSafeguardingProfile{
		ID:                          pgtypeUUIDToUUID(row.ID),
		TenantID:                    pgtypeUUIDToUUID(row.TenantID),
		BranchID:                    pgtypeUUIDToUUID(row.BranchID),
		ChildID:                     pgtypeUUIDToUUID(row.ChildID),
		SocialServicesStatus:        domain.YesNoUnknown(row.SocialServicesStatus),
		SocialServicesNotes:         pgtypeTextToPtr(row.SocialServicesNotes),
		SocialWorkerName:            pgtypeTextToPtr(row.SocialWorkerName),
		SocialWorkerPhone:           pgtypeTextToPtr(row.SocialWorkerPhone),
		SocialWorkerEmail:           pgtypeTextToPtr(row.SocialWorkerEmail),
		ConcernWalking:              domain.YesNoUnknown(row.ConcernWalking),
		ConcernSpeechLanguage:       domain.YesNoUnknown(row.ConcernSpeechLanguage),
		ConcernHearing:              domain.YesNoUnknown(row.ConcernHearing),
		ConcernSight:                domain.YesNoUnknown(row.ConcernSight),
		ConcernEmotionalWellbeing:   domain.YesNoUnknown(row.ConcernEmotionalWellbeing),
		ConcernBehaviour:            domain.YesNoUnknown(row.ConcernBehaviour),
		RestrictedNotes:             pgtypeTextToPtr(row.RestrictedNotes),
		CreatedAt:                   pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:                   pgtypeTimestamptzToTime(row.UpdatedAt),
	}
	if row.ProfessionalReferrals != nil {
		_ = json.Unmarshal(row.ProfessionalReferrals, &p.ProfessionalReferrals)
	}
	if p.ProfessionalReferrals == nil {
		p.ProfessionalReferrals = []domain.ProfessionalReferral{}
	}
	return p
}

// --- Child Consent ---

func (r *ChildRepository) GetConsentByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildConsent, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildConsentGetByChild(ctx, sqlc.ChildConsentGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get child consent: %w", err)
	}
	return mapConsentRow(row), true, nil
}

func (r *ChildRepository) InsertConsent(ctx context.Context, tx pgx.Tx, p *domain.ChildConsent) (*domain.ChildConsent, error) {
	q := sqlc.New(tx)
	row, err := q.ChildConsentInsert(ctx, sqlc.ChildConsentInsertParams{
		ID:                                   uuidToPgtype(p.ID),
		TenantID:                             uuidToPgtype(p.TenantID),
		BranchID:                             uuidToPgtype(p.BranchID),
		ChildID:                              uuidToPgtype(p.ChildID),
		UrgentMedicalTreatment:               p.UrgentMedicalTreatment,
		Column6:                              stringPtrToText(p.UrgentMedicalTreatmentExceptions),
		Plasters:                             p.Plasters,
		SafeguardingReportingAcknowledgement: p.SafeguardingReportingAcknowledgement,
		InformationSharingConsent:            p.InformationSharingConsent,
		GdprDataProcessingConsent:            p.GDPRDataProcessingConsent,
		AreaSencoLiaison:                     p.AreaSENCOLiaison,
		HealthVisitorLiaison:                 p.HealthVisitorLiaison,
		TransitionDocuments:                  p.TransitionDocuments,
		LocalOutings:                         p.LocalOutings,
		FacePainting:                         p.FacePainting,
		ParentSuppliedSunCream:               p.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:             p.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:             p.DevelopmentProfilePhotos,
		NurseryDisplayBoards:                 p.NurseryDisplayBoards,
		PromotionalLiterature:                p.PromotionalLiterature,
		NurseryWebsite:                       p.NurseryWebsite,
		StaffStudentCoursework:               p.StaffStudentCoursework,
		SocialMedia:                          p.SocialMedia,
		Column24:                             stringPtrToText(p.SocialMediaChannelNotes),
		Column25:                             stringPtrToText(p.NotesExceptions),
		SignerName:                           p.SignerName,
		SignedDate:                           timeToPgtypeDate(p.SignedDate),
		PaperFormOnFile:                      p.PaperFormOnFile,
		EnteredByUserID:                      uuidToPgtype(p.EnteredByUserID),
		EnteredByMembershipID:                uuidToPgtype(p.EnteredByMembershipID),
	})
	if err != nil {
		return nil, fmt.Errorf("insert child consent: %w", err)
	}
	return mapConsentRow(row), nil
}

func (r *ChildRepository) UpdateConsent(ctx context.Context, tx pgx.Tx, p *domain.ChildConsent) (*domain.ChildConsent, error) {
	q := sqlc.New(tx)
	row, err := q.ChildConsentUpdate(ctx, sqlc.ChildConsentUpdateParams{
		TenantID:                             uuidToPgtype(p.TenantID),
		BranchID:                             uuidToPgtype(p.BranchID),
		ChildID:                              uuidToPgtype(p.ChildID),
		ID:                                   uuidToPgtype(p.ID),
		UrgentMedicalTreatment:               p.UrgentMedicalTreatment,
		Column6:                              stringPtrToText(p.UrgentMedicalTreatmentExceptions),
		Plasters:                             p.Plasters,
		SafeguardingReportingAcknowledgement: p.SafeguardingReportingAcknowledgement,
		InformationSharingConsent:            p.InformationSharingConsent,
		GdprDataProcessingConsent:            p.GDPRDataProcessingConsent,
		AreaSencoLiaison:                     p.AreaSENCOLiaison,
		HealthVisitorLiaison:                 p.HealthVisitorLiaison,
		TransitionDocuments:                  p.TransitionDocuments,
		LocalOutings:                         p.LocalOutings,
		FacePainting:                         p.FacePainting,
		ParentSuppliedSunCream:               p.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:             p.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:             p.DevelopmentProfilePhotos,
		NurseryDisplayBoards:                 p.NurseryDisplayBoards,
		PromotionalLiterature:                p.PromotionalLiterature,
		NurseryWebsite:                       p.NurseryWebsite,
		StaffStudentCoursework:               p.StaffStudentCoursework,
		SocialMedia:                          p.SocialMedia,
		Column24:                             stringPtrToText(p.SocialMediaChannelNotes),
		Column25:                             stringPtrToText(p.NotesExceptions),
		SignerName:                           p.SignerName,
		SignedDate:                           timeToPgtypeDate(p.SignedDate),
		PaperFormOnFile:                      p.PaperFormOnFile,
		EnteredByUserID:                      uuidToPgtype(p.EnteredByUserID),
		EnteredByMembershipID:                uuidToPgtype(p.EnteredByMembershipID),
	})
	if err != nil {
		return nil, fmt.Errorf("update child consent: %w", err)
	}
	return mapConsentRow(row), nil
}

func mapConsentRow(row sqlc.ChildConsentRecord) *domain.ChildConsent {
	return &domain.ChildConsent{
		ID:                                   pgtypeUUIDToUUID(row.ID),
		TenantID:                             pgtypeUUIDToUUID(row.TenantID),
		BranchID:                             pgtypeUUIDToUUID(row.BranchID),
		ChildID:                              pgtypeUUIDToUUID(row.ChildID),
		UrgentMedicalTreatment:               row.UrgentMedicalTreatment,
		UrgentMedicalTreatmentExceptions:     pgtypeTextToPtr(row.UrgentMedicalTreatmentExceptions),
		Plasters:                             row.Plasters,
		SafeguardingReportingAcknowledgement: row.SafeguardingReportingAcknowledgement,
		InformationSharingConsent:            row.InformationSharingConsent,
		GDPRDataProcessingConsent:            row.GdprDataProcessingConsent,
		AreaSENCOLiaison:                     row.AreaSencoLiaison,
		HealthVisitorLiaison:                 row.HealthVisitorLiaison,
		TransitionDocuments:                  row.TransitionDocuments,
		LocalOutings:                         row.LocalOutings,
		FacePainting:                         row.FacePainting,
		ParentSuppliedSunCream:               row.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:             row.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:             row.DevelopmentProfilePhotos,
		NurseryDisplayBoards:                 row.NurseryDisplayBoards,
		PromotionalLiterature:                row.PromotionalLiterature,
		NurseryWebsite:                       row.NurseryWebsite,
		StaffStudentCoursework:               row.StaffStudentCoursework,
		SocialMedia:                          row.SocialMedia,
		SocialMediaChannelNotes:              pgtypeTextToPtr(row.SocialMediaChannelNotes),
		NotesExceptions:                      pgtypeTextToPtr(row.NotesExceptions),
		SignerName:                           row.SignerName,
		SignedDate:                           pgtypeDateToTime(row.SignedDate),
		PaperFormOnFile:                      row.PaperFormOnFile,
		EnteredByUserID:                      pgtypeUUIDToUUID(row.EnteredByUserID),
		EnteredByMembershipID:                pgtypeUUIDToUUID(row.EnteredByMembershipID),
		CreatedAt:                            pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:                            pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

// --- Child Funding Record ---

func (r *ChildRepository) GetFundingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildFundingRecord, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildFundingRecordGetByChild(ctx, sqlc.ChildFundingRecordGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get child funding: %w", err)
	}
	return mapFundingRow(row), true, nil
}

func (r *ChildRepository) UpsertFunding(ctx context.Context, tx pgx.Tx, p *domain.ChildFundingRecord) (*domain.ChildFundingRecord, error) {
	q := sqlc.New(tx)
	row, err := q.ChildFundingRecordUpsert(ctx, sqlc.ChildFundingRecordUpsertParams{
		ID:                          uuidToPgtype(p.ID),
		TenantID:                    uuidToPgtype(p.TenantID),
		BranchID:                    uuidToPgtype(p.BranchID),
		ChildID:                     uuidToPgtype(p.ChildID),
		BenefitsContributeToFees:    string(p.BenefitsContributeToFees),
		WorkingTaxCredit:            string(p.WorkingTaxCredit),
		CollegeUniPaidToParent:      string(p.CollegeUniPaidToParent),
		CollegeUniPaidToNursery:     string(p.CollegeUniPaidToNursery),
		Funding3yoTermTime:          string(p.Funding3yoTermTime),
		Funding2yoTermTime:          string(p.Funding2yoTermTime),
		Column11:                    stringPtrToText(p.FundingSupportNotes),
		FundingSupportReviewed:      p.FundingSupportReviewed,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert child funding: %w", err)
	}
	return mapFundingRow(row), nil
}

func mapFundingRow(row sqlc.ChildFundingRecord) *domain.ChildFundingRecord {
	return &domain.ChildFundingRecord{
		ID:                          pgtypeUUIDToUUID(row.ID),
		TenantID:                    pgtypeUUIDToUUID(row.TenantID),
		BranchID:                    pgtypeUUIDToUUID(row.BranchID),
		ChildID:                     pgtypeUUIDToUUID(row.ChildID),
		BenefitsContributeToFees:    domain.YesNoUnknown(row.BenefitsContributeToFees),
		WorkingTaxCredit:            domain.YesNoUnknown(row.WorkingTaxCredit),
		CollegeUniPaidToParent:      domain.YesNoUnknown(row.CollegeUniPaidToParent),
		CollegeUniPaidToNursery:     domain.YesNoUnknown(row.CollegeUniPaidToNursery),
		Funding3yoTermTime:          domain.YesNoUnknown(row.Funding3yoTermTime),
		Funding2yoTermTime:          domain.YesNoUnknown(row.Funding2yoTermTime),
		FundingSupportNotes:         pgtypeTextToPtr(row.FundingSupportNotes),
		FundingSupportReviewed:      row.FundingSupportReviewed,
		CreatedAt:                   pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:                   pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

// --- Child Collection Settings ---

func (r *ChildRepository) GetCollectionSettingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildCollectionSetting, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildCollectionSettingGetByChild(ctx, sqlc.ChildCollectionSettingGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get child collection setting: %w", err)
	}
	return mapCollectionSettingRow(row), nil
}

func (r *ChildRepository) UpsertCollectionSetting(ctx context.Context, tx pgx.Tx, p *domain.ChildCollectionSetting) (*domain.ChildCollectionSetting, error) {
	q := sqlc.New(tx)
	row, err := q.ChildCollectionSettingUpsert(ctx, sqlc.ChildCollectionSettingUpsertParams{
		ID:                          uuidToPgtype(p.ID),
		TenantID:                    uuidToPgtype(p.TenantID),
		BranchID:                    uuidToPgtype(p.BranchID),
		ChildID:                     uuidToPgtype(p.ChildID),
		Over18CollectionAcknowledged: p.Over18CollectionAcknowledged,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert child collection setting: %w", err)
	}
	return mapCollectionSettingRow(row), nil
}

func (r *ChildRepository) SetCollectionPassword(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID, id uuid.UUID, hash string, updatedAt time.Time, userID, membershipID uuid.UUID) error {
	q := sqlc.New(tx)
	_, err := q.ChildCollectionSettingSetPassword(ctx, sqlc.ChildCollectionSettingSetPasswordParams{
		TenantID:                                uuidToPgtype(tenantID),
		BranchID:                                uuidToPgtype(branchID),
		ChildID:                                 uuidToPgtype(childID),
		ID:                                      uuidToPgtype(id),
		CollectionPasswordHash:                  pgtype.Text{String: hash, Valid: true},
		CollectionPasswordUpdatedAt:             pgtype.Timestamptz{Time: updatedAt, Valid: true},
		CollectionPasswordUpdatedByUserID:       uuidToPgtype(userID),
		CollectionPasswordUpdatedByMembershipID: uuidToPgtype(membershipID),
	})
	if err != nil {
		return fmt.Errorf("set child collection password: %w", err)
	}
	return nil
}

func mapCollectionSettingRow(row sqlc.ChildCollectionSetting) *domain.ChildCollectionSetting {
	return &domain.ChildCollectionSetting{
		ID:                                pgtypeUUIDToUUID(row.ID),
		TenantID:                          pgtypeUUIDToUUID(row.TenantID),
		BranchID:                          pgtypeUUIDToUUID(row.BranchID),
		ChildID:                           pgtypeUUIDToUUID(row.ChildID),
		Over18CollectionAcknowledged:      row.Over18CollectionAcknowledged,
		CollectionPasswordIsSet:           row.CollectionPasswordHash.String != "",
		CollectionPasswordUpdatedAt:       pgtypeTimestamptzToTimePtr(row.CollectionPasswordUpdatedAt),
		CollectionPasswordUpdatedByUserID: pgtypeUUIDToUUIDPtr(row.CollectionPasswordUpdatedByUserID),
		CollectionPasswordUpdatedByMembershipID: pgtypeUUIDToUUIDPtr(row.CollectionPasswordUpdatedByMembershipID),
		CreatedAt:                         pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:                         pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

// --- Child Room Assignments ---

func (r *ChildRepository) ListRoomAssignmentsByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]domain.ChildRoomAssignment, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ChildRoomAssignmentsListByChild(ctx, sqlc.ChildRoomAssignmentsListByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if err != nil {
		return nil, fmt.Errorf("list child room assignments: %w", err)
	}
	out := make([]domain.ChildRoomAssignment, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRoomAssignmentRow(row))
	}
	return out, nil
}

func (r *ChildRepository) GetCurrentRoomAssignmentByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildRoomAssignment, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildRoomAssignmentsGetCurrentByChild(ctx, sqlc.ChildRoomAssignmentsGetCurrentByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get current child room assignment: %w", err)
	}
	a := mapRoomAssignmentRow(row)
	return &a, true, nil
}

func (r *ChildRepository) InsertRoomAssignment(ctx context.Context, tx pgx.Tx, a *domain.ChildRoomAssignment) (*domain.ChildRoomAssignment, error) {
	q := sqlc.New(tx)
	row, err := q.ChildRoomAssignmentsInsert(ctx, sqlc.ChildRoomAssignmentsInsertParams{
		ID:        uuidToPgtype(a.ID),
		TenantID:  uuidToPgtype(a.TenantID),
		BranchID:  uuidToPgtype(a.BranchID),
		ChildID:   uuidToPgtype(a.ChildID),
		RoomID:    uuidToPgtype(a.RoomID),
		StartDate: timeToPgtypeDate(a.StartDate),
	})
	if err != nil {
		return nil, fmt.Errorf("insert child room assignment: %w", err)
	}
	out := mapRoomAssignmentRow(row)
	return &out, nil
}

func (r *ChildRepository) CloseCurrentRoomAssignment(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, endDate time.Time) error {
	q := sqlc.New(tx)
	if err := q.ChildRoomAssignmentsCloseCurrent(ctx, sqlc.ChildRoomAssignmentsCloseCurrentParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		ChildID:   uuidToPgtype(childID),
		EndDate:   timeToPgtypeDate(endDate),
	}); err != nil {
		return fmt.Errorf("close current child room assignment: %w", err)
	}
	return nil
}

func (r *ChildRepository) GetRoomAssignmentByID(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (*domain.ChildRoomAssignment, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ChildRoomAssignmentsGetByID(ctx, sqlc.ChildRoomAssignmentsGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get child room assignment by id: %w", err)
	}
	a := mapRoomAssignmentRow(row)
	return &a, true, nil
}

func (r *ChildRepository) CloseRoomAssignmentByID(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, endDate time.Time) (bool, error) {
	q := sqlc.New(tx)
	err := q.ChildRoomAssignmentsCloseByID(ctx, sqlc.ChildRoomAssignmentsCloseByIDParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		ID:        uuidToPgtype(id),
		EndDate:   timeToPgtypeDate(endDate),
	})
	if err != nil {
		return false, fmt.Errorf("close child room assignment: %w", err)
	}
	return true, nil
}

func mapRoomAssignmentRow(row interface{}) domain.ChildRoomAssignment {
	type fields struct {
		ID        pgtype.UUID
		TenantID  pgtype.UUID
		BranchID  pgtype.UUID
		ChildID   pgtype.UUID
		RoomID    pgtype.UUID
		StartDate pgtype.Date
		EndDate   pgtype.Date
		CreatedAt pgtype.Timestamptz
	}
	var f fields
	var isCurrent bool
	switch v := row.(type) {
	case sqlc.ChildRoomAssignment:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			RoomID: v.RoomID, StartDate: v.StartDate, EndDate: v.EndDate,
			CreatedAt: v.CreatedAt,
		}
		isCurrent = v.IsCurrent
	case sqlc.ChildRoomAssignmentsListByChildRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			RoomID: v.RoomID, StartDate: v.StartDate, EndDate: v.EndDate,
			CreatedAt: v.CreatedAt,
		}
		isCurrent = !v.EndDate.Valid
	case sqlc.ChildRoomAssignmentsGetCurrentByChildRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			RoomID: v.RoomID, StartDate: v.StartDate, EndDate: v.EndDate,
			CreatedAt: v.CreatedAt,
		}
		isCurrent = true
	case sqlc.ChildRoomAssignmentsInsertRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			RoomID: v.RoomID, StartDate: v.StartDate, EndDate: v.EndDate,
			CreatedAt: v.CreatedAt,
		}
		isCurrent = !v.EndDate.Valid
	case sqlc.ChildRoomAssignmentsGetByIDRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			RoomID: v.RoomID, StartDate: v.StartDate, EndDate: v.EndDate,
			CreatedAt: v.CreatedAt,
		}
		isCurrent = !v.EndDate.Valid
	default:
		return domain.ChildRoomAssignment{}
	}
	return domain.ChildRoomAssignment{
		ID:        pgtypeUUIDToUUID(f.ID),
		TenantID:  pgtypeUUIDToUUID(f.TenantID),
		BranchID:  pgtypeUUIDToUUID(f.BranchID),
		ChildID:   pgtypeUUIDToUUID(f.ChildID),
		RoomID:    pgtypeUUIDToUUID(f.RoomID),
		StartDate: pgtypeDateToTime(f.StartDate),
		EndDate:   pgtypeDateToTimePtr(f.EndDate),
		IsCurrent: isCurrent,
		CreatedAt: pgtypeTimestamptzToTime(f.CreatedAt),
	}
}

// --- Child Billing Profile ---

func (r *ChildRepository) GetBillingProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildBillingProfile, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildBillingProfileGetByChild(ctx, sqlc.ChildBillingProfileGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get child billing profile: %w", err)
	}
	return mapBillingProfileRow(row), true, nil
}

func (r *ChildRepository) UpsertBillingProfile(ctx context.Context, tx pgx.Tx, p *domain.ChildBillingProfile) (*domain.ChildBillingProfile, error) {
	q := sqlc.New(tx)
	effectiveFrom := ""
	if !p.EffectiveFrom.IsZero() {
		effectiveFrom = p.EffectiveFrom.Format("2006-01-02")
	}
	row, err := q.ChildBillingProfileUpsert(ctx, sqlc.ChildBillingProfileUpsertParams{
		ID:              uuidToPgtype(p.ID),
		TenantID:        uuidToPgtype(p.TenantID),
		BranchID:        uuidToPgtype(p.BranchID),
		ChildID:         uuidToPgtype(p.ChildID),
		BillingBasis:    string(p.BillingBasis),
		CustomRateMinor: int32PtrToPgtype(p.CustomRateMinor),
		Column7:         effectiveFrom,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert child billing profile: %w", err)
	}
	return mapBillingProfileRow(row), nil
}

func mapBillingProfileRow(row sqlc.ChildBillingProfile) *domain.ChildBillingProfile {
	return &domain.ChildBillingProfile{
		ID:              pgtypeUUIDToUUID(row.ID),
		TenantID:        pgtypeUUIDToUUID(row.TenantID),
		BranchID:        pgtypeUUIDToUUID(row.BranchID),
		ChildID:         pgtypeUUIDToUUID(row.ChildID),
		BillingBasis:    domain.BillingBasis(row.BillingBasis),
		CustomRateMinor: pgtypeInt4ToIntPtr(row.CustomRateMinor),
		EffectiveFrom:   pgtypeDateToTime(row.EffectiveFrom),
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

// --- Child Leaving Record ---

func (r *ChildRepository) GetLeavingRecordByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildLeavingRecord, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildLeavingRecordGetByChild(ctx, sqlc.ChildLeavingRecordGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get child leaving record: %w", err)
	}
	return mapLeavingRecordRow(row), true, nil
}

func (r *ChildRepository) InsertLeavingRecord(ctx context.Context, tx pgx.Tx, p *domain.ChildLeavingRecord) error {
	q := sqlc.New(tx)
	_, err := q.ChildLeavingRecordInsert(ctx, sqlc.ChildLeavingRecordInsertParams{
		ID:         uuidToPgtype(p.ID),
		TenantID:   uuidToPgtype(p.TenantID),
		BranchID:   uuidToPgtype(p.BranchID),
		ChildID:    uuidToPgtype(p.ChildID),
		LeftAt:     pgtype.Timestamptz{Time: p.LeftAt, Valid: true},
		ReasonCode: p.ReasonCode,
		Column7:    stringPtrToText(p.ReasonNote),
	})
	if err != nil {
		return fmt.Errorf("insert child leaving record: %w", err)
	}
	return nil
}

func mapLeavingRecordRow(row sqlc.ChildLeavingRecord) *domain.ChildLeavingRecord {
	return &domain.ChildLeavingRecord{
		ID:         pgtypeUUIDToUUID(row.ID),
		TenantID:   pgtypeUUIDToUUID(row.TenantID),
		BranchID:   pgtypeUUIDToUUID(row.BranchID),
		ChildID:    pgtypeUUIDToUUID(row.ChildID),
		LeftAt:     pgtypeTimestamptzToTime(row.LeftAt),
		ReasonCode: row.ReasonCode,
		ReasonNote: pgtypeTextToPtr(row.ReasonNote),
		CreatedAt:  pgtypeTimestamptzToTime(row.CreatedAt),
	}
}

// --- Child Booking Patterns ---

func (r *ChildRepository) ListByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]domain.BookingPattern, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ChildBookingPatternsListByChild(ctx, sqlc.ChildBookingPatternsListByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if err != nil {
		return nil, fmt.Errorf("list child booking patterns: %w", err)
	}
	out := make([]domain.BookingPattern, 0, len(rows))
	for _, row := range rows {
		bp := mapBookingPatternRow(row)
		entries, eerr := r.entriesForPattern(ctx, tenantID, branchID, bp.ID)
		if eerr != nil {
			return nil, eerr
		}
		bp.Entries = entries
		out = append(out, bp)
	}
	return out, nil
}

func (r *ChildRepository) GetPatternByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (*domain.BookingPattern, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildBookingPatternsGetByID(ctx, sqlc.ChildBookingPatternsGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get child booking pattern by id: %w", err)
	}
	bp := mapBookingPatternRow(row)
	entries, eerr := r.entriesForPattern(ctx, tenantID, branchID, bp.ID)
	if eerr != nil {
		return nil, false, eerr
	}
	bp.Entries = entries
	return &bp, true, nil
}

func (r *ChildRepository) GetActiveForDate(ctx context.Context, tenantID, branchID, childID uuid.UUID, date time.Time) (*domain.BookingPattern, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildBookingPatternsGetActiveForDate(ctx, sqlc.ChildBookingPatternsGetActiveForDateParams{
		TenantID:      uuidToPgtype(tenantID),
		BranchID:      uuidToPgtype(branchID),
		ChildID:       uuidToPgtype(childID),
		EffectiveFrom: timeToPgtypeDate(date),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get active child booking pattern: %w", err)
	}
	bp := mapBookingPatternRow(row)
	entries, eerr := r.entriesForPattern(ctx, tenantID, branchID, bp.ID)
	if eerr != nil {
		return nil, false, eerr
	}
	bp.Entries = entries
	return &bp, true, nil
}

func (r *ChildRepository) GetCurrentOpenByChild(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (*domain.BookingPattern, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ChildBookingPatternsGetCurrentOpenByChild(ctx, sqlc.ChildBookingPatternsGetCurrentOpenByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get current open child booking pattern: %w", err)
	}
	bp := mapBookingPatternRow(row)
	// No entries: open pattern body is not yet hydrated here; create flow uses InsertPattern.
	return &bp, true, nil
}

func (r *ChildRepository) GetPreviousClosedByChild(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (*domain.BookingPattern, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ChildBookingPatternsGetPreviousClosedByChild(ctx, sqlc.ChildBookingPatternsGetPreviousClosedByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get previous closed child booking pattern: %w", err)
	}
	bp := mapBookingPatternRow(row)
	return &bp, true, nil
}

func (r *ChildRepository) InsertPattern(ctx context.Context, tx pgx.Tx, p *domain.BookingPattern, entries []domain.BookingPatternEntry) (*domain.BookingPattern, error) {
	q := sqlc.New(tx)
	row, err := q.ChildBookingPatternsInsert(ctx, sqlc.ChildBookingPatternsInsertParams{
		ID:            uuidToPgtype(p.ID),
		TenantID:      uuidToPgtype(p.TenantID),
		BranchID:      uuidToPgtype(p.BranchID),
		ChildID:       uuidToPgtype(p.ChildID),
		EffectiveFrom: timeToPgtypeDate(p.EffectiveFrom),
		EffectiveTo:   timeToPgtypeDatePtr(p.EffectiveTo),
	})
	if err != nil {
		return nil, fmt.Errorf("insert child booking pattern: %w", err)
	}
	for i := range entries {
		e := &entries[i]
		e.PatternID = p.ID
		e.TenantID = p.TenantID
		e.BranchID = p.BranchID
		if err := q.ChildBookingPatternEntriesInsert(ctx, sqlc.ChildBookingPatternEntriesInsertParams{
			ID:            uuidToPgtype(e.ID),
			TenantID:      uuidToPgtype(e.TenantID),
			BranchID:      uuidToPgtype(e.BranchID),
			PatternID:     uuidToPgtype(e.PatternID),
			DayOfWeek:     int32(e.DayOfWeek),
			SessionTypeID: uuidToPgtype(e.SessionType.ID),
		}); err != nil {
			return nil, fmt.Errorf("insert child booking pattern entry: %w", err)
		}
	}
	bp := mapBookingPatternRow(row)
	// Re-load entries with joined session type for return value.
	loadedEntries, lerr := r.entriesForPatternTx(ctx, tx, p.TenantID, p.BranchID, p.ID)
	if lerr != nil {
		return nil, lerr
	}
	bp.Entries = loadedEntries
	return &bp, nil
}

func (r *ChildRepository) CloseCurrentPattern(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, effectiveTo time.Time) error {
	q := sqlc.New(tx)
	if err := q.ChildBookingPatternsCloseCurrent(ctx, sqlc.ChildBookingPatternsCloseCurrentParams{
		TenantID:    uuidToPgtype(tenantID),
		BranchID:    uuidToPgtype(branchID),
		ChildID:     uuidToPgtype(childID),
		EffectiveTo: timeToPgtypeDate(effectiveTo),
	}); err != nil {
		return fmt.Errorf("close current child booking pattern: %w", err)
	}
	return nil
}

func (r *ChildRepository) ClosePatternByID(ctx context.Context, tx pgx.Tx, tenantID, branchID, patternID uuid.UUID, effectiveTo time.Time) error {
	q := sqlc.New(tx)
	if err := q.ChildBookingPatternsCloseByID(ctx, sqlc.ChildBookingPatternsCloseByIDParams{
		TenantID:    uuidToPgtype(tenantID),
		BranchID:    uuidToPgtype(branchID),
		ID:          uuidToPgtype(patternID),
		EffectiveTo: timeToPgtypeDate(effectiveTo),
	}); err != nil {
		return fmt.Errorf("close child booking pattern by id: %w", err)
	}
	return nil
}

func (r *ChildRepository) ReplaceEntries(ctx context.Context, tx pgx.Tx, tenantID, branchID, patternID uuid.UUID, entries []domain.BookingPatternEntry) error {
	q := sqlc.New(tx)
	if err := q.ChildBookingPatternEntriesDeleteByPattern(ctx, sqlc.ChildBookingPatternEntriesDeleteByPatternParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		PatternID: uuidToPgtype(patternID),
	}); err != nil {
		return fmt.Errorf("delete child booking pattern entries: %w", err)
	}
	for i := range entries {
		e := &entries[i]
		e.PatternID = patternID
		e.TenantID = tenantID
		e.BranchID = branchID
		if err := q.ChildBookingPatternEntriesInsert(ctx, sqlc.ChildBookingPatternEntriesInsertParams{
			ID:            uuidToPgtype(e.ID),
			TenantID:      uuidToPgtype(e.TenantID),
			BranchID:      uuidToPgtype(e.BranchID),
			PatternID:     uuidToPgtype(e.PatternID),
			DayOfWeek:     int32(e.DayOfWeek),
			SessionTypeID: uuidToPgtype(e.SessionType.ID),
		}); err != nil {
			return fmt.Errorf("insert child booking pattern entry: %w", err)
		}
	}
	return nil
}

func (r *ChildRepository) UpdateEffectiveFrom(ctx context.Context, tx pgx.Tx, tenantID, branchID, patternID uuid.UUID, effectiveFrom time.Time) error {
	q := sqlc.New(tx)
	if err := q.ChildBookingPatternsUpdateEffectiveFrom(ctx, sqlc.ChildBookingPatternsUpdateEffectiveFromParams{
		TenantID:      uuidToPgtype(tenantID),
		BranchID:      uuidToPgtype(branchID),
		ID:            uuidToPgtype(patternID),
		EffectiveFrom: timeToPgtypeDate(effectiveFrom),
	}); err != nil {
		return fmt.Errorf("update child booking pattern effective_from: %w", err)
	}
	return nil
}

func (r *ChildRepository) entriesForPattern(ctx context.Context, tenantID, branchID, patternID uuid.UUID) ([]domain.BookingPatternEntry, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ChildBookingPatternEntriesListByPattern(ctx, sqlc.ChildBookingPatternEntriesListByPatternParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		PatternID: uuidToPgtype(patternID),
	})
	if err != nil {
		return nil, fmt.Errorf("list child booking pattern entries: %w", err)
	}
	out := make([]domain.BookingPatternEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapBookingPatternEntryRow(row))
	}
	return out, nil
}

func (r *ChildRepository) entriesForPatternTx(ctx context.Context, tx pgx.Tx, tenantID, branchID, patternID uuid.UUID) ([]domain.BookingPatternEntry, error) {
	q := sqlc.New(tx)
	rows, err := q.ChildBookingPatternEntriesListByPattern(ctx, sqlc.ChildBookingPatternEntriesListByPatternParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		PatternID: uuidToPgtype(patternID),
	})
	if err != nil {
		return nil, fmt.Errorf("list child booking pattern entries (tx): %w", err)
	}
	out := make([]domain.BookingPatternEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapBookingPatternEntryRow(row))
	}
	return out, nil
}

func mapBookingPatternRow(row interface{}) domain.BookingPattern {
	type fields struct {
		ID            pgtype.UUID
		TenantID      pgtype.UUID
		BranchID      pgtype.UUID
		ChildID       pgtype.UUID
		EffectiveFrom pgtype.Date
		EffectiveTo   pgtype.Date
		CreatedAt     pgtype.Timestamptz
		UpdatedAt     pgtype.Timestamptz
	}
	var f fields
	var isCurrent bool
	switch v := row.(type) {
	case sqlc.ChildBookingPattern:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			EffectiveFrom: v.EffectiveFrom, EffectiveTo: v.EffectiveTo,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
		isCurrent = v.IsCurrent
	case sqlc.ChildBookingPatternsListByChildRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			EffectiveFrom: v.EffectiveFrom, EffectiveTo: v.EffectiveTo,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
		isCurrent = !v.EffectiveTo.Valid
	case sqlc.ChildBookingPatternsGetByIDRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			EffectiveFrom: v.EffectiveFrom, EffectiveTo: v.EffectiveTo,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
		isCurrent = !v.EffectiveTo.Valid
	case sqlc.ChildBookingPatternsGetActiveForDateRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			EffectiveFrom: v.EffectiveFrom, EffectiveTo: v.EffectiveTo,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
		isCurrent = !v.EffectiveTo.Valid
	case sqlc.ChildBookingPatternsGetCurrentOpenByChildRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			EffectiveFrom: v.EffectiveFrom, EffectiveTo: v.EffectiveTo,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
		isCurrent = true
	case sqlc.ChildBookingPatternsGetPreviousClosedByChildRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			EffectiveFrom: v.EffectiveFrom, EffectiveTo: v.EffectiveTo,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
		isCurrent = false
	case sqlc.ChildBookingPatternsInsertRow:
		f = fields{
			ID: v.ID, TenantID: v.TenantID, BranchID: v.BranchID, ChildID: v.ChildID,
			EffectiveFrom: v.EffectiveFrom, EffectiveTo: v.EffectiveTo,
			CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
		}
		isCurrent = !v.EffectiveTo.Valid
	default:
		return domain.BookingPattern{}
	}
	return domain.BookingPattern{
		ID:            pgtypeUUIDToUUID(f.ID),
		TenantID:      pgtypeUUIDToUUID(f.TenantID),
		BranchID:      pgtypeUUIDToUUID(f.BranchID),
		ChildID:       pgtypeUUIDToUUID(f.ChildID),
		EffectiveFrom: pgtypeDateToTime(f.EffectiveFrom),
		EffectiveTo:   pgtypeDateToTimePtr(f.EffectiveTo),
		IsCurrent:     isCurrent,
		CreatedAt:     pgtypeTimestamptzToTime(f.CreatedAt),
		UpdatedAt:     pgtypeTimestamptzToTime(f.UpdatedAt),
	}
}

func mapBookingPatternEntryRow(row sqlc.ChildBookingPatternEntriesListByPatternRow) domain.BookingPatternEntry {
	return domain.BookingPatternEntry{
		ID:        pgtypeUUIDToUUID(row.ID),
		TenantID:  pgtypeUUIDToUUID(row.TenantID),
		BranchID:  pgtypeUUIDToUUID(row.BranchID),
		PatternID: pgtypeUUIDToUUID(row.PatternID),
		DayOfWeek: int(row.DayOfWeek),
		SessionType: &domain.EntrySessionType{
			ID:           pgtypeUUIDToUUID(row.SessionTypeID),
			Name:         row.SessionTypeName,
			StartMinutes: pgtypeTimeToMinutes(row.SessionTypeStartTime),
			EndMinutes:   pgtypeTimeToMinutes(row.SessionTypeEndTime),
			IsActive:     row.SessionTypeIsActive,
		},
		CreatedAt: pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt: pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func pgtypeTimeToMinutes(t pgtype.Time) int {
	if !t.Valid {
		return 0
	}
	return int(t.Microseconds / 60 / 1_000_000)
}
