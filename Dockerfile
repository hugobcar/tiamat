FROM golang:1.12.8 as builder
WORKDIR /go/src/github.com/hugobcar/tiamat
ADD . /go/src/github.com/hugobcar/tiamat
# RUN GO111MODULE=on go mod vendor
RUN CGO_ENABLED=0 go build -o tiamat

FROM alpine:3.10.1
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /go/src/github.com/hugobcar/tiamat/tiamat .
CMD ["./tiamat"]
