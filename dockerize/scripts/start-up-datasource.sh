#！ /bin/bash

if [[ "$(docker images -q tail-based-sampling:data-source 2> /dev/null)" == "" ]]; then
  docker build -f dockerize/Dockerfile.datasource -t tail-based-sampling:data-source .
fi

docker run -d \
    -p 8080:8080 \
    --name tail-based-sampling-data-source \
    -v $(pwd)/datasource:/app/data \
    tail-based-sampling:data-source /app/node_modules/http-server/bin/http-server /app/data
