FROM golang:alpine as builder-base

RUN echo http://mirror.yandex.ru/mirrors/alpine/v3.12/main > /etc/apk/repositories; \
    echo http://mirror.yandex.ru/mirrors/alpine/v3.12/community >> /etc/apk/repositories

RUN echo http://mirrors.ustc.edu.cn/alpine/v3.12/main > /etc/apk/repositories; \
    echo http://mirrors.ustc.edu.cn/alpine/v3.12/community >> /etc/apk/repositories

RUN apk update && \
    apk add --no-cache bash

FROM builder-base
WORKDIR /tail-based-sampling
RUN go mod init tail-based-sampling && \
    go get github.com/gofiber/fiber/v2 && \
    go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct

ADD . /tail-based-sampling

RUN go build

FROM builder-base

ENV SERVER_PORT 8000

COPY --from=1 /tail-based-sampling/tail-based-sampling /usr/bin

CMD ["tail-based-sampling"]
