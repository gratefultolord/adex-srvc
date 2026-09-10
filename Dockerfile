FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /adex ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o /mockdsp ./cmd/mockdsp


FROM alpine:3.22 AS app

WORKDIR /app

COPY --from=builder /adex /app/adex

EXPOSE 8080

CMD ["/app/adex"]


FROM alpine:3.22 AS mockdsp

WORKDIR /app

COPY --from=builder /mockdsp /app/mockdsp

ENTRYPOINT ["/app/mockdsp"]