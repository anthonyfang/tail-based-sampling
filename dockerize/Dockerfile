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

RUN go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct

RUN go mod init tail-based-sampling; \
    go get github.com/gin-gonic/gin; \
    go get github.com/golang/protobuf; \
    go get github.com/gorilla/websocket; \
    go get github.com/soheilhy/cmux; \
    go get golang.org/x/net; \
    go get google.golang.org/grpc; \
    go get gopkg.in/yaml.v2; \
    go get google.golang.org/genproto; \
    go get github.com/gorilla/websocket; \
    go get github.com/gin-contrib/sse; \
    go get golang.org/x/text; \
    go get github.com/mattn/go-isatty; \
    go get golang.org/x/sys; \
    go get github.com/go-playground/validator/v10; \
    go get github.com/ugorji/go; \
    go get github.com/ugorji/go/codec; \
    go get github.com/go-playground/universal-translator; \
    go get github.com/leodido/go-urn; \
    go get github.com/go-playground/locales
# go get github.com/gofiber/fiber/v2 && \

ADD . /tail-based-sampling

RUN go build

FROM builder-base

ENV SERVER_PORT 8000

COPY --from=1 /tail-based-sampling/tail-based-sampling /usr/bin

CMD ["tail-based-sampling"]
