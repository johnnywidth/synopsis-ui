# Synopsis UI [![CircleCI](https://circleci.com/gh/johnnywidth/synopsis-ui.svg?style=svg)](https://circleci.com/gh/johnnywidth/synopsis-ui)

Synopsis - it ui for [synopsis](https://github.com/johnnywidth/synopsis) tool.
It is work like [satis](https://getcomposer.org/doc/articles/handling-private-packages-with-satis.md).
But more faster, because build time not depend from quantity of repository.

## To run application need install [golang](https://golang.org/doc/install):
    go get github.com/johnnywidth/synopsis-ui
    cd $GOPATH/src/github.com/johnnywidth/synopsis-ui/
    go build
    ./synopsis-ui

## Run by Docker
    docker run
     -p 9091:8080
     --name synopsis-ui
     -v $(pwd)/data:/data
     -v $HOME/.ssh:/root/.ssh
     johnnywidth/synopsis-ui

## Docker ENV Variables
    HOST        ""                  # dy default is empty, you can set `localhost` and use nginx
    FILE        "/data/config.json" # default path for config file with repositories
    THREAD      50                  # default nomber of concurrent processes
    OUTPUT      "/data/output"      # default path for archived packages

## Web access
`http://localhost:9091/` - Info about building packages.
`http://localhost:9091/admin` - Admin panel.

## License
Licensed under the Apache License, Version 2.0. See [LICENSE](https://github.com/johnnywidth/synopsis-ui/blob/master/LICENSE) for the full license text.
