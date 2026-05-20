FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/vnodex-node ./services/node/cmd/vnodex-node
RUN go build -o /out/vnodex-cli ./services/node/cmd/vnodex-cli

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /out/vnodex-node /usr/local/bin/vnodex-node
COPY --from=build /out/vnodex-cli /usr/local/bin/vnodex-cli
ENTRYPOINT ["/usr/local/bin/vnodex-node"]
