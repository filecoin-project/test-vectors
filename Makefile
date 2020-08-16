GOTFLATS ?=
SHELL = /bin/bash

.PHONY: gen

gen:
	find gen/suites -type d -print0 -maxdepth 1 -mindepth 1 | xargs -I '{}' -n1 -0 bash -c 'dir="$$(basename {})" && echo "=== $${dir} ===" && cd {} && go run . -o "../../../corpus/$${dir}"'
