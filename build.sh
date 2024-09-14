#!/bin/bash

APP_NAME="Bolt"
VERSION="0.1.0"

PLATFORMS=(
    "darwin/amd64" 
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

mkdir -p builds

for PLATFORM in "${PLATFORMS[@]}"
do
    IFS='/' read -r -a array <<< "$PLATFORM"
    GOOS=${array[0]}
    GOARCH=${array[1]}

    OUTPUT="builds/${APP_NAME}_${VERSION}_${GOOS}_${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi

    echo "Building new Bolt release for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT"
    if [ $? -ne 0 ]; then
        echo 'An error occurred during build phase! Aborting build process....'
        exit 1
    fi

    # Creates the SHA256 checksum
    if [ "$GOOS" = "windows" ]; then
        sha256sum "$OUTPUT" > "${OUTPUT}.sha256"
    else
        shasum -a 256 "$OUTPUT" > "${OUTPUT}.sha256"
    fi
done

echo "Successfully completed build phase, check build directory for binaries"