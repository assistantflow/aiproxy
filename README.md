# AI Proxy

A Gin-powered AI proxy.

## How to use

To start a proxy server with a custom prefix, use the following command:
```
$ go run ./cmd/proxy/main.go --prefix=openai
```

To make a request, use the following command:
```
$ curl http://localhost:8080/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

The response will look like this:

{
  "id": "chatcmpl-6uaXSSLw3i6dn428HYT2835KJ14Qh",
  "object": "chat.completion",
  "created": 1678944842,
  "model": "gpt-3.5-turbo-0301",
  "usage": {
    "prompt_tokens": 9,
    "completion_tokens": 10,
    "total_tokens": 19
  },
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "\n\nHello! How may I assist you today?"
      },
      "finish_reason": "stop",
      "index": 0
    }
  ]
}
```
Please note that you need to replace `$OPENAI_API_KEY` with your actual OpenAI API key.

## Expose promtheus metrics

```
$ curl http://localhost:2112/metrics

The response will look like this:

# HELP token_used_total How many token used, partitioned by api key.
# TYPE token_used_total counter
token_used_total{key="$OPENAI_API_KEY"} 12
```
