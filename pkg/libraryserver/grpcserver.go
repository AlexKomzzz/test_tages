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
	fileDir = "./srvDir/"
)

type GRPCserver struct {
	api.UnimplementedFileStorageServer
}

func NewGRPCServer() *GRPCserver {
	return &GRPCserver{}
}

// Получение файла от клиента
func (s *GRPCserver) SendFile(ctx context.Context, fileData *api.File) (*empty.Empty, error) {

	// сохраняем полученный файл в директории
	err := os.WriteFile(fmt.Sprintf("%s%s", fileDir, fileData.Filename), fileData.Data, 0)
	if err != nil {
		logrus.Println("error SendFile/WriteFile: ", err)
		return nil, err
	}

	return nil, nil
}

// Отправка списка файлов
func (s *GRPCserver) GetListFiles(ctx context.Context, empty *empty.Empty) (*api.ListFiles, error) {
	result := make([]string, 0)

	files, err := os.ReadDir(fileDir)
	if err != nil {
		logrus.Println("error GetListFiles/ReadDir: ", err)
		return nil, err
	}

	for _, file := range files {

		// время создания и изменения файла
		t, err := times.Stat(fmt.Sprintf("%s%s", fileDir, file.Name()))
		if err != nil {
			logrus.Println("error GetListFiles/times.Stat: ", err)
			return nil, err
		}
		// if t.HasBirthTime() {
		// 	log.Println(t.BirthTime())
		// }

		// время изменения можно получить из fileInfo

		infoFile := fmt.Sprintf("%s | %v | %v", file.Name(), t.BirthTime(), t.ModTime())
		result = append(result, infoFile)
	}

	return &api.ListFiles{Files: result}, nil
}

func (s *GRPCserver) GetFile(ctx context.Context, fileName *api.Req) (*api.File, error) {

	// для ограничения подключений реализовать счетчик работающих горутин на передачу файлов
	// ожидать, если кол-во больше 9
	// можно в структуре s иметь канал с буфером 9
	// здесь перед запуском горутины на передачу файла записывать пустую структуру в канал, если он переполнен,
	// то текущая горутина будет заблокирована до завершения какой-либо работающей

	// После запершения горутины (которая отправит файл) вычитаем из канала одну структуру, как сигнал о завершении

	// получаем имя файла, находим его в директории и возвращаем его

	// result := make([]byte, 0)

	filename := fmt.Sprintf("%s%s", fileDir, fileName.Filename)

	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Println("file does not exist") // это_true
			return nil, errors.New("file does not exist")
		} else {
			logrus.Println("error GetFile/Stat: ", err)
			return nil, err
		}
	}

	dataFile, err := os.ReadFile(filename)
	if err != nil {
		logrus.Println("error GetFile/ReadFile: ", err)
		return nil, err
	}

	return &api.File{Data: dataFile}, nil
}
