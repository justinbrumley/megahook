# Megahook
Megahook is a utility for forwarding webhooks to your local environment.
---
### Install
[Make sure that you have go installed first.](https://golang.org/doc/install) Then run the following:
```bash
go get github.com/justinbrumley/megahook
go install github.com/justinbrumley/megahook
```
And to connect to the server and start passing traffic around:
```
megahook --name my-little-webhook http://localhost:8080
```
You should be given a URL you can start using for your webhooks. If the name you chose
is already taken, you will be given a randomly generated one.
