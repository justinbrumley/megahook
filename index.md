## Megahook

Test webhooks on localhost

### How it works

Megahook's API receives traffic from your webhook and forwards the request to the client running on your machine. The client will then forward the http request to your local env, get a response back, then send that response back through the API to the original source.

### Dependencies

- [Go v1.14](https://golang.org/doc/install)
- or [Docker](https://docs.docker.com/get-docker/)

### Installing & Running

**Using Go**

```bash
# Installing
go get github.com/justinbrumley/megahook
go install github.com/justinbrumley/megahook

# Running Local Forward
megahook http://localhost:8080/my/favorite/webhook my-little-webhook
```

**Using Docker**

https://hub.docker.com/repository/docker/justinbrumley/megahook/general

```bash
docker run -d \
  -e WEBHOOK_URL=http://localhost:8080/my/favorite/webhook -e WEBHOOK_NAME=my-little-webhook \
  --network host \
  justinbrumley/megahook:latest 
```

### Namespaces

If you would like to use a custom namespace (i.e. subdomain) on megahook, you can register from the CLI using:

```bash
megahook register <subdomain>
```

If successful, this command will return your API token for using the subdomain. Then you can export the api key when using megahook to route webhooks to your custom subdomain:

```bash
MEGAHOOK_API_TOKEN=b85df400e0d5ef78ea0707cc megahook http://localhost:8080/my/favorite/webhook my-little-webhook 
```

(this example will forward traffic through https://test-subdomain-1234.api.megahook.in/m/my-little-webhook)
