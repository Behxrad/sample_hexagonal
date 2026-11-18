FROM golang:1.25-alpine AS builder

RUN go install github.com/fullstorydev/grpcui/cmd/grpcui@latest

FROM alpine:3.22

COPY --from=builder /go/bin/grpcui /usr/local/bin/grpcui

EXPOSE 8080

ENTRYPOINT ["grpcui"]