GOTFLATS ?=
SHELL = /bin/bash
GENCOMMIT = `git rev-list -1 HEAD`

.PHONY: gen validate

gen:
	find gen/suites -maxdepth 1 -mindepth 1 -type d -print0 | xargs -I '{}' -n1 -0 bash -c 'dir="$$(basename {})" && echo "=== $${dir} ===" && cd {} && go run -ldflags "-X github.com/filecoin-project/test-vectors/gen/builders.GenscriptCommit=${GENCOMMIT}" . -o "../../../corpus/$${dir}"'

compare-with-lotus-next:
	pushd gen && go get github.com/filecoin-project/specs-actors@master && go get github.com/filecoin-project/lotus@a0c0d9c98aae && popd && cp -R corpus corpus-current && make gen && mv corpus corpus-new && ./diff.sh

validate:
	cd ./validate && go run ./main.go
