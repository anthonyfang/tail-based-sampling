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
