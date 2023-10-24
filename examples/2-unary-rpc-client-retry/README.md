# 2-unary-rpc-client-retry
Server is unchanged we still use the one from example 1. The only change to the client is the addition of
```go
grpc.WithDefaultServiceConfig(`
{
	"methodConfig": [
		{
			"name": [
				{ "service": "proto.users.v1.UsersService" }
			],
			"retryPolicy": {
				"maxAttempts": 5,
				"initialBackoff" : "0.2s",
				"maxBackoff": "5s",
				"backoffMultiplier": 3,
				"retryableStatusCodes": [ "UNAVAILABLE" ]
			}
		}
	]
}`),
```

to the `defaultDialOptions` in the `connect` function. Please also have a look to get an idea how you can test that a
retryPolicy is configured correctly.
