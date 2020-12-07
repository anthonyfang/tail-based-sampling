dev:
	sh ./dockerize/scripts/dev.sh

dev-4g:
	sh ./dockerize/scripts/dev-4g.sh

shutdown:
	sh ./dockerize/scripts/shutdown.sh

build-bin:
	sh ./dockerize/scripts/build-bin.sh

run-prd: build-bin
	sh ./dockerize/scripts/run-prd.sh

run-prd-4g: build-bin
	sh ./dockerize/scripts/run-prd-4g.sh

start-up-datasource:
	sh ./dockerize/scripts/start-up-datasource.sh

start-scoring:
	sh ./dockerize/scripts/start-scoring.sh

start-scoring-4g:
	sh ./dockerize/scripts/start-scoring-4g.sh

attach:
	sh ./dockerize/scripts/attach.sh
