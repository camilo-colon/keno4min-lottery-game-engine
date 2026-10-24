.PHONY: build-DrawVideoFunction

build-DrawVideoFunction:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -trimpath -ldflags="-s -w" -o "$(ARTIFACTS_DIR)/bootstrap" ./cmd/lambda
