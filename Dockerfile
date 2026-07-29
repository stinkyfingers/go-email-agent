# Builds both entrypoints into one image; ECS task defs select the binary to
# run via the container's `command`, so there's a single image/ECR repo to
# build, push, and version instead of two.
FROM golang:1.25-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gmail_login ./cmd/gmail_login
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/check_emails ./cmd/check_emails

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gmail_login /usr/local/bin/gmail_login
COPY --from=build /out/check_emails /usr/local/bin/check_emails

# No default ENTRYPOINT/CMD — each ECS task definition sets `command`
# explicitly to pick which binary runs.
