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
	clDir    = "./cl_dir/" // директория, в которой хранятся файлы клиента
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
		case request[0] == "send": // если запрос send значит вызываем метод загрузки файла. Дальше указывается имя файла

			var fileName string
			// считываем название файла
			for i := 1; i < len(request); i++ {
				fileName = fmt.Sprint(fileName, " ", request[i])
			}

			// открываем файл в клиентской директории
			dataSend, err := os.ReadFile(fmt.Sprintf("%s%s", clDir, fileName))
			if err != nil {
				fmt.Println("error: ", err)
				continue
			}

			fileSend := &api.File{
				Data:     dataSend,
				Filename: fileName,
			}

			_, err = client.SendFile(context.Background(), fileSend)
			if err != nil {
				log.Fatal("Client: error SendFile: ", err)
			}

			fmt.Println("file sent!")

		case request[0] == "list" && request[1] == "files": // если запрос list files значит вызываем метод передачи списка всех файлов

			res, err := client.GetListFiles(context.Background(), nil)
			if err != nil {
				logrus.Fatal("Client: error GetListFiles: ", err)
			}

			fmt.Println(res.GetFiles())

		case request[0] == "get": // если запрос file значит вызываем метод передачи файла. Дальше указывается имя файла

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
				logrus.Fatal("Client: error GetFile: ", err)
			}

			// полученный файл
			receivedFile := res.GetData()
			// сохранить в дир
			// newFile, err := os.Create(fmt.Sprintf("%s%s", clDir, fileName))
			// if err != nil {
			// 	fmt.Println("error: ", err)
			// 	fmt.Println("Repeat, pls")
			// 	continue
			// }

			err = os.WriteFile(fmt.Sprintf("%s%s", clDir, fileName), receivedFile, 0)
			if err != nil {
				fmt.Println("error: ", err)
				fmt.Println("Repeat, pls")
				continue
			}

			fmt.Println("file received!")

		}
	}
}
