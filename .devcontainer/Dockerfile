# Base image
FROM ubuntu:22.04

# Set environment variables
ENV GO_VERSION=1.24.2 \
    GO_OS=linux \
    GO_ARCH=arm64 \
    GOROOT=/usr/local/go \
    GOPATH=/go \
    PATH=/usr/local/go/bin:/go/bin:$PATH \
    DEBIAN_FRONTEND=noninteractive \
    # Go environment variables
    GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=arm64 \
    # Go proxy settings
    GOPROXY=https://proxy.golang.org,direct \
    GOSUMDB=sum.golang.org \
    GOPRIVATE="" \
    # Go development settings
    GOBIN=/go/bin \
    GOTOOLDIR=/usr/local/go/pkg/tool/linux_arm64

# Install required packages
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    git \
    build-essential \
    ca-certificates \
    procps \
    lsb-release \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Download and install Go manually
RUN wget https://go.dev/dl/go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz && \
    rm go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz && \
    chown -R root:root /usr/local/go

# Create Go directories and set permissions
RUN mkdir -p /go/src /go/bin /go/pkg && \
    chmod -R 777 /go && \
    chown -R root:root /go

# Install Go tools
RUN go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install golang.org/x/tools/gopls@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Set up Git configuration directly
RUN git config --global url."https://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@github.com/".insteadOf "https://github.com/" && \
    git config --global core.editor "vim" && \
    git config --global credential.helper store && \
    git config --global credential.helper 'cache --timeout=3600' && \
    git config --global url."https://".insteadOf git://
# Create common Go directories for projects
RUN mkdir -p /go/src/github.com && \
    chmod -R 777 /go/src && \
    chown -R root:root /go/src

# Set working directory
WORKDIR /workspace

# Add Go binary directory to PATH
RUN echo "export PATH=$PATH:/usr/local/go/bin:/go/bin" >> /root/.bashrc

# Set up Go environment in bashrc
RUN echo "export GO111MODULE=on" >> /root/.bashrc && \
    echo "export GOPATH=/go" >> /root/.bashrc && \
    echo "export GOBIN=/go/bin" >> /root/.bashrc && \
    echo "export GOROOT=/usr/local/go" >> /root/.bashrc

# Add Git credential setup to bashrc
RUN echo 'if [ -n "${GIT_USERNAME}" ] && [ -n "${GIT_TOKEN}" ]; then' >> /root/.bashrc && \
    echo '    envsubst < /root/.git-credentials.template > /root/.git-credentials' >> /root/.bashrc && \
    echo '    chmod 600 /root/.git-credentials' >> /root/.bashrc && \
    echo 'fi' >> /root/.bashrc

# Copy go.mod and go.sum files
COPY go.mod ./

# Download dependencies and run go mod tidy
RUN go mod download && \
    go mod tidy

# Expose common ports
EXPOSE 8080 8000 2345

# Expose the port the app runs on
EXPOSE 8080
