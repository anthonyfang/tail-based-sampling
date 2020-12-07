#！ /bin/bash
version=${VERSION:=latest}

docker stop $(docker ps -q)

docker run --rm -it --net host -e "SERVER_PORT=8080" --name "datasource" -d -v $(pwd)/datasource:/app/data \
    tail-based-sampling:data-source /app/node_modules/http-server/bin/http-server /app/data dockerize_datasource:$version 
docker run --rm -it  --net host -e "SERVER_PORT=8000" --name "clientprocess1" -d tail-base-sampling:$version
docker run --rm -it  --net host -e "SERVER_PORT=8001" --name "clientprocess2" -d tail-base-sampling:$version
docker run --rm -it  --net host -e "SERVER_PORT=8002" --name "backendprocess" -d tail-base-sampling:$version
