FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/employee-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates && adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=builder /out/employee-service ./employee-service
USER app
EXPOSE 8080
ENTRYPOINT ["./employee-service"]
