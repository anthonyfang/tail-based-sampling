FROM golang:alpine as builder-base

RUN echo http://mirror.yandex.ru/mirrors/alpine/v3.12/main > /etc/apk/repositories; \
    echo http://mirror.yandex.ru/mirrors/alpine/v3.12/community >> /etc/apk/repositories

RUN echo "ipv6" >> /etc/modules && \
    apk update && \
    apk add --no-cache bash

FROM builder-base
WORKDIR /tail-based-sampling
RUN go mod init tail-based-sampling && \
    go get github.com/gofiber/fiber/v2

ADD . /tail-based-sampling

RUN go build

FROM builder-base

ENV SERVER_PORT 8000

COPY --from=1 /tail-based-sampling/tail-based-sampling /usr/bin

CMD ["tail-based-sampling"]
