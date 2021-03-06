#!/usr/bin/env bash

# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

export GO111MODULE=on

source $(dirname "$0")/../vendor/knative.dev/hack/library.sh

# There is a directory named `internal`, so we need to move the backup files out
# of the repo root, otherwise go tools will complain that something is trying to
# import an internal package that it can't see. E.g.
# ${REPO_ROOT_DIR}/tmpdiffroot.abcdef/pkg/apis/messaging/v1beta1/deprecated_condition.go
# tries to import ${REPO_ROOT_DIR}/pkg/apis/messaging/internal.
readonly TMP_DIFFROOT="$(mktemp -d -t tmpdiffroot.XXXXXX)"

cleanup() {
  rm -rf "${TMP_DIFFROOT}"
}

trap "cleanup" EXIT SIGINT

cleanup

# Save working tree state
mkdir -p "${TMP_DIFFROOT}"

cp -aR \
  "${REPO_ROOT_DIR}/go.sum" \
  "${REPO_ROOT_DIR}/pkg" \
  "${REPO_ROOT_DIR}/third_party" \
  "${REPO_ROOT_DIR}/vendor" \
  "${TMP_DIFFROOT}"

"${REPO_ROOT_DIR}/hack/update-codegen.sh"
echo "Diffing ${REPO_ROOT_DIR} against freshly generated codegen"
ret=0

diff -Naupr --no-dereference \
  "${REPO_ROOT_DIR}/go.sum" "${TMP_DIFFROOT}/go.sum" || ret=1

diff -Naupr --no-dereference \
  "${REPO_ROOT_DIR}/pkg" "${TMP_DIFFROOT}/pkg" || ret=1

diff -Naupr --no-dereference \
  "${REPO_ROOT_DIR}/third_party" "${TMP_DIFFROOT}/third_party" || ret=1

diff -Naupr --no-dereference \
  "${REPO_ROOT_DIR}/vendor" "${TMP_DIFFROOT}/vendor" || ret=1

# Restore working tree state
rm -fr \
  "${REPO_ROOT_DIR}/go.sum" \
  "${REPO_ROOT_DIR}/pkg" \
  "${REPO_ROOT_DIR}/third_party" \
  "${REPO_ROOT_DIR}/vendor"

cp -aR "${TMP_DIFFROOT}"/* "${REPO_ROOT_DIR}"

if [[ $ret -eq 0 ]]
then
  echo "${REPO_ROOT_DIR} up to date."
 else
  echo "${REPO_ROOT_DIR} is out of date. Please run hack/update-codegen.sh"
  exit 1
fi
