# --- build stage ---
FROM 10.29.230.150:31381/library/golang:1.19-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/sms-webhook

# --- run stage ---
FROM 10.29.230.150:31381/library/golang:1.19-alpine
WORKDIR /root/
COPY --from=builder /app/sms-webhook .
EXPOSE 8080
CMD ["./sms-webhook"]
