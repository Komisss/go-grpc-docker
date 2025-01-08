- установить последний go
- установить docker
- открыть go-proj через любой редактор и запустите команду docker-compose up --build через консоль. Если клиент не запустился, значит он не дождался подключения сервера к бд, необходимо запустить клиент через докер руками

docker-compose down
docker-compose up --build
docker volume ls
docker volume rm go-proj_postgres-data
docker exec -it postgres-db psql -U myuser -d mydb
protoc --go_out=. --go-grpc_out=. service.proto
