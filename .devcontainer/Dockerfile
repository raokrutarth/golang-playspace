# User's dev env image must derive from devsh/base-dev-env.
# See https://github.com/recode-sh/base-dev-env source.
# All development environments will be Ubuntu-based

# simplifies basic bash tools and python version
FROM mcr.microsoft.com/vscode/devcontainers/go:1.19-buster
ARG DEBIAN_FRONTEND=noninteractive

# RUN will use bash
SHELL ["/bin/bash", "-c"]

# Install system dependencies
RUN set -euo pipefail \
    && apt-get --assume-yes --quiet --quiet update \
    && apt-get --assume-yes --quiet --quiet install \
    apt-transport-https \
    apt-utils \
    build-essential \
    ca-certificates \
    curl \
    git \
    gnupg \
    locales \
    lsb-release \
    man-db \
    nano \
    sudo \
    tzdata \
    unzip \
    vim \
    wget \
    libpq-dev \
    jq \
    postgresql-client \
    software-properties-common \
    && rm --recursive --force /var/lib/apt/lists/*

WORKDIR /workspaces/golang-playspace
USER vscode

ENV USER=vscode
ENV HOME=/home/${USER}
ENV EDITOR=/usr/bin/vim
ENV TZ=America/Los_Angeles

# Install VSCode Go extension's dependencies
RUN set -euo pipefail \
    && /usr/local/go/bin/go install github.com/ramya-rao-a/go-outline@latest \
    && /usr/local/go/bin/go install github.com/cweill/gotests/gotests@latest \
    && /usr/local/go/bin/go install github.com/fatih/gomodifytags@latest \
    && /usr/local/go/bin/go install github.com/josharian/impl@latest \
    && /usr/local/go/bin/go install github.com/haya14busa/goplay/cmd/goplay@latest \
    && /usr/local/go/bin/go install github.com/go-delve/delve/cmd/dlv@latest \
    && /usr/local/go/bin/go install honnef.co/go/tools/cmd/staticcheck@latest \
    && /usr/local/go/bin/go install golang.org/x/tools/gopls@latest \
    && /usr/local/go/bin/go install github.com/create-go-app/cli/v3/cmd/cgapp@latest \
    && /usr/local/go/bin/go install github.com/go-critic/go-critic/cmd/gocritic@latest \
    && /usr/local/go/bin/go install github.com/securego/gosec/v2/cmd/gosec@latest \
    && /usr/local/go/bin/go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && /usr/local/go/bin/go install github.com/swaggo/swag/cmd/swag@latest \
    && /usr/local/go/bin/go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
    && /usr/local/go/bin/go install github.com/cosmtrek/air@latest

ENTRYPOINT ["/bin/bash"]



