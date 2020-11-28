FROM golang:alpine as builder-base

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
