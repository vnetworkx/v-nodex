FROM golang:1.22-alpine AS build

WORKDIR /src

COPY . .

WORKDIR /src/services/api
RUN go mod download
RUN go build -o /out/vnodex-api ./cmd/vnodex-api

FROM alpine:3.20

WORKDIR /app

COPY --from=build /out/vnodex-api /app/vnodex-api

EXPOSE 8090

ENTRYPOINT ["/app/vnodex-api"]