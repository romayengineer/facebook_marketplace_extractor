module github.com/romayengineer/fme/go/capture

go 1.25.0

require (
	github.com/mxschmitt/playwright-go v0.6201.1
	github.com/romayengineer/fme/go/browser v0.0.0
	github.com/romayengineer/fme/go/core v0.0.0
	github.com/romayengineer/fme/go/marketplace v0.0.0
	github.com/romayengineer/fme/go/store v0.0.0
)

require (
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/deckarep/golang-set/v2 v2.8.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/saintfish/chardet v0.0.0-20230101081208-5e3ef4b5456d // indirect
	golang.org/x/text v0.37.0 // indirect
)

replace github.com/romayengineer/fme/go/browser => ../browser

replace github.com/romayengineer/fme/go/core => ../core

replace github.com/romayengineer/fme/go/marketplace => ../marketplace

replace github.com/romayengineer/fme/go/store => ../store
