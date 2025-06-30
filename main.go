package main

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/a-h/templ"
	"github.com/lukasmwerner/rezepte/components"
	"github.com/lukasmwerner/rezepte/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

//go:embed assets
var assetsFS embed.FS

func main() {

	md := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
	cwd := os.DirFS(".")

	recipie_files, _ := fs.Glob(cwd, "*.md")
	recipies := map[string]models.Recipe{}
	for _, recipieFileName := range recipie_files {
		f, err := cwd.Open(recipieFileName)
		if err != nil {
			log.Printf("error opening '%s': %s\n", recipieFileName, err.Error())
			continue
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			log.Printf("error reading '%s': %s\n", recipieFileName, err.Error())
			continue
		}

		buf := bytes.NewBuffer(nil)
		parserCtx := parser.NewContext()
		err = md.Convert(b, buf, parser.WithContext(parserCtx))
		if err != nil {
			log.Printf("error parsing markdown '%s': %s\n", recipieFileName, err.Error())
			continue
		}

		d := frontmatter.Get(parserCtx)
		recipie := models.Recipe{}
		err = d.Decode(&recipie)
		if err != nil {
			log.Printf("error parsing frontmatter '%s': %s\n", recipie, err.Error())
			continue
		}

		recipie.Contents = buf.String()
		recipies[recipieFileName] = recipie
	}

	http.Handle("/assets/", http.FileServer(http.FS(assetsFS)))
	http.Handle("/recipie/{name}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" || !strings.HasSuffix(name, ".md") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		recipie, ok := recipies[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		templ.Handler(components.Page(recipie)).ServeHTTP(w, r)

	}))

	recipie_files = []string{}
	recipies_list := []models.Recipe{}
	for file, recipie := range recipies {
		recipie_files = append(recipie_files, file)
		recipies_list = append(recipies_list, recipie)
	}

	http.Handle("/", templ.Handler(components.LandingPage(recipie_files, recipies_list)))

	log.Println(http.ListenAndServe(":8080", nil))
}
