#！ /bin/bash

if [[ "$(docker images -q registry.cn-hangzhou.aliyuncs.com/cloud_native_match/scoring:0.1 2> /dev/null)" == "" ]]; then
  docker pull registry.cn-hangzhou.aliyuncs.com/cloud_native_match/scoring:0.1
fi

if [[ "$(docker ps -f "name=scoring" -q 2> /dev/null)" != "" ]]; then
  docker stop $(docker ps -f "name=scoring" -q)
  docker rm $(docker ps -f "name=scoring" -q)
fi

docker run -it --rm --net host -e "SERVER_PORT=8081" -v $(pwd)/datasource/checkSum-small.data:/tmp/checkSum.data --name scoring registry.cn-hangzhou.aliyuncs.com/cloud_native_match/scoring:0.1
