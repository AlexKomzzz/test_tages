package libraryserver

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/AlexKomzzz/test_tages/pkg/api"
	"github.com/djherbis/times"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
)

const (
	fileDir = "./pkg/server/srvDir/"
	limUp   = 10
	limDown = 10
	limView = 100
)

type GRPCserver struct {
	api.UnimplementedFileStorageServer
	limUpFile   chan struct{} // семафор на ограничение загрузки файлов
	limDownFile chan struct{} // семафор на ограничение скачиваний файлов
	limViewFile chan struct{} // семафор на ограничение просмотра списка файлов
	chErr       chan error
}

func NewGRPCServer() *GRPCserver {
	return &GRPCserver{
		limUpFile:   make(chan struct{}, limUp),
		limDownFile: make(chan struct{}, limDown),
		limViewFile: make(chan struct{}, limView),
		chErr:       make(chan error),
	}
}

// Получение файла от клиента
func (s *GRPCserver) SendFile(ctx context.Context, fileData *api.File) (*empty.Empty, error) {
	fileName := fmt.Sprintf("%s%s", fileDir, fileData.Filename)

	s.limUpFile <- struct{}{}

	go saveFile(fileData.Data, fileName, s.chErr)
	defer func() {
		<-s.limUpFile // или будет лучше делать в горутине?
	}()

	select {
	case err := <-s.chErr:
		return nil, err
	default:
		var res empty.Empty
		return &res, nil
	}
}

func saveFile(dataFile []byte, fileName string, chErr chan<- error) {
	file, err := os.Create(fileName)
	if err != nil {
		logrus.Println("error SendFile/Create: ", err)
		chErr <- err
		return
	}

	defer file.Close()

	// сохраняем полученный файл в директории
	_, err = file.Write(dataFile)
	// err := os.WriteFile(fmt.Sprintf("%s%s", fileDir, fileData.Filename), fileData.Data, 0)
	if err != nil {
		logrus.Println("error SendFile/Write: ", err)
		chErr <- err
		return
	}
}

// Отправка списка файлов
func (s *GRPCserver) GetListFiles(ctx context.Context, empty *empty.Empty) (*api.ListFiles, error) {

	s.limViewFile <- struct{}{}

	chRes := make(chan []string)

	go getList(chRes, s.chErr)
	defer func() {
		<-s.limViewFile // или будет лучше делать в горутине?
	}()

	result := <-chRes

	return &api.ListFiles{Files: result}, nil
}

func getList(chRes chan<- []string, chErr chan<- error) {

	result := make([]string, 0)

	files, err := os.ReadDir(fileDir)
	if err != nil {
		logrus.Println("error GetListFiles/ReadDir: ", err)
		chErr <- err
		return
	}

	if len(files) == 0 {
		chErr <- errors.New("files not found")
		return
	}

	for _, file := range files {

		// время создания и изменения файла
		t, err := times.Stat(fmt.Sprintf("%s%s", fileDir, file.Name()))
		if err != nil {
			logrus.Println("error GetListFiles/times.Stat: ", err)
			chErr <- err
			return
		}
		if !t.HasBirthTime() {
			logrus.Println("отсутствует время создания файла")
			infoFile := fmt.Sprintf("%s | %v | %v", file.Name(), "NULL", t.ModTime().Format("2006-01-02 15:04:05"))
			result = append(result, infoFile)
		} else {

			// время изменения можно получить из fileInfo

			infoFile := fmt.Sprintf("%s | %v | %v", file.Name(), t.BirthTime(), t.ModTime().Format("2006-01-02 15:04:05"))
			result = append(result, infoFile)
		}
	}

	chRes <- result
}

func (s *GRPCserver) GetFile(ctx context.Context, fileName *api.Req) (*api.File, error) {

	// для ограничения подключений реализовать счетчик работающих горутин на передачу файлов
	// ожидать, если кол-во больше 9
	// можно в структуре s иметь канал с буфером 10
	// здесь перед запуском горутины на передачу файла записывать пустую структуру в канал, если он переполнен,
	// то текущая горутина будет заблокирована до завершения какой-либо работающей

	// После запершения горутины (которая отправит файл) вычитаем из канала одну структуру, как сигнал о завершении

	filename := fmt.Sprintf("%s%s", fileDir, fileName.Filename)

	chRes := make(chan []byte)

	s.limDownFile <- struct{}{} // блокирует, если уже запущено 10 горутин
	// logrus.Println("len semafor: ", len(s.limDownFile)) // тестовая

	go getFile(chRes, filename, s.chErr)
	defer func() {
		<-s.limDownFile // или будет лучше делать в горутине?
	}()

	select {
	case dataFile := <-chRes:
		return &api.File{Data: dataFile}, nil
	case err := <-s.chErr:
		return nil, err
	}
}

func getFile(chRes chan<- []byte, filename string, chErr chan<- error) {

	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Println("file does not exist") // это_true
			chErr <- errors.New("file does not exist")
			return
		} else {
			logrus.Println("error GetFile/Stat: ", err)
			chErr <- err
			return
		}
	}

	dataFile, err := os.ReadFile(filename)
	if err != nil {
		logrus.Println("error GetFile/ReadFile: ", err)
		chErr <- err
		return
	}

	// time.Sleep(time.Second * 10) // для отладки

	chRes <- dataFile
}
