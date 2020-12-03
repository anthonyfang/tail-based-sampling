#！ /bin/bash
container_name=$NAME
docker attach $(docker ps -f "name=$container_name" -q)
