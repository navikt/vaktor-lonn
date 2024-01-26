FROM golang:1.21-alpine as builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY main.go .
COPY pkg pkg/

RUN go build -v -o /usr/src/app/lonn

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=builder /usr/src/app/lonn /app/lonn

CMD ["/app/lonn"]
