#！ /bin/bash
docker-compose -f ./dockerize/docker-compose.yml up --force-recreate -d 
docker exec -it backendprocess bash
