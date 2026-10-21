FROM golang:1.27-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/employee-service .

FROM alpine:3.22
RUN apk --no-cache add ca-certificates && adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=builder /out/employee-service ./employee-service
USER app
EXPOSE 8080 9090
ENTRYPOINT ["./employee-service"]
