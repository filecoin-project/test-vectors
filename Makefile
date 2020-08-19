GOTFLATS ?=
SHELL = /bin/bash
GENCOMMIT = `git rev-list -1 HEAD`

.PHONY: gen validate

gen:
	go generate ./...
	find gen/suites -maxdepth 1 -mindepth 1 -type d -print0 | xargs -I '{}' -n1 -0 bash -c 'dir="$$(basename {})" && echo "=== $${dir} ===" && cd {} && go run -ldflags "-X github.com/filecoin-project/test-vectors/gen/builders.GenscriptCommit=${GENCOMMIT}" . -o "../../../corpus/$${dir}"'

validate:
	go run ./cmd/validate
