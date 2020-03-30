#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -ux

# Build ctags docker image for universal-ctags-dev
IMAGE=ctags DOCKER_TARGET=ctags DOCKER_BUILD_FLAGS="--quiet" ./cmd/symbols/build.sh

# Build and run symbols binary
./dev/libsqlite3-pcre/build.sh
OUTPUT=./.bin ./cmd/symbols/go-build.sh
LIBSQLITE3_PCRE="$(./dev/libsqlite3-pcre/build.sh libpath)" ./.bin/symbols