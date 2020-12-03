FROM golang:1.15 as build

WORKDIR /go/src/verify
COPY verify verify
COPY notes notes
WORKDIR /go/src/verify/verify

ENV CGO_ENABLED=0
RUN go build -o /go/bin/verifypr ./

FROM gcr.io/distroless/static-debian10

COPY --from=build /go/bin/verifypr /verifypr

ENTRYPOINT ["/verifypr"]
