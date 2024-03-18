FROM golang:alpine
RUN apk add --no-cache upx
WORKDIR /src
COPY src .
RUN go mod tidy

# ### Local build and run
# RUN go build -ldflags="-s -w" -o build/webrun
# ENTRYPOINT [ "build/webrun" ]]

### Build for known architectures:
RUN mkdir build
RUN GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o webrun.exe && upx webrun.exe && tar -czf build/webrun-windows-x86_64.tgz webrun.exe
RUN GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o webrun && upx webrun --force-macos && tar -czf build/webrun-darwin-x86_64.tgz webrun
RUN GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o webrun && upx webrun --force-macos && tar -czf build/webrun-darwin-aarch64.tgz webrun
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o webrun && upx webrun && tar -czf build/webrun-linux-x86_64.tgz webrun
RUN GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o webrun && upx webrun && tar -czf build/webrun-linux-aarch64.tgz webrun
RUN GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o webrun && upx webrun && tar -czf build/webrun-linux-armv6l.tgz webrun

ENTRYPOINT [ "/bin/sh", "-c", "cp -R build/* /build"]