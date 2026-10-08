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

func TestHostingerProvider(t *testing.T) {
	fixture := getFixture(t, "hostinger.html")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewHostingerProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Hostinger Fetch error: %v", err)
	}

	if len(plans) != 4 {
		t.Fatalf("Expected 4 Hostinger plans, got %d", len(plans))
	}

	if plans[0].Provider != "Hostinger" || plans[0].Price != "19.49" {
		t.Errorf("Unexpected Hostinger plan 0: %+v", plans[0])
	}
}

func TestNetcupProvider(t *testing.T) {
	fixture := getFixture(t, "netcup.html")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewNetcupProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Netcup Fetch error: %v", err)
	}

	if len(plans) != 5 {
		t.Fatalf("Expected 5 Netcup plans, got %d", len(plans))
	}

	if plans[0].Provider != "Netcup" || plans[0].Currency != "eur" {
		t.Errorf("Unexpected Netcup plan 0: %+v", plans[0])
	}
}

func TestIONOSProvider(t *testing.T) {
	fixture := getFixture(t, "ionos.html")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewIONOSProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("IONOS Fetch error: %v", err)
	}

	if len(plans) != 6 {
		t.Fatalf("Expected 6 IONOS plans, got %d", len(plans))
	}
}

func TestUpCloudProvider(t *testing.T) {
	fixture := getFixture(t, "upcloud.html")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewUpCloudProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("UpCloud Fetch error: %v", err)
	}

	if len(plans) != 6 {
		t.Fatalf("Expected 6 UpCloud plans, got %d", len(plans))
	}
}

func TestRackNerdProvider(t *testing.T) {
	fixture := getFixture(t, "racknerd.html")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewRackNerdProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("RackNerd Fetch error: %v", err)
	}

	if len(plans) != 7 {
		t.Fatalf("Expected 7 RackNerd plans, got %d", len(plans))
	}
}

func TestKamateraProvider(t *testing.T) {
	fixture := getFixture(t, "kamatera.html")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewKamateraProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Kamatera Fetch error: %v", err)
	}

	if len(plans) != 6 {
		t.Fatalf("Expected 6 Kamatera plans, got %d", len(plans))
	}
}

func TestVPSMartProvider(t *testing.T) {
	fixture := getFixture(t, "vpsmart.html")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	provider := NewVPSMartProvider(client, server.URL)

	plans, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("VPSMart Fetch error: %v", err)
	}

	if len(plans) != 8 {
		t.Fatalf("Expected 8 VPSMart plans, got %d", len(plans))
	}

	if plans[0].Provider != "VPSMart" || plans[0].Price != "2.88" {
		t.Errorf("Unexpected VPSMart plan 0: %+v", plans[0])
	}
}
