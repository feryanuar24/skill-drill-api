FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /skill-drill-api ./cmd/api

FROM alpine:3.20
RUN adduser -D appuser
RUN mkdir /app && chown appuser:appuser /app
COPY --from=build /skill-drill-api /skill-drill-api
USER appuser
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["/skill-drill-api"]