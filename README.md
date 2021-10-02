# http.sh

*Integrate a HTTP server into your shell pipelines*

Not quite sure where this is going to go, but... let's see!

This is a small CLI which writes HTTP requests to its STDOUT. It reads
potential responses for these requests from its STDIN. That part is counter
intuitive, but it works out well.

The expected usage is to pipe a subscription to a message queue into STDIN and
to pipe STDOUT to a message queue writer.

You can simulate a subscription to a message queue with `tail -F responses` and
writing to a message queue with `tee requests`.

## Example usage

Run the HTTP server in window 1:

```
$ cp /dev/null responses && tail -F responses | go run . | tee requests | jq
```

Create a request in window 2:

```
$ curl localhost:8080
```

You should see a HTTP request in the requests file, with a request_id.

In window 3, run a handler which writes responses:

```
# unbuffer jq by default. Life's better that way.
$ alias jq='jq --unbuffered'
$ tail -F requests | jq -c 'select(.app == "http.request") | .content | {request_id, "body": ("ok" | @base64)}' > responses
```

The curl in window 2 should now return with `ok`.

## TODO

If this does go somewhere, need to:

- Should be able to set `status` and `headers` in the response
- Support chunk transfer


