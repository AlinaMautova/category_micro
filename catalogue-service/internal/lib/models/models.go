package models

type Item struct {
	ID        int32  `json:"id,omitempty"`
	Name      string `json:"product_name,omitempty"`
	Price     int32  `json:"price,omitempty"`
	Type      string `json:"type,omitempty"`
	Quantity  int32  `json:"quantity,omitempty"`
	Photo_url string `json:"photo_url"`
}
