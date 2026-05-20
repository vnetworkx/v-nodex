FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/vnodex-api ./services/api/cmd/vnodex-api

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /out/vnodex-api /usr/local/bin/vnodex-api
ENTRYPOINT ["/usr/local/bin/vnodex-api"]
