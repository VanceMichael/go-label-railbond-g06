FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
RUN GOTOOLCHAIN=local CGO_ENABLED=0 go build -trimpath -o /out/railbond ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /out/railbond /app/railbond
COPY migrations /app/migrations
ENV PORT=8080
ENV DATABASE_URL=/data/railbond.db
RUN mkdir -p /data
EXPOSE 8080
ENTRYPOINT ["/app/railbond"]

