FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /skill-drill-api ./cmd/api

FROM alpine:3.20
RUN adduser -D appuser
COPY --from=build /skill-drill-api /skill-drill-api
USER appuser
EXPOSE 8080
ENTRYPOINT ["/skill-drill-api"]