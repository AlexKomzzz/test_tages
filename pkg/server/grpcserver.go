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
	limUpFile   chan struct{}     // семафор на ограничение загрузки файлов
	limDownFile chan struct{}     // семафор на ограничение скачиваний файлов
	limViewFile chan struct{}     // семафор на ограничение просмотра списка файлов
	setBtime    map[string]string // мапа для хранения времени создания файлов
}

func NewGRPCServer() *GRPCserver {
	return &GRPCserver{
		limUpFile:   make(chan struct{}, limUp),
		limDownFile: make(chan struct{}, limDown),
		limViewFile: make(chan struct{}, limView),
		setBtime:    make(map[string]string),
	}
}

// Получение файла от клиента
func (s *GRPCserver) SendFile(ctx context.Context, fileData *api.File) (*empty.Empty, error) {

	// формируем полное имя файла - относительный путь
	fileName := fmt.Sprintf("%s%s", fileDir, fileData.Filename)

	// канал для блокировки текущей функции для ожидания выполнения горутины
	chOk := make(chan struct{})
	// канал для возвращения ошибки
	chErr := make(chan error)

	// показываем, что обрабатывается очередной запрос, либо ждем, если превышен лимит
	s.limUpFile <- struct{}{}

	// создаем горутину для обработки запроса
	go s.saveFile(fileData.Data, fileName, chOk, chErr)
	defer func() {
		<-s.limUpFile // по окончании работы освобождаем буффер, сообщая о завершении работы текущего запроса
	}()

	// ждем завершения работы горутины, либо ошибки
	select {
	case err := <-chErr:
		return nil, err
	case <-chOk:
		var res empty.Empty
		return &res, nil
	}
}

func (s *GRPCserver) saveFile(dataFile []byte, fileName string, chOk chan<- struct{}, chErr chan<- error) {

	// проверка существования файла на сервере
	if ok, err := checkFile(fileName); ok && err != nil { // ошибка при проверке
		chErr <- err
		return
	} else if !ok && err != nil { // такого файла нет на сервере
		// запись времени создания файла
		s.setBtime[fileName] = time.Now().Format("2006-01-02 15:04:05")
	}

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

	// сигнал об успешном завершении текущей горутины
	chOk <- struct{}{}
}

// Отправка списка файлов
func (s *GRPCserver) GetListFiles(ctx context.Context, empty *empty.Empty) (*api.ListFiles, error) {

	// показываем, что обрабатывается очередной запрос, либо ждем, если превышен лимит
	s.limViewFile <- struct{}{}

	chRes := make(chan []string)
	chErr := make(chan error)

	// создаем горутину для обработки запроса
	go s.getList(chRes, chErr)
	defer func() { // по окончании работы освобождаем буффер, сообщая о завершении работы текущего запроса
		<-s.limViewFile
	}()

	// ждем возвращения результата работы горутины, либо ошибки
	select {
	case err := <-chErr:
		return nil, err
	case result := <-chRes:
		return &api.ListFiles{Files: result}, nil
	}
}

func (s *GRPCserver) getList(chRes chan<- []string, chErr chan<- error) {

	result := make([]string, 0)

	// получение данных о файлах в директории
	files, err := os.ReadDir(fileDir)
	if err != nil {
		logrus.Println("error GetListFiles/ReadDir: ", err)
		chErr <- err
		return
	}

	// если файлов нет
	if len(files) == 0 {
		chErr <- errors.New("files not found")
		return
	}

	// рассмотрим каждый файл
	for _, file := range files {

		// относительный путь файла
		fileName := fmt.Sprintf("%s%s", fileDir, file.Name())

		// время создания файла
		// сохраняется в мапе при первой загрузке на сервер
		btime, ok := s.setBtime[fileName]
		if !ok {
			btime = "empty"
		}

		// время изменения можно получить из fileInfo
		fInf, err := file.Info()
		if err != nil {
			logrus.Println("error GetListFiles/file.Info: ", err)
			chErr <- err
			return
		}

		infoFile := fmt.Sprintf("%s | %v | %v", file.Name(), btime, fInf.ModTime().Format("2006-01-02 15:04:05"))
		result = append(result, infoFile)

	}

	// возвращяем результат
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
	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// формируем полное имя файла - относительный путь
	filename := fmt.Sprintf("%s%s", fileDir, fileName.Filename)

	chRes := make(chan []byte)
	chErr := make(chan error)

	s.limDownFile <- struct{}{} // блокирует, если уже запущено 10 горутин

	go getFile(chRes, filename, chErr)
	defer func() {
		<-s.limDownFile
	}()

	select {
	case dataFile := <-chRes:
		return &api.File{Data: dataFile}, nil
	case err := <-chErr:
		return nil, err
	}
}

func getFile(chRes chan<- []byte, fileName string, chErr chan<- error) {

	// проверка наличия файла на сервере
	_, err := checkFile(fileName)
	if err != nil {
		chErr <- err
	}

	// читаем файл
	dataFile, err := os.ReadFile(fileName)
	if err != nil {
		logrus.Println("error GetFile/ReadFile: ", err)
		chErr <- err
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
