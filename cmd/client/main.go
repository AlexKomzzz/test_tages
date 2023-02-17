package main

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	clntPort = ":8080"
)

func main() {

	conn := createConn()
	defer conn.Close()

	client := newClient(conn)

	for {
		// получить запрос из командной строки
		if ok := client.createReq(); !ok {
			continue
		}

		switch {
		case client.req[0] == "send": // если запрос send значит вызываем метод загрузки файла. Дальше указывается имя файла

			if client.sendFile() {
				continue
			}

		case client.req[0] == "list" && client.req[1] == "files": // если запрос list files значит вызываем метод передачи списка всех файлов

			if client.getListFiles() {
				continue
			}

		case client.req[0] == "get": // если запрос file значит вызываем метод передачи файла. Дальше указывается имя файла

			if client.getFile() {
				continue
			}

		default:
			fmt.Println("unknown command! Wait: 'send' 'list files' 'get'")
			continue
		}

	}
}

func createConn() *grpc.ClientConn {

	conn, err := grpc.Dial(clntPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("error connect client: ", err)
	}

	return conn
}
