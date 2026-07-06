FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY utils/ /app/utils/
COPY services/notifications/ /app/services/notifications/
WORKDIR /app/services/notifications
RUN go mod download && CGO_ENABLED=0 go build -o /notifications .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget \
    && addgroup -g 1001 -S osship \
    && adduser -u 1001 -S osship -G osship
COPY --from=builder /notifications /notifications
USER 1001
EXPOSE 8086
CMD ["/notifications"]
