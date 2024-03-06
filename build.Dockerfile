FROM golang:alpine
WORKDIR /src
COPY src .
RUN go mod tidy

### Build for current architecture
# RUN go build -ldflags="-s -w" -o build/webrun
# RUN mv build/webrun build/webrun-$(uname -s)-$(uname -m)

### Build for known architectures:
RUN GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/webrun-windows-amd64.exe
# RUN GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o build/webrun-windows-arm64.exe
RUN GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/webrun-darwin-amd64
RUN GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/webrun-darwin-arm64
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/webrun-linux-amd64
RUN GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o build/webrun-linux-arm64
# RUN GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o build/webrun-linux-armv7

### Enable Compression (takes time)
RUN apk add --no-cache upx && upx build/webrun-* --force-macos

ENTRYPOINT [ "/bin/sh", "-c", "cp -R build/* /build"]