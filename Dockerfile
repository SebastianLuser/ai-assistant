FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o asistente ./cmd
FROM alpine:3.20
RUN apk add --no-cache ca-certificates sqlite
WORKDIR /app
COPY --from=builder /app/asistente .
COPY --from=builder /app/skills ./skills
EXPOSE 8080
CMD ["./asistente"]
