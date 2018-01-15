# Token-Ring

## Запуск сети
```
go run tokenring.go -n 4 -t 1
```

## Посылка сервисных сообщений
```
go run service.go -node 1 -dst 3
go run service.go -node 2 -action drop
```
