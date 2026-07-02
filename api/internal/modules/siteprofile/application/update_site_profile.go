package application

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/siteprofile/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

var ukPostcodeRegex = regexp.MustCompile(`^([A-Z]{1,2}\d[A-Z\d]?\s*\d[A-Z]{2})$`)

type UpdateSiteProfileInput struct {
	NurseryName     string `json:"nursery_name"`
	Description     string `json:"description"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Website         string `json:"website"`
	AddressStreet   string `json:"address_street"`
	AddressCity     string `json:"address_city"`
	AddressPostcode string `json:"address_postcode"`
}

type txManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type auditWriter interface {
	WriteWithTx(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, params audit.WriteParams) error
}

type UpdateSiteProfileUseCase struct {
	repo        domain.Repository
	auditWriter auditWriter
	txMgr       txManager
}

func NewUpdateSiteProfileUseCase(
	repo domain.Repository,
	auditWriter auditWriter,
	txMgr txManager,
) *UpdateSiteProfileUseCase {
	return &UpdateSiteProfileUseCase{
		repo:        repo,
		auditWriter: auditWriter,
		txMgr:       txMgr,
	}
}

func (uc *UpdateSiteProfileUseCase) Execute(ctx context.Context, actor tenant.ActorContext, input UpdateSiteProfileInput) (*domain.SiteProfile, error) {
	input = trimInput(input)

	if errs := validateInput(input); len(errs) > 0 {
		return nil, domain.ValidationError(errs)
	}

	var saved *domain.SiteProfile

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		profile := domain.SiteProfile{
			ID:              uid.NewUUID(),
			TenantID:        actor.TenantID,
			BranchID:        actor.BranchID,
			NurseryName:     input.NurseryName,
			Description:     input.Description,
			Phone:           input.Phone,
			Email:           input.Email,
			Website:         input.Website,
			AddressStreet:   input.AddressStreet,
			AddressCity:     input.AddressCity,
			AddressPostcode: input.AddressPostcode,
		}

		if err := uc.repo.Upsert(ctx, tx, actor.TenantID, actor.BranchID, profile); err != nil {
			return err
		}

		saved = &profile

		if auditErr := uc.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "site_profile_updated",
			EntityType: "branch",
			EntityID:   actor.BranchID,
			Details: map[string]any{
				"nursery_name":     input.NurseryName,
				"description":      input.Description,
				"phone":            input.Phone,
				"email":            input.Email,
				"website":          input.Website,
				"address_street":   input.AddressStreet,
				"address_city":     input.AddressCity,
				"address_postcode": input.AddressPostcode,
			},
		}); auditErr != nil {
			return auditErr
		}

		return nil
	})

	if txErr != nil {
		if domainerrors.IsValidation(txErr) {
			return nil, txErr
		}
		return nil, txErr
	}

	return saved, nil
}

func trimInput(in UpdateSiteProfileInput) UpdateSiteProfileInput {
	website := strings.TrimSpace(in.Website)
	if website != "" && !strings.Contains(website, "://") {
		website = "https://" + website
	}
	return UpdateSiteProfileInput{
		NurseryName:     strings.TrimSpace(in.NurseryName),
		Description:     strings.TrimSpace(in.Description),
		Phone:           strings.TrimSpace(in.Phone),
		Email:           strings.TrimSpace(in.Email),
		Website:         website,
		AddressStreet:   strings.TrimSpace(in.AddressStreet),
		AddressCity:     strings.TrimSpace(in.AddressCity),
		AddressPostcode: strings.TrimSpace(in.AddressPostcode),
	}
}

func validateInput(in UpdateSiteProfileInput) []domain.FieldError {
	var errs []domain.FieldError

	if in.NurseryName == "" {
		errs = append(errs, domain.FieldError{Field: "nursery_name", Message: "Enter your nursery name."})
	} else if len(in.NurseryName) > 120 {
		errs = append(errs, domain.FieldError{Field: "nursery_name", Message: "Nursery name must be 120 characters or fewer."})
	}

	if in.Description == "" {
		errs = append(errs, domain.FieldError{Field: "description", Message: "Enter a description for your nursery."})
	} else if len(in.Description) > 2000 {
		errs = append(errs, domain.FieldError{Field: "description", Message: "Description must be 2000 characters or fewer."})
	}

	if in.Phone == "" {
		errs = append(errs, domain.FieldError{Field: "phone", Message: "Enter your phone number."})
	} else if len(in.Phone) > 32 {
		errs = append(errs, domain.FieldError{Field: "phone", Message: "Phone number must be 32 characters or fewer."})
	}

	if in.Email == "" {
		errs = append(errs, domain.FieldError{Field: "email", Message: "Enter your email address."})
	} else if len(in.Email) > 254 {
		errs = append(errs, domain.FieldError{Field: "email", Message: "Email must be 254 characters or fewer."})
	} else if !isValidEmail(in.Email) {
		errs = append(errs, domain.FieldError{Field: "email", Message: "Enter a valid email address (e.g. name@example.com)."})
	}

	if in.Website == "" {
		errs = append(errs, domain.FieldError{Field: "website", Message: "Enter your website address."})
	} else if len(in.Website) > 2048 {
		errs = append(errs, domain.FieldError{Field: "website", Message: "Website must be 2048 characters or fewer."})
	} else if !isValidURL(in.Website) {
		errs = append(errs, domain.FieldError{Field: "website", Message: "Enter a valid website address (e.g. https://www.example.com)."})
	}

	if in.AddressStreet == "" {
		errs = append(errs, domain.FieldError{Field: "address_street", Message: "Enter your street address."})
	} else if len(in.AddressStreet) > 200 {
		errs = append(errs, domain.FieldError{Field: "address_street", Message: "Street address must be 200 characters or fewer."})
	}

	if in.AddressCity == "" {
		errs = append(errs, domain.FieldError{Field: "address_city", Message: "Enter your city."})
	} else if len(in.AddressCity) > 100 {
		errs = append(errs, domain.FieldError{Field: "address_city", Message: "City must be 100 characters or fewer."})
	}

	if in.AddressPostcode == "" {
		errs = append(errs, domain.FieldError{Field: "address_postcode", Message: "Enter your postcode."})
	} else if len(in.AddressPostcode) > 16 {
		errs = append(errs, domain.FieldError{Field: "address_postcode", Message: "Postcode must be 16 characters or fewer."})
	} else if !ukPostcodeRegex.MatchString(strings.ToUpper(in.AddressPostcode)) {
		errs = append(errs, domain.FieldError{Field: "address_postcode", Message: "Enter a valid UK postcode (e.g. M1 4BT)."})
	}

	return errs
}

func isValidEmail(email string) bool {
	at := strings.LastIndex(email, "@")
	if at <= 0 || at >= len(email)-1 {
		return false
	}
	local := email[:at]
	domainPart := email[at+1:]
	if len(local) == 0 || len(domainPart) < 3 {
		return false
	}
	if !strings.Contains(domainPart, ".") {
		return false
	}
	return true
}

func isValidURL(raw string) bool {
	if raw != "" && !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}
