module github.com/opendatahub-io/odh-linter/linters/errordemote

go 1.23.1

require golang.org/x/tools v0.28.0

require (
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
)

// For local development
replace github.com/opendatahub-io/odh-linter/linters/errordemote => ./

