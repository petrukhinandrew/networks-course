Вместо позиционных аргументов используются флаги. Сервер можно запустить через `go run main.go -p server_port -m mode -b limit` где `mode` - `simple` (задание А), `par` (задание Б), `bound` (задание С), `limit` - concurrency limit из задания Г. По умолчанию будет запущено решение задачи А. 

Host по умолчанию - `localhost`, порт - `8080`

# A

`go run main.go -p 8080 -m simple`

# Б 

`go run main.go -p 8080 -m par`

# В 

Клиент можно запустить через `go run main.go -h localhost -p 8080 -f Quickstart.md`

# Г

`go run main.go -p 8080 -m bound -b 100`

