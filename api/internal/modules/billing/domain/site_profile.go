package domain

// ParentInvoiceSiteProfile is the site profile data attached to parent invoice detail responses.
type ParentInvoiceSiteProfile struct {
	NurseryName     string `json:"nursery_name"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Website         string `json:"website"`
	AddressStreet   string `json:"address_street"`
	AddressCity     string `json:"address_city"`
	AddressPostcode string `json:"address_postcode"`
}

// InvoiceSiteProfile is the site profile data attached to manager invoice detail responses.
type InvoiceSiteProfile struct {
	NurseryName     string `json:"nursery_name"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Website         string `json:"website"`
	AddressStreet   string `json:"address_street"`
	AddressCity     string `json:"address_city"`
	AddressPostcode string `json:"address_postcode"`
}
