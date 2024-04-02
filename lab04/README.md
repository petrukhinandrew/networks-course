Для запуска прокси 

```bash
cd proxy
go run main.go -h host -p port
```

По умолчанию `host == localhost`, `port = 8080`

По умолчанию, `//` в запросе вернет 301, это чинится через `proxy/slashfix`

Для тестов: `curl -v localhost:8080/http:/emkn.ru`, `curl -v localhost:8080/http://emkn.ru`
