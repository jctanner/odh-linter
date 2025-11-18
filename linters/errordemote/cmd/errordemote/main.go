package main

import (
	"github.com/opendatahub-io/odh-linter/linters/errordemote"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(errordemote.Analyzer)
}

