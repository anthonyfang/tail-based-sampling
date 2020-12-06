#！ /bin/bash

docker tag tail-base-sampling:latest registry.cn-hangzhou.aliyuncs.com/tail-base-sampling/tail-test:0.1

docker push registry.cn-hangzhou.aliyuncs.com/tail-base-sampling/tail-test:0.1
