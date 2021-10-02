# http.sh

*Integrate a HTTP server into your shell pipelines*

Not quite sure where this is going to go...

This is a small CLI which writes HTTP requests to its STDOUT. It reads
potential responses for these requests from its STDIN. That part is counter
intuitive, but it works out well.

Example usage, run the HTTP server in window 1:

```
$ tail -F responses | go run . | tee requests
tail: cannot open 'responses' for reading: No such file or directory
```

In window 2:

```
$ curl localhost:8080
```

You should see a HTTP request in the requests file, with a request_id.

In window 3, run a handler which writes responses:

```
tail -F requests | jq -c 'select(.app == "http.request") | .content | {request_id, "body": ("ok" | @base64)}' > responses
```

The curl in window 2 should now return with `ok`.

## TODO

If this does go somewhere, need to:

- Formalize the packets written to STDOUT and read from STDIN
- It'll likely be nice if the `body` in the response can be a plain string
- Should be able to set `status` and `headers` in the response
- Support chunk transfer


