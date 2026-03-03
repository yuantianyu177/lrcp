#!/usr/bin/env bash
# Build release binaries for multiple platforms with minimal size.

set -euo pipefail

APP="lrcp"
VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo "dev")}"
OUTPUT_DIR="dist"
LDFLAGS="-s -w -X github.com/dislab/lrcp/cmd.Version=${VERSION}"

TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

for target in "${TARGETS[@]}"; do
    GOOS="${target%/*}"
    GOARCH="${target#*/}"
    output="${OUTPUT_DIR}/${APP}-${GOOS}-${GOARCH}"
    if [[ "${GOOS}" == "windows" ]]; then
        output="${output}.exe"
    fi

    echo "Building ${GOOS}/${GOARCH}..."
    CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" \
        go build -trimpath -ldflags "${LDFLAGS}" -o "${output}" .
done

echo ""
echo "Build complete:"
ls -lh "${OUTPUT_DIR}"/
