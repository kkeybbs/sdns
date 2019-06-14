# Simple dns server

A simple dns server with exact, suffix and regex rules.

The server parses matched domains to any ip addresses or the server running `sdns` with rule value `self`, and forwards unmatched ones to other dns servers.

Rules will be checked with order: exact => suffix => regex.

## Install
```
go get -u github.com/kkeybbs/sdns
```

If you want to build for linux on windows.
```
set GOOS=linux
set GOARCH=amd64
set GOPATH=%cd%\gopath
go get -v github.com/kkeybbs/sdns
go build -ldflags="-s -w" github.com/kkeybbs/sdns
```

## Install as linux systemd service
```
# put compiled sdns to /data/sdns, or edit the sdns.service to use another path.
mkdir -p /data/sdns
cp sdns /data/sdns

cp sdns.service /usr/lib/systemd/system/sdns.service
systemctl daemon-reload
systemctl enable --now sdns
```

## Config
You can find a demo here: [sdns.yaml](sdns.yaml)

## Run
```
# start a dns server, both udp and tcp port will be listened.
./sdns -c sdns.yaml

# test rules in sdns.yaml
./sdns -t github.com
```

