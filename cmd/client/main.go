package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AlexKomzzz/test_tages/pkg/api"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	clntPort = ":8080"
	clDir    = "./cmd/client/cl_dir/" // директория, в которой хранятся файлы клиента
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
			fmt.Println("not enough arguments")
			continue
		}

		switch {
		case request[0] == "send": // если запрос send значит вызываем метод загрузки файла. Дальше указывается имя файла

			// считываем название файла
			fileName := request[1]
			if len(request) > 2 {
				for i := 2; i < len(request); i++ {
					fileName = fmt.Sprint(fileName, " ", request[i])
				}
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
				fmt.Println(err)
				continue
			}

			fmt.Println("file sent!")

		case request[0] == "list" && request[1] == "files": // если запрос list files значит вызываем метод передачи списка всех файлов
			var req empty.Empty
			res, err := client.GetListFiles(context.Background(), &req)
			if err != nil {
				fmt.Println(err)
				continue
			}

			list := res.GetFiles()

			for _, fInf := range list {

				fmt.Println(fInf)
			}

		case request[0] == "get": // если запрос file значит вызываем метод передачи файла. Дальше указывается имя файла

			fileName := request[1]
			if len(request) > 2 {
				for i := 2; i < len(request); i++ {
					fileName = fmt.Sprint(fileName, " ", request[i])
				}
			}

			req := &api.Req{
				Filename: fileName,
			}

			res, err := client.GetFile(context.Background(), req)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// полученный файл
			receivedFile := res.GetData()

			// сохранить в дир
			newFile, err := os.Create(fmt.Sprintf("%s%s", clDir, fileName))
			if err != nil {
				fmt.Println("error: ", err)
				fmt.Println("Repeat, pls")
				continue
			}

			// сохраняем полученный файл в директории
			_, err = newFile.Write(receivedFile)
			// err = os.WriteFile(fmt.Sprintf("%s%s", clDir, fileName), receivedFile, 0)
			if err != nil {
				fmt.Println("error: ", err)
				fmt.Println("Repeat, pls")
				newFile.Close()
				continue
			}

			newFile.Close()
			fmt.Println("file received!")

		default:
			fmt.Println("unknown command! Wait: 'send' 'list files' 'get'")
			continue
		}

	}
}
