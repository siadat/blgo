package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	yaml "gopkg.in/yaml.v2"

	"github.com/russross/blackfriday"
)

const (
	shortTimeFormat = "2006-01-02"
	longTimeFormat  = "January 02, 2006"
)

type Post struct {
	Body         string
	Date         time.Time
	Description  string
	GUID         string
	Link         string
	RelativeLink string
	Title        string
	BlogTitle    string
	XMLDesc      string
	XMLTitle     string
	Draft        bool
}

type Index struct {
	Title     string
	Posts     []Post
	URL       string
	XMLURL    string
	UpdatedAt time.Time
}

func (index Index) Len() int           { return len(index.Posts) }
func (index Index) Swap(i, j int)      { index.Posts[i], index.Posts[j] = index.Posts[j], index.Posts[i] }
func (index Index) Less(i, j int) bool { return index.Posts[i].Date.Before(index.Posts[j].Date) }

func outputFilename(filename, ext string) string {
	filename = strings.TrimSuffix(filepath.Base(filename), ".md") + ext
	return filepath.Join("post", filename)
}

func parseFrontmatter(body *[]byte) (frontmatter map[interface{}]interface{}) {
	var frontmatterBuf bytes.Buffer
	frontmatter = make(map[interface{}]interface{})

	buf := bytes.NewBuffer(*body)
	stared, ended := false, false
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			log.Fatalln("Could not parse frontmatter:", err)
		}

		if line == "---\n" {
			if stared == false {
				stared = true
			} else if ended == false {
				ended = true
			}
		}
		if stared != false {
			frontmatterBuf.Write([]byte(line))
		}
		if ended != false {
			break
		}
	}

	if err := yaml.Unmarshal(frontmatterBuf.Bytes(), &frontmatter); err != nil {
		log.Fatalln("yaml.Unmarshal:", err)
	}

	// rest of the bytes:
	*body = buf.Bytes()
	return frontmatter
}

func buildAll(cmd string, mdFiles []string) {
	log.SetFlags(log.LstdFlags)
	tmpl := template.Must(template.ParseFiles(
		"templates/post.tmpl.html",
		"templates/index.tmpl.html",
		"templates/index.tmpl.xml",
	))

	var outfile *os.File
	var err error
	var body []byte

	index := Index{
		Title:     "Sina Siadat",
		URL:       "https://siadat.github.io/",
		XMLURL:    "https://siadat.github.io/index.xml",
		UpdatedAt: time.Now(),
	}

	for _, mdFilename := range mdFiles {
		body, err = ioutil.ReadFile(mdFilename)
		if err != nil {
			log.Fatal("ioutil.ReadFile:", err)
		}

		log.Println(mdFilename)
		var title string
		var draft bool
		var date time.Time
		frontmatter := parseFrontmatter(&body)

		if v, ok := frontmatter["title"]; ok {
			title = v.(string)
		}

		if v, ok := frontmatter["draft"]; ok {
			draft = v.(bool)
		}
		if draft {
			continue
		}

		if v, ok := frontmatter["date"]; ok {
			if date, err = time.Parse(shortTimeFormat, v.(string)); err != nil {
				log.Println("time.Parse:", err)
			}
		}

		outfile, err = os.Create(outputFilename(mdFilename, ".html"))
		if err != nil {
			log.Fatalln("os.Create:", err)
		}

		var descBuf, titleBuf bytes.Buffer
		xml.EscapeText(&descBuf, bytes.Trim(body[:200], " \n\r"))
		xml.EscapeText(&titleBuf, []byte(title))

		index.Posts = append(index.Posts, Post{
			Body:         string(blackfriday.MarkdownOptions(body, renderer, blackfriday.Options{Extensions: commonExtensions})),
			Date:         date,
			Link:         index.URL + outputFilename(mdFilename, ""),
			RelativeLink: "/" + outputFilename(mdFilename, ""),
			Title:        title,
			BlogTitle:    "Sina Siadat",
			XMLDesc:      descBuf.String(),
			XMLTitle:     titleBuf.String(),
			Draft:        draft,
		})

		err = tmpl.ExecuteTemplate(outfile, "post.tmpl.html", index.Posts[len(index.Posts)-1])
		if err != nil {
			log.Fatalln("tmpl.ExecuteTemplate:", err)
		}
	}

	sort.Sort(sort.Reverse(index))

	// index.html
	if outfile, err = os.Create("index.html"); err != nil {
		log.Fatalln("os.Create:", err)
	}
	if err := tmpl.ExecuteTemplate(outfile, "index.tmpl.html", index); err != nil {
		log.Fatalln("tmpl.ExecuteTemplate:", err)
	}
	log.Println("index.html")

	// index.xml
	if outfile, err = os.Create("index.xml"); err != nil {
		log.Fatalln("os.Create:", err)
	}
	if err := tmpl.ExecuteTemplate(outfile, "index.tmpl.xml", index); err != nil {
		log.Fatalln("tmpl.ExecuteTemplate:", err)
	}
	log.Println("index.xml")
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	if flag.NArg() <= 1 {
		log.Print("Usage: $0 build")
		log.Print("$0 build src/*.md")
		log.Print("$0 watch src/*.md")
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	mdFiles := flag.Args()[1:]
	os.MkdirAll("post", 0777)
	log.Printf("started with cmd=%q", cmd)

	switch cmd {
	case "build":
		buildAll(cmd, mdFiles)
	case "watch":
		go buildAll(cmd, mdFiles)
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		for _, filename := range mdFiles {
			log.Println("adding", filename)
			if err := watcher.Add(filename); err != nil {
				log.Fatal(err)
			}
		}
		for _, filename := range []string{"index.tmpl.html", "index.tmpl.xml", "post.tmpl.html"} {
			if err := watcher.Add("templates/" + filename); err != nil {
				log.Fatal(err)
			}
		}

		go func() {
			for {
				select {
				case event := <-watcher.Events:
					log.Println(event.Op, event.Name)
					if event.Op&fsnotify.Remove == fsnotify.Remove {
						buildAll(cmd, mdFiles)
						watcher.Add(event.Name)
					}
				case err := <-watcher.Errors:
					log.Println(err)
				}
			}
		}()

		done := make(chan bool)
		<-done
	}
}
