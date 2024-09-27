# build stage
FROM golang:1.23.1-alpine as builder

WORKDIR /builder

ENV GO111MODULE=on
ADD go.mod /builder/go.mod
ADD go.sum /builder/go.sum
RUN go mod download

ADD . /builder
RUN go build -o app .

# final stage
FROM alpine:3.19

COPY --from=builder /builder/app /app

EXPOSE 8000

ENTRYPOINT [ "/app" ]