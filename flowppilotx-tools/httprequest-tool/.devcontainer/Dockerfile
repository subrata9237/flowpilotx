FROM mcr.microsoft.com/devcontainers/go:1.24.2

# Install additional OS packages
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    curl \
    git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set up Git configuration directly
RUN git config --global url."https://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@github.com/".insteadOf "https://github.com/" && \
    git config --global core.editor "vim" && \
    git config --global credential.helper store && \
    git config --global credential.helper 'cache --timeout=3600' && \
    git config --global url."https://".insteadOf git://

# Create common Go directories
RUN mkdir -p /go/src /go/bin /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

# Set the working directory
WORKDIR /workspace 