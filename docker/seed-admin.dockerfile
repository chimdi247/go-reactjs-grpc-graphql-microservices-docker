FROM golang:1.23-alpine AS build
RUN apk add --no-cache gcc g++ make ca-certificates git
WORKDIR /go/src/github.com/akhilsharma90/go-graphql-microservice
COPY go.mod go.sum* ./
COPY account account
ENV GOFLAGS=-mod=mod

RUN go get google.golang.org/genproto@v0.0.0-20241007155032-5fefd90f89a9 && go mod tidy
RUN go build -o /go/bin/seed-admin ./account/cmd/seed-admin

FROM alpine:3.20
WORKDIR /usr/bin
COPY --from=build /go/bin/seed-admin .
CMD ["seed-admin"]
