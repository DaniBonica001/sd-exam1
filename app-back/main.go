package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"

    "github.com/stacktitan/smb/smb"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // ParseMultipartForm analiza una solicitud multipart y
    // carga hasta maxMemory bytes en la memoria y el resto en un archivo temporal.
    err := r.ParseMultipartForm(10 << 20) // 10 MB máximo
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // obtén el archivo del campo 'file' del formulario
    file, handler, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Error al obtener el archivo del formulario", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Creamos un archivo temporal para guardar el archivo
    tempFile, err := os.CreateTemp("./uploads", "upload-*.tmp")
    if err != nil {
        http.Error(w, "Error al crear el archivo temporal", http.StatusInternalServerError)
        return
    }
    defer tempFile.Close()

    // Copiamos el contenido del archivo recibido al archivo temporal
    _, err = io.Copy(tempFile, file)
    if err != nil {
        http.Error(w, "Error al copiar el contenido del archivo", http.StatusInternalServerError)
        return
    }

    // Obtenemos la extensión del archivo
    ext := filepath.Ext(handler.Filename)

    // Establecemos la conexión a la base de datos MongoDB
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, err := mongo.Connect(r.Context(), clientOptions)
    if err != nil {
        http.Error(w, "Error al conectar con la base de datos MongoDB", http.StatusInternalServerError)
        return
    }
    defer client.Disconnect(r.Context())

    // Obtenemos la colección
    collection := client.Database("test").Collection("files")

    // Creamos un registro para guardar en MongoDB
    document := bson.M{
        "filename": handler.Filename,
        "size":     handler.Size,
        "extension": ext,
    }

    // Insertamos el registro en MongoDB
    _, err = collection.InsertOne(r.Context(), document)
    if err != nil {
        http.Error(w, "Error al insertar el registro en la base de datos MongoDB", http.StatusInternalServerError)
        return
    }

    // Conexión al recurso compartido SAMBA
    options := smb.Options{
        Host:        "192.168.1.100",
        User:        "tu_usuario",
        Password:    "tu_contraseña",
        Share:       "nombre_del_recurso_compartido",
        Domain:      "",
        EncryptData: true,
    }
    err = options.Connect()
    if err != nil {
        http.Error(w, "Error al conectar con el recurso compartido SAMBA", http.StatusInternalServerError)
        return
    }
    defer options.Close()

    // Subimos el archivo a SAMBA
    sambaFilePath := "/carpeta_en_samba/" + handler.Filename
    sambaFile, err := options.Create(sambaFilePath)
    if err != nil {
        http.Error(w, "Error al crear el archivo en SAMBA", http.StatusInternalServerError)
        return
    }
    defer sambaFile.Close()

    // Copiamos el contenido del archivo temporal al archivo en SAMBA
    tempFile.Seek(0, 0)
    _, err = io.Copy(sambaFile, tempFile)
    if err != nil {
        http.Error(w, "Error al copiar el contenido del archivo a SAMBA", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Archivo subido con éxito a SAMBA y metadatos guardados en MongoDB.")
}

func main() {
    http.HandleFunc("/upload", uploadHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
