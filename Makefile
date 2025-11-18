.PHONY: build build-go-linter build-bundle-linter

build: build-go-linter build-bundle-linter

build-go-linter:
	cd linters/odhlint && go build -o ./../../cmd/odhlint ./cmd/odhlint
	@echo "Built: cmd/odhlint"

build-bundle-linter:
	cd bundle-linters && go build -o odhlint-bundle ./cmd/odhlint-bundle
	@echo "Built: bundle-linters/odhlint-bundle"

