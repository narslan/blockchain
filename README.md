### Making curl request to the api

```sh
curl http://localhost:4443/chain | jq
curl -d '{"sender":"you", "recipient":"me","amount": 1}' -H "Content-Type: application/json" -X POST http://localhost:4443/transaction/new

```
