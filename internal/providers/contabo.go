package providers

import (
	"context"
	"errors"

	"github.com/lesichkovm/vps-prices/internal/model"
)

var ErrManualProvider = errors.New("provider requires manual review due to anti-bot protection or missing public API")

type ContaboProvider struct {
	client *HTTPClient
	url    string
}

func NewContaboProvider(client *HTTPClient, customURL string) *ContaboProvider {
	u := "https://contabo.com/en/vps/"
	if customURL != "" {
		u = customURL
	}
	return &ContaboProvider{client: client, url: u}
}

func (p *ContaboProvider) Name() string {
	return "Contabo"
}

func (p *ContaboProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	return nil, ErrManualProvider
}
