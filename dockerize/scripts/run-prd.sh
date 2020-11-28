#！ /bin/bash
version=${VERSION:=latest}
docker run --rm -it  --net host -e "SERVER_PORT=8000" --name "clientprocess1" -d tail-base-sampling:$version
docker run --rm -it  --net host -e "SERVER_PORT=8001" --name "clientprocess2" -d tail-base-sampling:$version
docker run --rm -it  --net host -e "SERVER_PORT=8002" --name "backendprocess" -d tail-base-sampling:$version
