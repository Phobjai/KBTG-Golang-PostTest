FROM golang:1.21.6-alpine3.19 as build-base

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -v

RUN go build -o ./out/go-sample .

# ====================


FROM alpine:3.19.1
COPY --from=build-base /app/out/go-sample /app/go-sample

CMD ["/app/go-sample"]
