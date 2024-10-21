goose -dir ./migrations postgres "postgres://postgres:password@localhost:5432/restRefs?sslmode=disable" status

goose -dir ./migrations postgres "postgres://postgres:password@localhost:5432/restRefs?sslmode=disable" up