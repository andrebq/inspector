FROM golang:alpine as builder

WORKDIR /app/inspector
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/inspector

FROM alpine
COPY --from=builder /usr/local/bin/inspector /usr/local/bin/inspector
CMD [ "inspector" ]
