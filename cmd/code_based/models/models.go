package models

type (
	Order struct {
		ProductID       string `json:"product_id"`
		CustomerID      string `json:"customer_id"`
		ShippingAddress string `json:"shipping_address"`
	}
	OrderResponse struct {
		Order  *Order `json:"order"`
		Status string `json:"status"`
	}
)
