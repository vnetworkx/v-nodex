FROM golang:1.22-alpine AS build

WORKDIR /src

COPY . .

WORKDIR /src/services/node
RUN go mod download
RUN go build -o /out/vnodex-node ./cmd/vnodex-node
RUN go build -o /out/vnodex-cli ./cmd/vnodex-cli

FROM alpine:3.20

WORKDIR /app

COPY --from=build /out/vnodex-node /app/vnodex-node
COPY --from=build /out/vnodex-cli /app/vnodex-cli

EXPOSE 8080

ENTRYPOINT ["/app/vnodex-node"]