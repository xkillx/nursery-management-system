package application

import (
	"testing"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

func strPtr(s string) *string { return &s }

func TestSubmitCompleteRegistrationValidateInput_RequiresPrimaryRoomAndRegistrationDate(t *testing.T) {
	uc := &SubmitCompleteRegistration{}

	cases := []struct {
		name           string
		input          domain.CompleteRegistrationInput
		wantCode       string
		wantField      string
		wantMessageSub string
	}{
		{
			name: "missing primary room and registration date",
			input: domain.CompleteRegistrationInput{
				Child: domain.ChildRegistrationInfo{
					FirstName:   "James",
					DateOfBirth: "2022-01-01",
					StartDate:   "2026-09-01",
				},
				Profile: domain.ProfileSectionsInput{
					RegistrationDate: nil,
				},
				Consents: domain.ConsentInput{
					SafeguardingReportingAcknowledgement: true,
				},
			},
			wantCode:       "validation_error",
			wantField:      "primary_room_id",
			wantMessageSub: "primary room",
		},
		{
			name: "missing registration date",
			input: domain.CompleteRegistrationInput{
				Child: domain.ChildRegistrationInfo{
					FirstName:     "James",
					DateOfBirth:   "2022-01-01",
					StartDate:     "2026-09-01",
					PrimaryRoomID: strPtr("11111111-1111-1111-1111-111111111111"),
				},
				Profile: domain.ProfileSectionsInput{
					RegistrationDate: nil,
				},
				Consents: domain.ConsentInput{
					SafeguardingReportingAcknowledgement: true,
				},
			},
			wantCode:       "validation_error",
			wantField:      "registration_date",
			wantMessageSub: "registration date",
		},
		{
			name: "blank primary room string",
			input: domain.CompleteRegistrationInput{
				Child: domain.ChildRegistrationInfo{
					FirstName:     "James",
					DateOfBirth:   "2022-01-01",
					StartDate:     "2026-09-01",
					PrimaryRoomID: strPtr("   "),
				},
				Profile: domain.ProfileSectionsInput{
					RegistrationDate: strPtr("2026-06-17"),
				},
				Consents: domain.ConsentInput{
					SafeguardingReportingAcknowledgement: true,
				},
			},
			wantCode:       "validation_error",
			wantField:      "primary_room_id",
			wantMessageSub: "primary room",
		},
		{
			name: "malformed primary room uuid",
			input: domain.CompleteRegistrationInput{
				Child: domain.ChildRegistrationInfo{
					FirstName:     "James",
					DateOfBirth:   "2022-01-01",
					StartDate:     "2026-09-01",
					PrimaryRoomID: strPtr("not-a-uuid"),
				},
				Profile: domain.ProfileSectionsInput{
					RegistrationDate: strPtr("2026-06-17"),
				},
				Consents: domain.ConsentInput{
					SafeguardingReportingAcknowledgement: true,
				},
			},
			wantCode:       "validation_error",
			wantField:      "primary_room_id",
			wantMessageSub: "primary room",
		},
		{
			name: "malformed registration date",
			input: domain.CompleteRegistrationInput{
				Child: domain.ChildRegistrationInfo{
					FirstName:     "James",
					DateOfBirth:   "2022-01-01",
					StartDate:     "2026-09-01",
					PrimaryRoomID: strPtr("11111111-1111-1111-1111-111111111111"),
				},
				Profile: domain.ProfileSectionsInput{
					RegistrationDate: strPtr("2026-13-99"),
				},
				Consents: domain.ConsentInput{
					SafeguardingReportingAcknowledgement: true,
				},
			},
			wantCode:       "validation_error",
			wantField:      "registration_date",
			wantMessageSub: "registration date",
		},
		{
			name: "happy path",
			input: domain.CompleteRegistrationInput{
				Child: domain.ChildRegistrationInfo{
					FirstName:     "James",
					DateOfBirth:   "2022-01-01",
					StartDate:     "2026-09-01",
					PrimaryRoomID: strPtr("11111111-1111-1111-1111-111111111111"),
				},
				Profile: domain.ProfileSectionsInput{
					RegistrationDate: strPtr("2026-06-17"),
				},
				Consents: domain.ConsentInput{
					SafeguardingReportingAcknowledgement: true,
				},
			},
			wantCode: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := uc.validateInput(tc.input)
			if tc.wantCode == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			domainErr, ok := err.(*domainerrors.DomainError)
			if !ok {
				t.Fatalf("expected *DomainError, got %T", err)
			}
			if domainErr.Code != tc.wantCode {
				t.Fatalf("expected code %q, got %q", tc.wantCode, domainErr.Code)
			}
			if domainErr.Message == "" {
				t.Fatal("expected non-empty message")
			}
			fields, ok := domainErr.Details["field_errors"].([]domainerrors.FieldError)
			if !ok {
				t.Fatalf("expected field_errors detail, got %v", domainErr.Details)
			}
			found := false
			for _, f := range fields {
				if f.Field == tc.wantField {
					found = true
					if !contains(f.Message, tc.wantMessageSub) {
						t.Fatalf("expected message to contain %q, got %q", tc.wantMessageSub, f.Message)
					}
				}
			}
			if !found {
				t.Fatalf("expected field_error for %q, got %+v", tc.wantField, fields)
			}
		})
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || indexOf(haystack, needle) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
