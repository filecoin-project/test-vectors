#!/bin/bash

set -o errexit
set -o pipefail
set -e

REPO=$1
BRANCH=$2

# Validate required arguments
if [ -z "$REPO" ]; then
  echo -e "Please provider REPO input variable. For example: \`./diff.sh lotus next\`"
  exit 2
fi
if [ -z "$BRANCH" ]; then
  echo -e "Please provider BRANCH input variable. For example: \`./diff.sh lotus next\`"
  exit 2
fi

pushd gen
go get github.com/filecoin-project/${REPO}@${BRANCH}
popd

tmp_dir=$(mktemp -d -t ci-compare-corpus-XXXX)
echo "temporary dir: $tmp_dir"

cp -R corpus corpus-current
cp -R corpus-current $tmp_dir/corpus-current
make gen
mv corpus $tmp_dir/corpus-new
mv corpus-current corpus

pushd $tmp_dir

# Update in-place all newly generated JSON corpus files and remove the "_meta" field.
for f in $(find ./corpus-new -type f); do
  echo "$(jq 'del(._meta)' $f)" > $f;
done

# Update in-place all existing JSON corpus files and remove the "_meta" field.
for f in $(find ./corpus-current -type f); do
  echo "$(jq 'del(._meta)' $f)" > $f;
done

# Run diff between the old and the new test vectors
diff -qr corpus-new/ corpus-current/

popd

rm -rf $tmp_dir
