#!/usr/bin/env bash
# rename_project.sh
# Usage: ./rename_project.sh <owner>/<repo>
# Example: ./rename_project.sh wilhelmrauston/kademlia
#
# What it does:
#  - Rewrites module path in go.mod to github.com/<owner>/<repo>
#  - Replaces occurrences of the old repo basename (e.g., "go-template") in file contents and names
#  - Rewrites old import prefixes (e.g., github.com/<oldowner>/ -> github.com/<owner>/)
#  - Skips itself and common vendor/VCS dirs
#  - Tolerates "no matches" without exiting

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <owner>/<repo>"
  exit 1
fi

if [[ ! -f go.mod ]]; then
  echo "ERROR: go.mod not found at repo root"
  exit 1
fi

# Derive current module info from go.mod
OLD_MODULE="$(awk '/^module[[:space:]]+/ {print $2; exit}' go.mod)"
if [[ -z "${OLD_MODULE}" ]]; then
  echo "ERROR: could not parse current module from go.mod"
  exit 1
fi

OLD_NAME="$(basename "${OLD_MODULE}")"        # e.g., go-template
OLD_PREFIX="$(dirname  "${OLD_MODULE}")/"     # e.g., github.com/eislab-cps/

# Desired new module info
NEW_NAME="$1"                                 # e.g., wilhelmrauston/kademlia
NEW_REPO_BASENAME="$(basename "${NEW_NAME}")" # e.g., kademlia
NEW_MODULE="github.com/${NEW_NAME}"
NEW_PREFIX="github.com/${NEW_NAME%%/*}/"      # e.g., github.com/wilhelmrauston/

echo "====================================="
echo "Renaming from module: '${OLD_MODULE}' -> '${NEW_MODULE}'"
echo "Project name token:   '${OLD_NAME}'   -> '${NEW_REPO_BASENAME}'"
echo "Prefix rewrite:       '${OLD_PREFIX}' -> '${NEW_PREFIX}'"
echo "====================================="

# sed wrapper (macOS vs Linux)
if [[ "${OSTYPE:-}" == darwin* ]]; then
  SED_INPLACE() { sed -i '' -e "$1" "$2"; }
else
  SED_INPLACE() { sed -i -e "$1" "$2"; }
fi

# Helper to list files safely (tolerate no matches)
list_matches() {
  local pattern="$1"
  grep -rl --null "${pattern}" . \
    --exclude-dir=.git \
    --exclude-dir=.github \
    --exclude-dir=vendor \
    --exclude-dir=bin \
    --exclude-dir=build \
    --exclude="$(basename "$0")" \
    --exclude=rename_project.sh \
    --exclude=rename-project.sh \
    || true
}

# Step 1: Replace old repo basename tokens in file contents
echo "Step 1: Updating file contents (project name tokens)..."
while IFS= read -r -d '' file; do
  echo "  • ${file}"
  SED_INPLACE "s#${OLD_NAME}#${NEW_REPO_BASENAME}#g" "${file}"
done < <(list_matches "${OLD_NAME}")

# Step 2: Replace exact old module path with new
echo "Step 2: Updating import paths (module path)..."
while IFS= read -r -d '' file; do
  echo "  • ${file}"
  SED_INPLACE "s#${OLD_MODULE}#${NEW_MODULE}#g" "${file}"
done < <(list_matches "${OLD_MODULE}")

# Step 2b: Rewrite owner prefix if it actually changed
if [[ "${OLD_PREFIX}" != "${NEW_PREFIX}" ]]; then
  echo "Step 2b: Rewriting owner prefix in import paths..."
  while IFS= read -r -d '' file; do
    echo "  • ${file}"
    SED_INPLACE "s#${OLD_PREFIX}#${NEW_PREFIX}#g" "${file}"
  done < <(list_matches "${OLD_PREFIX}")
else
  echo "Step 2b: Owner prefix unchanged; skipping."
fi

# Step 3: Rename files and directories containing the old basename
echo "Step 3: Renaming files and directories..."
while IFS= read -r path; do
  new_path="$(dirname "${path}")/$(basename "${path}" | sed "s#${OLD_NAME}#${NEW_REPO_BASENAME}#g")"
  if [[ "${path}" != "${new_path}" ]]; then
    echo "  • '${path}' -> '${new_path}'"
    mv "${path}" "${new_path}"
  fi
done < <(find . -depth -name "*${OLD_NAME}*")

# Step 4: Update go.mod module line explicitly
echo "Step 4: Updating 'module' line in go.mod..."
SED_INPLACE "s#^module[[:space:]].*#module ${NEW_MODULE}#g" go.mod

echo "====================================="
echo "✅ Renaming completed!"
echo "✅ Project name: '${NEW_REPO_BASENAME}'"
echo "✅ Module path:  '${NEW_MODULE}'"
echo "====================================="
echo "Next steps:"
echo " 1) go mod tidy"
echo " 2) go build ./..."
echo " 3) git add -A && git commit -m \"Rename module to ${NEW_MODULE}\""
