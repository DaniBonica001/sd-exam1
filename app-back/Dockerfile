# Utiliza la imagen base de Golang
FROM golang:1.22 AS builder
WORKDIR /usr/src/app

ENV GO111MODULE=on
ENV SAMBAHOST=my-samba-container
ENV SAMBAPORT=445
ENV SAMBAUSER=admin
ENV SAMBAPASSWORD=password1
RUN go mod init app
RUN go clean -modcache
COPY . .
#copia los archivos de configuracion donde esta el archivo dockerfile (.go) al workdir despues de que esto quede copiado se inicializa
RUN go mod tidy

#RUN go mod download && go mod verify
EXPOSE 9000

RUN go build -v -o /usr/local/bin/app .

ENV CONSUL_HTTP_ADDR localhost:8500
#RUN go build -o /go/bin/myapp cmd/main.go

CMD [ "app" ]