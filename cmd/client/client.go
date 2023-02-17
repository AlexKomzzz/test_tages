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
)

const clDir = "./cmd/client/cl_dir/" // директория, в которой хранятся файлы клиента

type Client struct {
	cl  api.FileStorageClient
	req []string
}

func newClient(conn *grpc.ClientConn) *Client {
	return &Client{
		cl:  api.NewFileStorageClient(conn),
		req: make([]string, 0),
	}
}

// получение команды(запроса) из командной строки
func (c *Client) createReq() bool {

	in := bufio.NewReader(os.Stdin)

	req, _, err := in.ReadLine()
	if err != nil {
		log.Fatal("the request failed: ", err)
	}

	request := strings.Split(string(req), " ")

	if len(request) < 2 {
		fmt.Println("not enough arguments")
		return false
	}

	c.req = request

	return true
}

// загрузить файл на сервер
func (c *Client) sendFile() bool {
	// считать название файла
	fileName := createFName(c.req)

	// читать данные их файла клиентской директории
	dataSend, err := os.ReadFile(fmt.Sprintf("%s%s", clDir, fileName))
	if err != nil {
		fmt.Println("error: ", err)
		return false
	}

	fileSend := &api.File{
		Data:     dataSend,
		Filename: fileName,
	}

	_, err = c.cl.SendFile(context.Background(), fileSend)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println("file sent!")

	return true
}

// запросить список файлов с сервера
func (c *Client) getListFiles() bool {

	// запрос на получение списка файлов
	var req empty.Empty
	res, err := c.cl.GetListFiles(context.Background(), &req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	list := res.GetFiles()

	// вывод файлов в консоль
	for _, fInf := range list {
		fmt.Println(fInf)
	}

	return true
}

// запросить файл с сервера по имени
func (c *Client) getFile() bool {

	// определение имени файла из введенного запроса в консоли
	fileName := createFName(c.req)

	req := &api.Req{
		Filename: fileName,
	}

	// запрос файла по имени
	res, err := c.cl.GetFile(context.Background(), req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// получение файла
	receivedFile := res.GetData()

	// сохранение файла в клиентской директории
	if err := saveFile(fileName, receivedFile); err {
		return false
	}

	fmt.Println("file received!")

	return true
}

func createFName(request []string) (fileName string) {

	fileName = request[1]
	if len(request) > 2 {
		for i := 2; i < len(request); i++ {
			fileName = fmt.Sprint(fileName, " ", request[i])
		}
	}
	return
}

func saveFile(fileName string, dataFile []byte) bool {

	// сохранить в дир
	newFile, err := os.Create(fmt.Sprintf("%s%s", clDir, fileName))
	if err != nil {
		fmt.Println("error: ", err)
		fmt.Println("Repeat, pls")
		return true
	}
	defer newFile.Close()

	// сохраняем полученный файл в директории
	_, err = newFile.Write(dataFile)
	// err = os.WriteFile(fmt.Sprintf("%s%s", clDir, fileName), receivedFile, 0)
	if err != nil {
		fmt.Println("error: ", err)
		fmt.Println("Repeat, pls")
		return true
	}

	return false
}
