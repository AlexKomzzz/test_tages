package libraryserver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AlexKomzzz/test_tages/pkg/api"
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
	setBtime    map[string]string // мапа для хранения времени создания файлов
}

func NewGRPCServer() *GRPCserver {
	return &GRPCserver{
		limUpFile:   make(chan struct{}, limUp),
		limDownFile: make(chan struct{}, limDown),
		limViewFile: make(chan struct{}, limView),
		chErr:       make(chan error),
		setBtime:    make(map[string]string),
	}
}

// Получение файла от клиента
func (s *GRPCserver) SendFile(ctx context.Context, fileData *api.File) (*empty.Empty, error) {
	fileName := fmt.Sprintf("%s%s", fileDir, fileData.Filename)

	// канал для блокировки текущей функции для ожидания выполнения горутины
	chOk := make(chan struct{})

	s.limUpFile <- struct{}{}

	go s.saveFile(fileData.Data, fileName, chOk)
	defer func() {
		<-s.limUpFile // или будет лучше делать в горутине?
	}()

	select {
	case err := <-s.chErr:
		return nil, err
	case <-chOk:
		var res empty.Empty
		return &res, nil
	}
}

func (s *GRPCserver) saveFile(dataFile []byte, fileName string, chOk chan<- struct{}) {
	defer func() { // сигнал о завершении текущей горутины
		chOk <- struct{}{}
	}()

	// проверка существования файла на сервере
	if ok, err := checkFile(fileName); ok && err != nil { // ошибка при проверке
		s.chErr <- err
		return
	} else if !ok && err != nil { // такого файла еще нет на сервере
		// запись времеми создания файла
		s.setBtime[fileName] = time.Now().Format("2006-01-02 15:04:05")
	}

	file, err := os.Create(fileName)
	if err != nil {
		logrus.Println("error SendFile/Create: ", err)
		s.chErr <- err
		return
	}

	defer file.Close()

	// сохраняем полученный файл в директории
	_, err = file.Write(dataFile)
	// err := os.WriteFile(fmt.Sprintf("%s%s", fileDir, fileData.Filename), fileData.Data, 0)
	if err != nil {
		logrus.Println("error SendFile/Write: ", err)
		s.chErr <- err
		return
	}
}

// Отправка списка файлов
func (s *GRPCserver) GetListFiles(ctx context.Context, empty *empty.Empty) (*api.ListFiles, error) {

	s.limViewFile <- struct{}{}

	chRes := make(chan []string)

	go s.getList(chRes)
	defer func() {
		<-s.limViewFile // или будет лучше делать в горутине?
	}()

	result := <-chRes

	return &api.ListFiles{Files: result}, nil
}

func (s *GRPCserver) getList(chRes chan<- []string) {

	result := make([]string, 0)

	// получение данных о файлах в директории
	files, err := os.ReadDir(fileDir)
	if err != nil {
		logrus.Println("error GetListFiles/ReadDir: ", err)
		s.chErr <- err
		return
	}

	// если файлов нет
	if len(files) == 0 {
		s.chErr <- errors.New("files not found")
		return
	}

	for _, file := range files {

		fileName := fmt.Sprintf("%s%s", fileDir, file.Name())

		// время создания файла
		btime, ok := s.setBtime[fileName]
		if !ok {
			btime = "empty"
		}

		// время изменения можно получить из fileInfo
		fInf, err := file.Info()
		if err != nil {
			logrus.Println("error GetListFiles/file.Info: ", err)
			s.chErr <- err
			return
		}

		infoFile := fmt.Sprintf("%s | %v | %v", file.Name(), btime, fInf.ModTime().Format("2006-01-02 15:04:05"))
		result = append(result, infoFile)

	}

	chRes <- result
}

// отправка файла по имени
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

	go s.getFile(chRes, filename)
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

func (s *GRPCserver) getFile(chRes chan<- []byte, fileName string) {

	// проверка наличия файла на сервере
	_, err := checkFile(fileName)
	if err != nil {
		s.chErr <- err
	}

	dataFile, err := os.ReadFile(fileName)
	if err != nil {
		logrus.Println("error GetFile/ReadFile: ", err)
		s.chErr <- err
		return
	}

	chRes <- dataFile
}

// проверка наличия файла на сервере
func checkFile(fileName string) (bool, error) {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Println("file does not exist") // это_true
			return false, errors.New("file does not exist")
		} else {
			logrus.Println("error GetFile/Stat: ", err)
			return true, err
		}
	}

	return true, nil
}
