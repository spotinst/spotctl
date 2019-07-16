#!/usr/bin/env bash
set -e

# Set the service name and version.
SERVICE_NAME=$(cat pkg/service/service.go | grep ServiceName | awk -F'"' '{print $2}' | xargs)
SERVICE_VERSION=$(cat VERSION | awk -F'-' '{print $1}' | xargs)
SERVICE_VERSION_PRERELEASE=$(cat VERSION | awk -F'-' '{print $2}' | xargs)

# Set the build date.
if [ -z "$BUILD_DATE" ]; then
    BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
fi

# Get the git commit.
GIT_COMMIT="$(git rev-parse --short HEAD)"
GIT_DIRTY="$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)"
GIT_DESCRIBE="$(git describe --tags --always)"
GIT_IMPORT="$(pwd | sed "s:^$GOPATH/src/::")"

# Set the location of util package.
UTIL_PKG="$(dirname $GIT_IMPORT)/util"

# Get rid of existing binaries.
echo "==> Removing old directory..."
rm -rf dist
mkdir dist

# Determine the arch/os combos we're building for.
OS_PLATFORM_ARG=(darwin)
OS_ARCH_ARG=(amd64)

# Build!
for OS in ${OS_PLATFORM_ARG[@]}; do
  for ARCH in ${OS_ARCH_ARG[@]}; do
    echo "==> Building binary for $OS/$ARCH..."
    GOARCH=$ARCH GOOS=$OS CGO_ENABLED=0 \
    go build \
    -ldflags "-s -w
        -X ${UTIL_PKG}/version.Version=${SERVICE_VERSION}
        -X ${UTIL_PKG}/version.VersionPrerelease=${SERVICE_VERSION_PRERELEASE}
        -X ${UTIL_PKG}/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY}
        -X ${UTIL_PKG}/version.GitDescribe=${GIT_DESCRIBE}
        -X ${UTIL_PKG}/version.BuildDate=${BUILD_DATE}
        -X ${UTIL_PKG}/version.Platform=$OS/$ARCH" \
    -o "dist/bin/$SERVICE_NAME.$OS-$ARCH" .
  done
done

# Packaging.
echo "==> Packaging:"
mkdir dist/v$SERVICE_VERSION
mv dist/bin dist/v$SERVICE_VERSION
cd dist ; tar -zcf v$SERVICE_VERSION.tar.gz v$SERVICE_VERSION

# Done!
echo "==> Results:"
ls -lh .
