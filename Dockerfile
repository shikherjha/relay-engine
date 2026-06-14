FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/relay-engine .

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/relay-engine /relay-engine
EXPOSE 8002
ENTRYPOINT ["/relay-engine"]
