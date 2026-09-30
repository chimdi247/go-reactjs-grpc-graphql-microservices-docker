FROM golang:1.22-alpine AS build
RUN apk add --no-cache gcc g++ make ca-certificates git
WORKDIR /go/src/github.com/akhilsharma90/go-graphql-microservice
COPY go.mod go.sum* ./
COPY pkg pkg
COPY catalog catalog
ENV GOFLAGS=-mod=mod
RUN go mod tidy
RUN go build -o /go/bin/app ./catalog/cmd/catalog

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
WORKDIR /usr/bin
COPY --from=build /go/bin/app .
EXPOSE 8080 8081
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=15s \
  CMD wget -qO- http://localhost:8081/health || exit 1
CMD ["app"]
