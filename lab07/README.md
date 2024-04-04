Для запуска сервера и клиента `go run main.go` в `echo-server/server`, `echo-server/client` соответственно. 
По умолчанию будет запущено решение заданий А-В

Севрер запускается на `localhost:8080`

Для запуска сервера UDP Heartbeat `go run main.go -m heartbeat -t X` - где `X` - предельное время бездействия клиента в секундах

Для запуска клиента(-ов) UDP Heartbeat `go run main.go -m heartbeat -c N` - где `N` - число клиентов, запускающихся одновременно

