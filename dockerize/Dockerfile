FROM golang:alpine as builder-base

# RUN echo http://mirror.yandex.ru/mirrors/alpine/v3.12/main > /etc/apk/repositories; \
#     echo http://mirror.yandex.ru/mirrors/alpine/v3.12/community >> /etc/apk/repositories


ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn

RUN echo http://mirrors.ustc.edu.cn/alpine/v3.12/main > /etc/apk/repositories; \
    echo http://mirrors.ustc.edu.cn/alpine/v3.12/community >> /etc/apk/repositories

RUN apk update && \
    apk add --no-cache bash

FROM builder-base
WORKDIR /tail-based-sampling

RUN cat /etc/resolv.conf

RUN go mod init tail-based-sampling && \
    go get github.com/gofiber/fiber/v2

ADD . /tail-based-sampling

RUN go build

FROM builder-base

ENV SERVER_PORT 8000

COPY --from=1 /tail-based-sampling/tail-based-sampling /usr/bin

CMD ["tail-based-sampling"]
