FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /polybot ./cmd/polybot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /polybot /polybot
EXPOSE 8086
ENTRYPOINT ["/polybot"]
