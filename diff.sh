#!/bin/bash

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
