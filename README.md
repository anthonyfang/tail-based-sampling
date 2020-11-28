# tail-based-sampling
aliyun tianchi project

## Install GO

Download golang from https://golang.org/dl to install for Windows

Mac:
```sh
brew install go

# Export below to the .zshrc/.bashrc profile
echo 'export GOPATH="${HOME}/.go"' >> ~/.zshrc
echo 'export GOROOT="$(brew --prefix golang)/libexec"' >> ~/.zshrc
echo 'export PATH="$PATH:${GOPATH}/bin:${GOROOT}/bin"' >> ~/.zshrc

. ~/.zshrc

go version
```

## Vscode
Reference to page https://medium.com/backend-habit/setting-golang-plugin-on-vscode-for-autocomplete-and-auto-import-30bf5c58138a to install the vscode golang plugins

Install `REST Client` plugin as well

## Run the project
```sh
go run server.go
```

## Run with docker compose
We can't access `locahost:8080` on our local machine because the network of the docker-compose is using `host` mode. We only can access `locahost:8080, localhost:8000, localhost:8001, localhost:8002` within the container.

Development/Debug:
+ Put those 2 track.data files under **datasource**
 folder
+ Run `make dev` commands to start up the servers.
+ change the code in `vscode` , it will sync into the container automatically 

```sh
# start up and logon the container for debug/develop
# if windows, run
# bash ./dockerize/scripts/dev.sh
make dev

# shutdown the containers
# if windows, run
# bash ./dockerize/scripts/shutdown.sh
make shutdown
```
