package http

import siteprofiledomain "nursery-management-system/api/internal/modules/siteprofile/domain"

type SiteProfileResponse struct {
	NurseryName     string `json:"nursery_name"`
	Description     string `json:"description"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Website         string `json:"website"`
	AddressStreet   string `json:"address_street"`
	AddressCity     string `json:"address_city"`
	AddressPostcode string `json:"address_postcode"`
}

type getSiteProfileResponse struct {
	SiteProfile *SiteProfileResponse `json:"site_profile"`
}

type updateSiteProfileRequest struct {
	NurseryName     string `json:"nursery_name" binding:"required"`
	Description     string `json:"description" binding:"required"`
	Phone           string `json:"phone" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Website         string `json:"website" binding:"required"`
	AddressStreet   string `json:"address_street" binding:"required"`
	AddressCity     string `json:"address_city" binding:"required"`
	AddressPostcode string `json:"address_postcode" binding:"required"`
}

func toSiteProfileResponse(profile *siteprofiledomain.SiteProfile) *SiteProfileResponse {
	if profile == nil {
		return nil
	}
	return &SiteProfileResponse{
		NurseryName:     profile.NurseryName,
		Description:     profile.Description,
		Phone:           profile.Phone,
		Email:           profile.Email,
		Website:         profile.Website,
		AddressStreet:   profile.AddressStreet,
		AddressCity:     profile.AddressCity,
		AddressPostcode: profile.AddressPostcode,
	}
}
