https://github.com/AlexKomzzz/library-app.git

Задание: Спроектировать базу данных, в которой содержится авторы
книг и сами книги. Необходимо написать сервис который будет по 
автору искать книги, а по книге искать её авторов.
Требования к сервису: 
 Сервис должен принимать запрос по GRPC.
 Должна быть использована база данных MySQL
 Код сервиса должен быть хорошо откомментирован
 Код должен быть покрыт unit тестами
 В сервисе должен лежать Dockerfile, для запуска базы данных с
тестовыми данными
 Должна быть написана документация, как запустить сервис. 
Плюсом будет если в документации будут указания на команды, 
для запуска сервиса и его окружения, через Makefile 
 код должен быть выложен на github.


### Создание protoc
    make protoc
    
либо

    protoc -I=grpc_api/proto               \
            --go_out=. --go-grpc_out=.\
            grpc_api/proto/library.proto

### Запрос клиента должен состоять из:
1. ключевого слова ("authors" - если ищем авторов по книге, или "book" - если ищем книгу по авторам)
2. после ключевого слова следуют аргументы в виде названия книги (через пробелы), либо фамилии авторов

Пример запроса: authors Harry Potter
    result: Rowling