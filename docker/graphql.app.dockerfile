FROM golang:1.23-alpine AS build
RUN apk add --no-cache gcc g++ make ca-certificates git
WORKDIR /go/src/github.com/akhilsharma90/go-graphql-microservice
COPY go.mod go.sum* ./
COPY pkg pkg
COPY account account
COPY catalog catalog
COPY order order
COPY graphql graphql
ENV GOFLAGS=-mod=mod

RUN go get google.golang.org/genproto@v0.0.0-20241007155032-5fefd90f89a9 && go mod tidy

# Regenerates graphql/generated.go + models_gen.go from schema.graphql via
# the go:generate directive in graphql/main.go — this is the project's
# own documented codegen mechanism (see README), just run at image build
# time instead of being committed to source, so it's always in sync with
# schema.graphql (including the login/me/totalAccounts additions).



# models.go (mapped in gqlgen.yml) needs the generated Order type, and the other
# hand-written files need generated.go. Break the cycle in two passes:
#  1) hide everything except models.go, add a stub Order, generate (models_gen.go)
#  2) drop the stub, generate again, then restore the hand-written files
RUN cd graphql && \
    mkdir /tmp/hand && \
    find . -maxdepth 1 -name '*.go' ! -name models.go -exec mv {} /tmp/hand/ \; && \
    printf 'package main\n\ntype Order struct{}\n' > stub.go && \
    (go run github.com/99designs/gqlgen || true) && \
    rm -f stub.go generated.go && \
    go run github.com/99designs/gqlgen && \
    mv /tmp/hand/*.go .


RUN go build -o /go/bin/app ./graphql



FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
WORKDIR /usr/bin
COPY --from=build /go/bin/app .
EXPOSE 8080 8081
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=15s \
  CMD wget -qO- http://localhost:8080/health || exit 1
CMD ["app"]
