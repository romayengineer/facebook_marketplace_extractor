set shell := ["bash", "-cu"]

binary := "facebook-extractor"
src := "src"

# List available recipes
default:
    @just --list

# Download and tidy module dependencies
tidy:
    go mod tidy

# Format Go source
fmt:
    go fmt ./...

# Run the Go vet static analyzer
vet:
    go vet ./...

# Run golangci-lint
lint:
    golangci-lint run ./...

# Run the test suite
test:
    go test ./...

# Run the test suite with coverage
cover:
    go test -cover ./...

# Build the scraper binary
build: tidy
    go build -o {{binary}} ./{{src}}

# Install the scraper binary into $GOBIN (defaults to $GOPATH/bin)
install: build
    #!/usr/bin/env bash
    set -euo pipefail
    dest="${GOBIN:-$(go env GOPATH)/bin}"
    mkdir -p "$dest"
    install -m 0755 {{binary}} "$dest/{{binary}}"
    echo "installed $dest/{{binary}}"

# Run the scraper with arbitrary flags, e.g. `just run -action=serve`
run *args:
    go run ./{{src}} {{args}}

# Search Marketplace for a keyword, e.g. `just search "macbook m1"`
search keyword *args:
    go run ./{{src}} -action=search -keywords="{{keyword}}" {{args}}

# Pull descriptions for all details missing one (opens any listing to grab the PDP template)
pull-description *args:
    go run ./{{src}} -action=pull_description {{args}}

# Remove duplicate PDP responses, keeping the newest per product id
dedupe-pdp:
    go run ./{{src}} -action=dedupe_pdp

# Visit each known listing to fetch details
get-details:
    go run ./{{src}} -action=get_details

# Extract products from scraped responses into the database
save:
    go run ./{{src}} -action=save

# Extract listings from raw responses into data/__Details (no descriptions)
process-data:
    go run ./{{src}} -action=process_data

# Backfill missing URLs on existing detail files
fill-empty:
    go run ./{{src}} -action=fill_empty

# Serve the PocketBase admin/web UI
serve:
    go run ./{{src}} -action=serve

# Archive the scraped data directory
backup:
    #!/usr/bin/env bash
    set -euo pipefail
    time 7z a "../data_$(date +%s).zip" data/

# Create the Python virtualenv for the ml pipeline
ml-venv:
    python3 -m venv ml/.venv

# Install ml pipeline dependencies
ml-install: ml-venv
    ml/.venv/bin/pip install -r ml/requirements.txt

# Train the price prediction models
ml-train:
    cd ml && .venv/bin/python train.py

# Remove build artifacts
clean:
    rm -f {{binary}}
