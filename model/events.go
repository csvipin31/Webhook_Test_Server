package model


type EventOptions struct {
    ExternalOrderId *string
    DealId *string
}

// BaseEvent struct holds common fields for all events.
type BaseEvent struct {
    Type         string `json:"$type"`
    EventId      string `json:"eventId"`
    LastUpdated  string `json:"lastUpdated"`
}

// EventTypeHolder is used to decode the JSON to determine the event type.
type EventTypeHolder struct {
    Type string `json:"$type"`
}

// OrderCreationFailed event represents an order creation failure event.
type OrderCreationFailed struct {
    BaseEvent
    ExternalOrderID string `json:"externalOrderId"`
    Errors          []struct {
        Code    string `json:"code"`
        Message string `json:"message"`
    } `json:"errors"`
}

// OrderCreated event represents an order created event.
type OrderCreated struct {
    BaseEvent
    ExternalOrderID string 			`json:"externalOrderId"`
    Details         []OrderDetail 	`json:"details"`
}

type OrderDetail struct {
    ExternalOrderGroupID string `json:"externalOrderGroupId"`
    ExternalOrderLineID  string `json:"externalOrderLineId"`
    Type                 string `json:"type"`
    InternalID           string `json:"internalId"`
}

// VariantStockUpdated event represents a variant stock update event.
type VariantStockUpdated struct {
    BaseEvent
    DealID    string      `json:"dealId"`
    VariantID interface{} `json:"variantId"`
    Stock     int         `json:"stock"`
}

// ProductUpdateV2 event represents a product update v2 event.
type ProductUpdateV2 struct {
    BaseEvent
	DealID             string      `json:"dealId"`
	CompanyID          string      `json:"companyId"`
	CategoryID         string      `json:"categoryId"`
	Type               string      `json:"type"`
	Name               string      `json:"name"`
	Specification      interface{} `json:"specification"`
	Details            string      `json:"details"`
	Brand              interface{} `json:"brand"`
	TextTags           string      `json:"textTags"`
	IncludesGst        bool        `json:"includesGst"`
	IsPOBoxDeliverable bool        `json:"isPOBoxDeliverable"`
	Images             []string    `json:"images"`
	Gtin               interface{}      `json:"gtin"`
	Mpn                interface{} `json:"mpn"`
	IsFreeShipping     bool        `json:"isFreeShipping"`
}

type ProductSubscribed struct {
	BaseEvent
	DealID             string      `json:"dealId"`
	CompanyID          string      `json:"companyId"`
	CategoryID         string      `json:"categoryId"`
	Type               string      `json:"type"`
	Name               string      `json:"name"`
	Specification      interface{} `json:"specification"`
	Details            string      `json:"details"`
	Brand              interface{} `json:"brand"`
	TextTags           string      `json:"textTags"`
	IncludesGst        bool        `json:"includesGst"`
	IsPOBoxDeliverable bool        `json:"isPOBoxDeliverable"`
	Images             []string    `json:"images"`
	Variants           []struct {
		VariantID      string   `json:"variantId"`
		VariantOptions []struct {
			VariantName  string `json:"variantName"`
			VariantValue string `json:"variantValue"`
		} `json:"variantOptions"`
		Image          interface{}   `json:"image"`
		Gtin           interface{}   `json:"gtin"`
		Mpn            string        `json:"mpn"`
		Price          int           `json:"price"`
		Stock          int           `json:"stock"`
		Promotion struct {
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Price     int       `json:"price"`
		} `json:"promotion"`
	} `json:"variants"`
}

// PriceUpdate event represents a price update event.
type PriceUpdate struct {
    BaseEvent
    DealID    string      `json:"dealId"`
    VariantID interface{} `json:"variantId"`
    Price     int         `json:"price"`
}
