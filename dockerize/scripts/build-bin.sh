#！ /bin/bash
version=${VERSION:=latest}
docker build -f dockerize/Dockerfile -t tail-base-sampling:$version .
# docker rmi -f $(docker images --filter "dangling=true" -q)
