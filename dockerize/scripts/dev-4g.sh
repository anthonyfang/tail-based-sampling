#！ /bin/bash
docker-compose -f ./dockerize/docker-compose-4g.yml up --force-recreate -d 
docker exec -it backendprocess bash