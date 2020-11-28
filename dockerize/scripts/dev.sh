#！ /bin/bash
docker-compose -f ./dockerize/docker-compose.yml up --force-recreate -d 
docker exec -it tail-base-sampling-backend-1 bash
