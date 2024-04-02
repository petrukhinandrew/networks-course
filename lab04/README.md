Для запуска прокси 

```bash
cd proxy
go run main.go -h host -p port
```

По умолчанию `host == localhost`, `port = 8080`

По умолчанию, `//` в запросе вернет 301, это чинится через `proxy/slashfix`

Для тестов: `curl -v localhost:8080/http:/emkn.ru`, `curl -v localhost:8080/http://emkn.ru`

Для тестирования методов есть mockserver 

```bash
cd mockserver
go run main.go
```

На любой `GET` возвращает 404, на любой `POST` - 200 и тело запроса. Запускается на `localhost:8000`

Пример теста с запущенными прокси и тестовым сервером: 

```bash
curl -v localhost:8080/http://localhost:8000
curl -v -d "lol:kek" localhost:8080/http://localhost:8000
```
Ожидаются ответы `404` и `200 lol:kek` соответственно