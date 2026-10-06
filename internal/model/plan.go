package model

import "context"

// Plan represents a single VPS plan record mapping to the data.json schema,
// along with internal fields for price conversion and research notes.
type Plan struct {
	Provider  string `json:"provider"`
	Memory    string `json:"memory"`
	CPU       string `json:"cpu"`
	Disk      string `json:"disk"`
	Price     string `json:"price"`
	Currency  string `json:"currency,omitempty"`
	UpdatedAt string `json:"updated_at"`

	// Internal fields not directly serialized to data.json
	PlanID        string  `json:"-"`
	OriginalPrice float64 `json:"-"`
	USDPrice      float64 `json:"-"`
	Region        string  `json:"-"`
	SourceURL     string  `json:"-"`
}

// Provider is the interface each VPS hosting provider adapter must implement.
type Provider interface {
	Name() string
	Fetch(ctx context.Context) ([]Plan, error)
}
