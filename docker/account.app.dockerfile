FROM golang:1.22-alpine AS build
RUN apk add --no-cache gcc g++ make ca-certificates git
WORKDIR /go/src/github.com/akhilsharma90/go-graphql-microservice
COPY go.mod go.sum* ./
COPY pkg pkg
COPY account account
# go.sum didn't ship with the new dependencies added for this project
# (bcrypt, OpenTelemetry) — go mod tidy resolves and pins them here, at
# real build time, with real network access, rather than a hand-edited
# go.sum that would just be wrong.
ENV GOFLAGS=-mod=mod
RUN go mod tidy
RUN go build -o /go/bin/app ./account/cmd/account

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
WORKDIR /usr/bin
COPY --from=build /go/bin/app .
EXPOSE 8080 8081
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=15s \
  CMD wget -qO- http://localhost:8081/health || exit 1
CMD ["app"]
