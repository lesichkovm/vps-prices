# VPS Prices

![VPS Price Comparison Banner](assets/banner.png)

A simple, interactive dashboard for comparing Virtual Private Server (VPS) prices, hardware specs (CPU, Memory, Disk), and price-to-performance across major cloud infrastructure providers.

Live Web Application: [https://lesichkovm.github.io/vps-prices/](https://lesichkovm.github.io/vps-prices/)

## Goal

The primary goal of this repository is to collect, normalize, and present up-to-date pricing and technical specifications for Virtual Private Servers across top cloud hosting providers.

Choosing a VPS can be overwhelming due to varying specs, tier structures, and pricing models. This project provides a centralized, searchable, and filterable table allowing developers, sysadmins, and organizations to quickly identify the best VPS plan for their workload and budget.

### Currency Standardization

<!-- FX-RATES:START -->
To allow a fair, apples-to-apples comparison across all providers, all prices originally quoted in Euros (EUR) or British Pounds (GBP) are converted to US Dollars (USD). The exchange rates fetched on 2026-10-05 from European Central Bank are:

- **1 EUR = 1.12 USD**
- **1 GBP = 1.32 USD**
<!-- FX-RATES:END -->

## Researched Providers

Below is the list of VPS hosting providers included in our comparison along with links to their official websites:

- [AWS Lightsail](https://aws.amazon.com/lightsail/)
- [CloudFanatic](https://cloudfanatic.net/)
- [Contabo](https://contabo.com/)
- [DigitalOcean](https://www.digitalocean.com/)
- [eVPS](https://www.evps.net/packages)
- [Hetzner](https://www.hetzner.com/)
- [Hostinger](https://www.hostinger.com/)
- [IONOS](https://www.ionos.com/)
- [Kamatera](https://www.kamatera.com/)
- [Linode (Akamai)](https://www.linode.com/)
- [Netcup](https://www.netcup.de/)
- [OVHcloud](https://www.ovhcloud.com/)
- [RackNerd](https://www.racknerd.com/)
- [Raff Technologies](https://rafftechnologies.com/)
- [Scaleway](https://www.scaleway.com/)
- [UpCloud](https://upcloud.com/)
- [Vultr](https://www.vultr.com/)

Detailed research notes and data sources are maintained in the [`research/`](research/) directory.

## Automated Updates

VPS pricing and plan data are automatically updated via a scheduled GitHub Action and local Go CLI tool (`vpsprices`).

### How the Workflow Runs
- **Schedule:** Runs automatically on a monthly schedule (1st of every month at 06:00 UTC) or manually via `workflow_dispatch`.
- **Process:** The workflow fetches daily exchange rates from the ECB, queries provider APIs or public catalogs, applies safety checks, regenerates `data.json` and `research/*.md`, and opens a Pull Request on GitHub for review.

### Running Locally
To run the price updater locally:

```bash
# Perform a dry-run (no files written)
go run ./cmd/vpsprices update --dry-run

# Run update for all providers
go run ./cmd/vpsprices update --report report.md

# Run update for a specific provider
go run ./cmd/vpsprices update --provider Linode --report report.md

# Validate data.json offline
go run ./cmd/vpsprices validate
```

### Optional Secrets
Certain provider APIs require read-only API tokens. Set these environment variables locally or as GitHub repository secrets:
- `HCLOUD_TOKEN`: Read-only token for Hetzner Cloud API (`https://api.hetzner.cloud/v1/server_types`).
- `DIGITALOCEAN_TOKEN`: Read-only token for DigitalOcean API (`https://api.digitalocean.com/v2/sizes`).

If a token is omitted, the tool logs a warning, skips network fetching for that provider, and preserves its existing data in `data.json`.

### Manual Providers
- **Contabo:** Contabo blocks automated requests with Cloudflare anti-bot protection (HTTP 403). It is marked as a `manual` provider, meaning its data is retained from existing records until updated manually.

### Adding a New Provider Adapter
To add support for a new VPS provider:
1. Create a new file under `internal/providers/<provider_name>.go`.
2. Define a struct implementing the `model.Provider` interface:
   ```go
   type Provider interface {
       Name() string
       Fetch(ctx context.Context) ([]model.Plan, error)
   }
   ```
3. Parse the provider's API or catalog, normalizing RAM/Disk to GB (float64 formatted string) and CPU to core count string.
4. Register the new provider adapter in `internal/providers/registry.go`.
5. Add fixture response data in `testdata/` and unit tests in `internal/providers/providers_test.go`.

## Project Structure

- `cmd/vpsprices/`: Entry point for the Go CLI (`update` and `validate` commands).
- `internal/`: Core logic including FX rate fetching, provider adapters, worker pool runner, and validation safety guards.
- `index.html`: Interactive Bootstrap Table interface with price tier filtering and search capabilities.
- `data.json`: Structured database containing VPS provider plans, CPU cores, RAM, storage size, and monthly pricing in USD.
- `research/`: Markdown records documenting verification dates, pricing tables, and links for each provider.
- `testdata/`: Offline fixture files for testing provider adapters.

## License

GPL-3.0
