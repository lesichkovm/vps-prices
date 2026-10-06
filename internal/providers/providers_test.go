package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func getFixture(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read fixture %s: %v", filename, err)
	}
	return data
}

func TestLinodeProvider(t *testing.T) {
	fixture := getFixture(t, "linode_types.json")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewLinodeProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Linode Fetch error: %v", err)
	}

	if len(plans) == 0 {
		t.Fatal("Expected Linode plans, got 0")
	}

	p0 := plans[0]
	if p0.Provider != "Linode" || p0.Memory != "1" || p0.CPU != "1" || p0.Disk != "25" || p0.Price != "5.00" {
		t.Errorf("Unexpected Linode plan 0: %+v", p0)
	}
}

func TestVultrProvider(t *testing.T) {
	fixture := getFixture(t, "vultr_plans.json")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewVultrProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Vultr Fetch error: %v", err)
	}

	if len(plans) == 0 {
		t.Fatal("Expected Vultr plans, got 0")
	}

	hasIPv6 := false
	for _, p := range plans {
		if p.Provider == "Vultr (IPv6)" {
			hasIPv6 = true
			break
		}
	}
	if !hasIPv6 {
		t.Error("Expected Vultr (IPv6) plan in Vultr results")
	}
}

func TestAWSLightsailProvider(t *testing.T) {
	fixture := getFixture(t, "aws_lightsail.json")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewAWSLightsailProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("AWS Lightsail Fetch error: %v", err)
	}

	if len(plans) == 0 {
		t.Fatal("Expected AWS Lightsail plans, got 0")
	}
}

func TestHetznerProvider(t *testing.T) {
	fixture := getFixture(t, "hetzner_server_types.json")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	t.Setenv("HCLOUD_TOKEN", "dummy-token")

	client := NewHTTPClient(5 * time.Second)
	provider := NewHetznerProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Hetzner Fetch error: %v", err)
	}

	if len(plans) != 2 {
		t.Fatalf("Expected 2 Hetzner plans, got %d", len(plans))
	}
}

func TestDigitalOceanProvider(t *testing.T) {
	fixture := getFixture(t, "digitalocean_sizes.json")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	t.Setenv("DIGITALOCEAN_TOKEN", "dummy-token")

	client := NewHTTPClient(5 * time.Second)
	provider := NewDigitalOceanProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("DigitalOcean Fetch error: %v", err)
	}

	if len(plans) != 2 {
		t.Fatalf("Expected 2 DigitalOcean plans, got %d", len(plans))
	}
}
