FROM golang:1.25-alpine AS build

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /bin/gateway ./cmd

FROM alpine:3.22
RUN adduser -D -H appuser
USER appuser
COPY --from=build /bin/gateway /bin/gateway
EXPOSE 8080
ENTRYPOINT ["/bin/gateway"]
