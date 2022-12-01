package main

import (
	"bufio"
	"cloud.google.com/go/firestore"
	cloud "cloud.google.com/go/storage"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/go-chi/chi"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	//GetImgUrl()
	router := chi.NewRouter()
	router.Route("/", func(api chi.Router) {
		api.Get("/imgUrl", GetImgUrl)
		api.Get("/upload", UploadHandler)
	})
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}
}
func GetImgUrl(w http.ResponseWriter, r *http.Request) {
	readFile, err := os.Open("levels.txt")
	if err != nil {
		fmt.Println(err)
	}
	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string
	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}
	readFile.Close()
	size := len(fileLines)
	rand.Seed(time.Now().UTC().UnixNano())
	indx := rand.Intn(size)
	line := fileLines[indx]
	img := line[6:]
	mp := make(map[string]string)
	mp["url"] = "https://storage.cloud.google.com/squizgame-2ac93.appspot.com/Images/" + img + ".png"
	mp["name"] = img
	jsonResponse, err := json.Marshal(mp)
	w.Write(jsonResponse)
	return
}

type App struct {
	Ctx     context.Context
	Client  *firestore.Client
	Storage *cloud.Client
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	files, err := ioutil.ReadDir("./WordPictures")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if f.IsDir() {
			ff, err := ioutil.ReadDir("./WordPictures/" + f.Name())
			if err != nil {
				log.Fatal(err)
			}
			for _, fff := range ff {
				fName := fff.Name()
				if strings.Contains(fName, ".jpg") {
					//fmt.Println(fName)
					FireBaseUpload("./WordPictures/"+f.Name()+"/"+fName, fff.Name())
				}
			}

		}
	}
}
func FireBaseUpload(imgFile, imagePath string) error {
	route := App{}
	route.Ctx = context.Background()
	serviceKey := os.Getenv("serviceKey")
	sa := option.WithCredentialsJSON([]byte(serviceKey))
	app, err := firebase.NewApp(route.Ctx, nil, sa)
	if err != nil {
		return err
	}
	route.Client, err = app.Firestore(route.Ctx)
	if err != nil {
		return err
	}
	route.Storage, err = cloud.NewClient(route.Ctx, sa)
	if err != nil {
		return err
	}
	bucket := "squizgame-2ac93.appspot.com"
	var newImg string
	for i := 0; i < len(imagePath)-3; i++ {
		c := imagePath[i]
		newImg = newImg + string(c)
	}
	newImg = newImg + "png"
	fmt.Println(newImg)
	wc := route.Storage.Bucket(bucket).Object("Images/" + newImg).NewWriter(route.Ctx)
	file, err := os.Open(imgFile)
	if err != nil {
		fmt.Println(err)
	}
	_, err = io.Copy(wc, file)
	if err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}
