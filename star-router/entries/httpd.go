package entries

import (
	"bytes"
	"fmt"
	"strings"
	"net/http"
	"log"

	"github.com/danopia/stardust/star-router/base"
	//"github.com/danopia/stardust/star-router/inmem"
)

// Function that creates a new HTTP server when invoked
func httpdFunc(input base.Entry) (output base.Entry) {
	svc := &httpd{
		root: input.(base.Folder),
		//rayFunc:   input.(base.Function), // TODO
		//tmpFolder: inmem.NewFolder("ray-ssh"),
	}

	http.Handle("/~~/", svc)
	go svc.listen()

	return nil // svc.tmpFolder
}

// Context for a running SSH server
type httpd struct {
	root base.Folder
	//rayFunc   base.Function
	//tmpFolder base.Folder
}

func (e *httpd) listen() {
	host := fmt.Sprint("localhost:", 9234)
	log.Printf("Listening on %s...", host)
	if err := http.ListenAndServe(host, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (e *httpd) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// r.Method, r.URL, r.Proto, r.Header, r.Body, r.Host, r.Form, r.RemoteAddr

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	handle := base.NewDetachedHandle(e.root)

	path := strings.TrimPrefix(r.RequestURI, "/~~")
	isDir := true
	if len(path) > 1 {
		isDir = strings.HasSuffix(path, "/")
		if isDir {
			path = strings.TrimSuffix(path, "/")
		}
	}
	log.Println("HTTP request for", path, "- isdir:", isDir)

	if ok := handle.Walk(path); !ok {
		http.Error(w, "Name not found", http.StatusNotFound)
		return
	}

	// If trailing slash, go right to folder mode
	if isDir {
		entry, ok := handle.GetFolder()
		if !ok {
			http.Error(w, "Folder not found", http.StatusNotFound)
			return
		}

		var buffer bytes.Buffer
		buffer.WriteString("<!doctype html><title>")
		buffer.WriteString(entry.Name())
		buffer.WriteString("</title>")

		path = "/~~"
		buffer.WriteString("<h3>")
		for idx, entry := range handle.Stack() {
			if idx > 0 {
				path = fmt.Sprintf("%s/%s", path, entry.Name())
			}
			if idx > 1 {
				buffer.WriteString(" / ")
			}
			
			buffer.WriteString("<a href=\"")
			buffer.WriteString(path)
			buffer.WriteString("/\">")
			buffer.WriteString(entry.Name())
			buffer.WriteString("</a> ")
		}
		buffer.WriteString("</h3>")

		buffer.WriteString("<ul>")
		for _, name := range entry.Children() {
			buffer.WriteString("<li><a href=\"")
			buffer.WriteString(name)
			buffer.WriteString("\">")
			buffer.WriteString(name)
			buffer.WriteString("</a></li>")
		}
		buffer.WriteString("</ul>")

		w.Header().Add("content-type", "text/html; charset=UTF-8")
		w.Write(buffer.Bytes())
		return
	}

	entry := handle.Get()
	switch entry := entry.(type) {

	case base.String:
		value, _ := entry.Get()
		w.Write([]byte(value))

	case base.Folder:
		// not in dir mode, redirect
		http.Redirect(w, r, fmt.Sprintf("%s/", r.RequestURI), http.StatusFound)

	default:
		http.Error(w, "Name cannot be rendered", http.StatusNotImplemented)
	}


	/*

	w.Header().Add("access-control-allow-origin", "*")
	w.Header().Add("cache-control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Add("content-type", "application/json; charset=UTF-8")
	w.Header().Add("vary", "origin")

	payload, err := json.Marshal(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
*/
}
