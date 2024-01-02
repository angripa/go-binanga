package merchant

import (
	"binanga/internal/merchant/model"
)

type MerchantResponse struct {
	Merchant Merchant `json:"merchant"`
}

type MerchantsResponse struct {
	Merchant       []Merchant `json:"merchants"`
	MerchantsCount int64      `json:"merchantsCount"`
}

type Merchant struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	DeletedAt int64  `json:"deleted_at"`
}

// NewMerchantsResponse converts merchant models and total count to MerchantsResponse
func NewMerchantsResponse(merchants []*model.Merchant, total int64) *MerchantsResponse {
	var a []Merchant
	for _, merchant := range merchants {
		a = append(a, NewMerchantResponse(merchant).Merchant)
	}

	return &MerchantsResponse{
		Merchant:       a,
		MerchantsCount: total,
	}
}

// NewMerchantResponse converts merchant model to MerchantResponse
func NewMerchantResponse(a *model.Merchant) *MerchantResponse {

	return &MerchantResponse{
		Merchant: Merchant{
			ID:        a.ID,
			Name:      a.Name,
			CreatedAt: a.CreatedAt.Unix(),
			DeletedAt: a.DeletedAt,
		},
	}
}
