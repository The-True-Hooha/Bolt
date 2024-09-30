#!/bin/bash

   VERSION="0.1.0"
   BINARY_NAME="bolt"

   # Build for macOS
   GOOS=darwin GOARCH=amd64 go build -o ${BINARY_NAME}_${VERSION}_darwin_amd64
   GOOS=darwin GOARCH=arm64 go build -o ${BINARY_NAME}_${VERSION}_darwin_arm64

   # Build for Linux
   GOOS=linux GOARCH=amd64 go build -o ${BINARY_NAME}_${VERSION}_linux_amd64
   GOOS=linux GOARCH=arm64 go build -o ${BINARY_NAME}_${VERSION}_linux_arm64

   # Build for Windows
   GOOS=windows GOARCH=amd64 go build -o ${BINARY_NAME}_${VERSION}_windows_amd64.exe

   # Create archives
   tar -czvf ${BINARY_NAME}_${VERSION}_darwin_amd64.tar.gz ${BINARY_NAME}_${VERSION}_darwin_amd64
   tar -czvf ${BINARY_NAME}_${VERSION}_darwin_arm64.tar.gz ${BINARY_NAME}_${VERSION}_darwin_arm64
   tar -czvf ${BINARY_NAME}_${VERSION}_linux_amd64.tar.gz ${BINARY_NAME}_${VERSION}_linux_amd64
   tar -czvf ${BINARY_NAME}_${VERSION}_linux_arm64.tar.gz ${BINARY_NAME}_${VERSION}_linux_arm64
   zip ${BINARY_NAME}_${VERSION}_windows_amd64.zip ${BINARY_NAME}_${VERSION}_windows_amd64.exe

   echo "Build complete!"