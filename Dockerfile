FROM golang:1.14.2-alpine3.11 AS builder
WORKDIR $GOPATH/src/github.com/dsociative/stats/
COPY . .
RUN go build -o /usr/bin/local/stats

FROM alpine:3.11
COPY --from=builder /usr/bin/local/stats /usr/bin/local/stats
EXPOSE 8089
ENTRYPOINT ["/usr/bin/local/stats"]