// dockerfileTemplate.go
package main

import (
	"errors"
	"strings"
)

func dockerfileTemplate(info RuntimeInfo) (string, error) {
	switch info.Runtime {

	// ==================== Go ====================
	case "go":
		version := "1.23-alpine"
		if info.Version != "default" {
			version = info.Version + "-alpine"
		}
		return `
# Multi-stage build for Go applications
FROM golang:` + version + ` AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /main .

# Final stage: minimal scratch image
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /main /main

EXPOSE ${PORT:-8080}

USER 65534:65534

ENTRYPOINT ["/main"]
`, nil

	// ==================== Node.js & Frameworks ====================
	case "node", "react", "nextjs", "vue", "svelte", "angular", "typescript":
		baseImage := "node:20-alpine" // modern, secure default
		if info.Version != "default" {
			// Use specified version, fallback to alpine for size
			if strings.Contains(info.Version, "alpine") {
				baseImage = "node:" + info.Version
			} else {
				baseImage = "node:" + info.Version + "-alpine"
			}
		}

		// Special handling for Next.js (needs standalone output)
		if info.Runtime == "nextjs" {
			return `

FROM ` + baseImage + `

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .

RUN npm run build

ENV NODE_ENV=production
ENV PORT=${PORT:-3000}
EXPOSE ${PORT:-3000}

CMD ["npx", "next", "start", "-p", "$PORT"]

`, nil
		}

		// Generic Node.js / React / Vue / etc.
		return `
FROM ` + baseImage + `

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

ENV PORT=${PORT:-3000}

EXPOSE ${PORT:-3000}

CMD ["npm", "start"]
`, nil

	// ==================== Python ====================
	case "python":
		version := "3.11-slim"
		if info.Version != "default" {
			version = info.Version + "-slim"
		}
		return `
FROM python:` + version + `

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV PORT=${PORT:-8000}

EXPOSE ${PORT:-8000}

CMD ["python", "main.py"]  # Common entrypoint; adjust if app.py or server.py
`, nil

	// ==================== Static Sites ====================
	case "static":
		return `
FROM nginx:alpine

COPY . /usr/share/nginx/html

EXPOSE ${PORT:-80}

CMD ["nginx", "-g", "daemon off;"]
`, nil

	// ==================== Rust ====================
	case "rust":
		return `
# Multi-stage build for Rust
FROM rust:1.75 AS builder

WORKDIR /app

COPY Cargo.toml Cargo.lock ./
RUN cargo fetch

COPY . .

RUN cargo build --release

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/app /usr/local/bin/app  # Assumes binary name = project name

EXPOSE ${PORT:-8080}

CMD ["app"]
`, nil

	// ==================== PHP ====================
	case "php":
		return `
FROM php:8.3-fpm-alpine

WORKDIR /var/www/html

COPY . .

RUN docker-php-ext-install mysqli pdo pdo_mysql  # Common extensions

EXPOSE ${PORT:-9000}

CMD ["php-fpm"]
`, nil

	// ==================== Java (Maven or Gradle) ====================
	case "java":
		return `
# Multi-stage Java build
FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /app

COPY pom.xml ./
RUN mvn dependency:go-offline

COPY src ./src
RUN mvn package -DskipTests

FROM eclipse-temurin:21-jre-alpine

WORKDIR /app

COPY --from=builder /app/target/*.jar app.jar

EXPOSE ${PORT:-8080}

CMD ["java", "-jar", "app.jar"]
`, nil

	// ==================== .NET ====================
	case "dotnet":
		return `
# Multi-stage .NET build
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS builder

WORKDIR /app

COPY *.csproj .
RUN dotnet restore

COPY . .
RUN dotnet publish -c Release -o out --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:8.0

WORKDIR /app
COPY --from=builder /app/out .

EXPOSE ${PORT:-8080}

ENTRYPOINT ["dotnet", "your-app.dll"]  # Replace with actual DLL name
`, nil

	// ==================== Dart / Flutter (Web) ====================
	case "dart/flutter":
		return `
FROM cirrusci/flutter:stable AS builder

WORKDIR /app

COPY . .

RUN flutter build web --release

FROM nginx:alpine

COPY --from=builder /app/build/web /usr/share/nginx/html

EXPOSE ${PORT:-80}
`, nil

	// ==================== Deno / Bun ====================
	case "deno_bun":
		return `
FROM denoland/deno:alpine-1.40

WORKDIR /app

COPY . .

EXPOSE ${PORT:-8000}

CMD ["deno", "run", "--allow-net", "--allow-env", "main.ts"]
`, nil

	// ==================== Default Fallback ====================
	default:
		return "", errors.New("no Dockerfile template for runtime: " + info.Runtime)
	}
}
