Подготовка образа PostgreSQL/запуск контейнера
```bash
docker build -f DockerfilePG -t yp-postgres .
```

Запуск БД
```bash
docker run -d -p 127.0.0.1:5432:5432 --name yp-pg-db yp-postgres
```

Полезные команды
```bash
# Проверить статус контейнера
docker ps

# Посмотреть логи
docker logs yp-pg-db

# Остановить контейнер
docker stop yp-pg-db

# Удалить контейнер (перед повторным запуском с тем же именем)
docker rm yp-pg-db
```
