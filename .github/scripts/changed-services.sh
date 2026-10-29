#!/usr/bin/env bash
# Prints the images that need rebuilding between two commits, one per line.
#
# Usage: changed-services.sh <base-sha> [head-sha]
#
# An empty or unusable base (first run, force-push, manual dispatch) rebuilds
# everything. Go services build from backend/shared plus their own folder, so
# a change in shared rebuilds all of them.
set -euo pipefail

GO_SERVICES=(auth user finance transcode livestream chat vod realtime)
ALL_SERVICES=("${GO_SERVICES[@]}" configserver web)

base="${1:-}"
head="${2:-HEAD}"

print_all() {
    printf '%s\n' "${ALL_SERVICES[@]}"
    exit 0
}

if [[ -z "$base" ]] ||
    ! git cat-file -e "${base}^{commit}" 2>/dev/null ||
    ! git merge-base --is-ancestor "$base" "$head"; then
    print_all
fi

changed="$(git diff --name-only "$base" "$head")"

touched() {
    grep -qE "$1" <<<"$changed"
}

if touched '^\.github/(workflows/build-and-publish-images\.yml|scripts/changed-services\.sh)$'; then
    print_all
fi

shared_changed=false
touched '^backend/shared/' && shared_changed=true

for service in "${GO_SERVICES[@]}"; do
    if $shared_changed || touched "^backend/${service}/"; then
        echo "$service"
    fi
done

touched '^backend/configserver/' && echo configserver
touched '^web/' && echo web

exit 0
