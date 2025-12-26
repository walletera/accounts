FROM golang:1.24-alpine3.20 AS builder

WORKDIR /root

COPY . ./

RUN go build -o bin/accounts cmd/main.go

FROM alpine:3.20

# Allow customization of user ID and group ID (it's useful when you use Docker bind mounts)
ARG UID=1000
ARG GID=1000

RUN addgroup -g ${GID} -S app && adduser -u ${UID} -S -G app app

WORKDIR /home/app

COPY --from=builder /root/bin/accounts ./


RUN chown app:app ./accounts
RUN chmod +x ./accounts

USER app

CMD ["./accounts"]
