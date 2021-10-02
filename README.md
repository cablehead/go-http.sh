# http.sh

*Integrate a HTTP server into your shell pipelines*

Not quite sure where this is going to go, but... let's see!

This is a small CLI which writes HTTP requests to its STDOUT. It reads
potential responses for these requests from its STDIN. That part is counter
intuitive, but it works out well.

The expected usage is to pipe a subscription from a message queue into STDIN
and to pipe STDOUT to a message queue writer.

### Example usage

You can simulate a subscription from a message queue with `tail -F responses` and
writing to a message queue with `tee requests`.

Run the HTTP server in window 1:

```
$ cp /dev/null responses && tail -F responses | go run . | tee requests | jq
```

Create a request in window 2:

```
$ curl localhost:8080
```

Note the curl is hanging, waiting for a response. You should see the request
in window 1's STDOUT.

```json
{
  "app": "http.request",
  "content": {
    "method": "GET",
    "header": {
      "Accept": [
        "*/*"
      ],
      "User-Agent": [
        "curl/7.64.1"
      ]
    },
    "remote_addr": "[::1]:50503",
    "uri": "/",
    "body": "",
    "request_id": "8684598f-2181-41fd-ba56-7c4b863e4fc8"
  }
}
```

We can use window 3 to create a response. So far, a response's schema is:

- **request_id**: the request_id this is a response for
- **body**: base64 encoded response body

Grab [jo](https://github.com/jpmens/jo) if you don't have it, it's great. We
can use it to build a response and then write that to `responses` which is
being piped to window 1's STDIN.

```
$ jo request_id=8684598f-2181-41fd-ba56-7c4b863e4fc8 body="$(echo oh hai | base64)" | \
    tee -a responses
```

```json
{
  "request_id": "8684598f-2181-41fd-ba56-7c4b863e4fc8",
  "body": "b2ggaGFpCg=="
}
```

You should see the curl in window 2 return with `oh hai`. The response is logged by the HTTP server in window 1:

```json
{
  "app": "http.response.log",
  "content": {
    "took": 32263.4,
    "response": {
      "body": "b2ggaGFpCg==",
      "request_id": "8684598f-2181-41fd-ba56-7c4b863e4fc8",
    }
  }
}
```

Note how long the request `took`. That time is in milliseconds, so in this
case, 32s. I need to get faster with `jo`!  We can automate reading requests
and writing responses.

Run this snippet in window 3. It tails `requests`, using `jq` to select only
`http.request` messages, as the server also emits `http.repsonse.log` messages.
The actual request is under the `content` key.  `jq` is used to generate the
response schema, relaying the `request_id` and generating a base64 encoded
`body` based on the `uri` of the request.

```
# unbuffer jq by default. Life works out better that way.
$ alias jq='jq --unbuffered'
$ tail -F -n1 requests | jq -c '
    select(.app == "http.request") |
    .content |
    {request_id, "body": ("oh hai. your path is: \(.uri)\n" | @base64)}
' | tee -a responses
```

In window 2, `curl` the server again. It should return immediately now.

```
$ curl localhost:8080
oh hai. your path is: /
```

In window 3 you should see something like this. `took` is 0.4ms.

```json
{
  "app": "http.response.log",
  "content": {
    "took": 0.4,
    "response": {
      "body": "b2ggaGFpLiB5b3VyIHBhdGggaXM6IC8K",
      "request_id": "53c59d7b-6533-4059-acbb-32c7e3761368"
    }
  }
}
```

### TODO

- Should be able to set `status` and `headers` in the response
- Supporting chunk transfer could be interesting
- Should be able to stream the body even if content-length can be known up front


