package main

import (
	"net/http"
	"os"
	"io"
	"io/ioutil"
	"log"
	"html/template"
	"path"
	"runtime/debug"
	"errors"
)
const (
	UPLOAD_DIR = "./uploads"
)

const (
	TEMPLATE_DIR = "./views"
)

const (
	ListDir = 0x0001
)
var templates = make(map[string]*template.Template)

func staticDirHandle(mux *http.ServeMux, prefix string, staticDir string, flags int) {
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		file := staticDir + r.URL.Path[len(prefix)-1 :]
		if (flags & ListDir) == 0 {
			if exits := isExists(file); !exits {
				http.NotFound(w, r)
				return
			}
		}
		http.ServeFile(w, r, file)
	})
}

func Init() {

	fileInfoArr, err := ioutil.ReadDir(TEMPLATE_DIR)
	if err != nil {
		panic(err)
		return
	}

	var templateName, templatePath string

	for _, fileInfo := range fileInfoArr  {
		templateName = fileInfo.Name()
		if ext := path.Ext(templateName); ext != ".html" {
			continue
		}
		templatePath = TEMPLATE_DIR + "/" + templateName
		log.Println("Loading template:", templatePath)
		t := template.Must(template.ParseFiles(templatePath))
		templates[templateName] = t
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}

}

func renderHtml(w http.ResponseWriter, tmpl string, locals map[string]interface{}) error {
	t, ok := templates[tmpl]
	if !ok {
		return errors.New("no templates")
	}
	err :=t.Execute(w, locals)
	return err
}

func uploadHandler(w http.ResponseWriter, r * http.Request)  {
	if r.Method == "GET" {

		//读取指定模版的内容
		if err := renderHtml(w, "upload.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//根据模版语法渲染
		return
	}

	if r.Method == "POST" {
		//return the first from the image
		f, h, err := r.FormFile("image")
		check(err)
		//h:Fileheader describes the file info
		filename := h.Filename
		//f:a interface to access the file
		defer f.Close()

		t, err := os.Create(UPLOAD_DIR + "/" + filename)
		check(err)
		defer t.Close()

		//t.writer, f.reader
		_, err = io.Copy(t, f)
		check(err)

		//redirect a new url
		//StatusFound RFC
		http.Redirect(w, r, "/view?id=" + filename, http.StatusFound)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	imageId := r.FormValue("id")
	imagePath := UPLOAD_DIR + "/" + imageId
	if exits := isExists(imagePath); !exits {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image")
	http.ServeFile(w, r, imagePath)
}

func isExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

func listHandler(w http.ResponseWriter, r *http.Request)  {
	fileInfoArr, err := ioutil.ReadDir("./uploads")
	check(err)
	locals := make(map[string] interface{})
	var images []string


	for _, fileInfo := range fileInfoArr {
		images = append(images, fileInfo.Name())
	}

	locals["images"] = images
	err = renderHtml(w, "list.html", locals)
	check(err)
}

func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(),http.StatusInternalServerError)
				log.Println("Warn: panic in ", fn, e)
				log.Println(string(debug.Stack()))
			}
		}()
		fn(w, r)
	}
}



func main() {

	Init()
	mux := http.NewServeMux()
	staticDirHandle(mux, "/assets/", "./public", 0)
	http.HandleFunc("/", safeHandler(listHandler))
	http.HandleFunc("/view", safeHandler(viewHandler))
	http.HandleFunc("/upload", safeHandler(uploadHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("listenAndServe: ", err.Error())
	}
}


