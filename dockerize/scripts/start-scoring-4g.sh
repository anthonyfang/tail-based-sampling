#！ /bin/bash

if [[ "$(docker images -q java:8 2> /dev/null)" == "" ]]; then
  docker pull java:8
fi

if [[ "$(docker ps -f "name=scoring" -q 2> /dev/null)" != "" ]]; then
  docker stop $(docker ps -f "name=scoring" -q)
  docker rm $(docker ps -f "name=scoring" -q)
fi

docker run --rm -it --net host -e "SERVER_PORT=8081" -v $(pwd)/../:/scoring/ --name scoring java:8 \
  java -Dserver.port=9000 -DcheckSumPath=/scoring/tmp/checkSum.data -jar /scoring/scoring-1.0-SNAPSHOT.jar 
