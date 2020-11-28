dev:
	sh ./dockerize/scripts/dev.sh

shutdown:
	sh ./dockerize/scripts/shutdown.sh

build-bin:
	sh ./dockerize/scripts/build-bin.sh

run-prd: build-bin
	sh ./dockerize/scripts/run-prd.sh
