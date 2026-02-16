FROM golang:1.26-alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY main.go .
COPY pkg pkg/

RUN go vet ./...
RUN go run golang.org/x/vuln/cmd/govulncheck@latest ./...
RUN go test ./...

RUN go build -v -o /usr/src/app/lonn

FROM gcr.io/distroless/static-debian12

COPY --from=builder /usr/src/app/lonn /lonn

CMD ["/lonn"]
