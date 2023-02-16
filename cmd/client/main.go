package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AlexKomzzz/test_tages/pkg/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	clntPort = ":8080"
)

func main() {
	conn, err := grpc.Dial(clntPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("error connect client: ", err)
	}

	defer conn.Close()

	client := api.NewFileStorageClient(conn)

	for {

		// запросы из командной строки

		in := bufio.NewReader(os.Stdin)

		req, _, err := in.ReadLine()
		if err != nil {
			log.Fatal("the request failed: ", err)
		}

		request := strings.Split(string(req), " ")

		if len(request) < 2 {
			log.Fatal("not enough arguments")
		}

		switch {
		case request[0] == "list" && request[1] == "files": // если запрос list files значит вызываем метод передачи списка всех файлов

			res, err := client.GetListFiles(context.Background(), &api.Nil{})
			if err != nil {
				logrus.Fatal("Client: error GetListFiles: ", err)
			}

			fmt.Println(res.GetFiles())

		case request[0] == "file": // если запрос file значит вызываем метод передачи файла. Длаьше указывается имя файла

			var fileName string
			// считываем название файла
			for i := 1; i < len(request); i++ {
				fileName = fmt.Sprint(fileName, " ", request[i])
			}

			req := &api.Req{
				Filename: fileName,
			}

			res, err := client.GetFile(context.Background(), req)
			if err != nil {
				log.Fatal("Client: error GetFile: ", err)
			}
			fmt.Println(res.GetData())
			// полученный файл
			// file := res.GetData()
			// сохранить в дир
		}
	}
}
