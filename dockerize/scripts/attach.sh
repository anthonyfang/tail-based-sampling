#！ /bin/bash
container_name=$1
docker attach $(docker ps -f "name=$container_name" -q)
