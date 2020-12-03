dev:
	sh ./dockerize/scripts/dev.sh

shutdown:
	sh ./dockerize/scripts/shutdown.sh

build-bin:
	sh ./dockerize/scripts/build-bin.sh

run-prd: build-bin
	sh ./dockerize/scripts/run-prd.sh

start-up-datasource:
	sh ./dockerize/scripts/start-up-datasource.sh

start-scoring:
	sh ./dockerize/scripts/start-scoring.sh

attach:
	sh ./dockerize/scripts/attach.sh $1
