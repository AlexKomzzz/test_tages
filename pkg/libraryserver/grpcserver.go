package libraryserver

import (
	"context"

	"github.com/AlexKomzzz/test_tages/pkg/api"
)

const (

	// собственные коды ошибок
	codeOK  = 5
	codeErr = 2

	// статусы ответов
	statusOk  = "successfully"
	statusErr = "error"
)

type GRPCserver struct {
	api.UnimplementedFileStorageServer
}

func NewGRPCServer() *GRPCserver {
	return &GRPCserver{}
}

// Получение файла от клиента
func (s *GRPCserver) SendFile(ctx context.Context, fileData *api.File) (*api.Resp, error) {

	// сохраняем полученный файл в директории
	// если все ОК, отвечаем ОК, при ошибке ERR

	return &api.Resp{Code: codeOK, Status: statusOk}, nil
}

// Отправка списка файлов
func (s *GRPCserver) GetListFiles(ctx context.Context, Nil *api.Nil) (*api.ListFiles, error) {

	// os.Dir проходит по директории и возвращает имена файлов
	result := make([]string, 0)

	return &api.ListFiles{Files: result}, nil
}

func (s *GRPCserver) GetFile(ctx context.Context, fileName *api.Req) (*api.File, error) {

	// получаем имя файла, находим его в директории и возвращаем его

	result := make([]byte, 0)

	return &api.File{Data: result}, nil
}
