package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
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

	postTmplFilename  = "post.tmpl.html"
	indexTmplFilename = "index.tmpl.html"
	feedTmplFilename  = "index.tmpl.xml"

	settingsFilename = "_index.md"
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
	Posts     []*Post
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

func parseFrontmatter(body *[]byte) map[string]interface{} {
	var frontmatterBuf bytes.Buffer
	frontmatter := make(map[string]interface{})

	buf := bytes.NewBuffer(*body)
	started, ended := false, false
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			log.Fatalln("Could not parse frontmatter:", err)
		}

		if line == "---\n" {
			if !started {
				started = true
			} else if !ended {
				ended = true
			}
		}
		if started {
			frontmatterBuf.Write([]byte(line))
		}
		if ended {
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

func sourceFiles(sourcePath string) (filenames []string, err error) {
	filenames, err = filepath.Glob(path.Join(sourcePath, "*.md"))
	return
}

func buildAll(templatesPath, outputPath string, sourcePath string) {
	log.SetFlags(log.LstdFlags)
	tmpl := template.Must(template.ParseFiles(
		path.Join(templatesPath, postTmplFilename),
		path.Join(templatesPath, indexTmplFilename),
		path.Join(templatesPath, feedTmplFilename),
	))

	var outfile *os.File
	var err error
	var body []byte

	files, err := sourceFiles(sourcePath)
	if err != nil {
		log.Fatal("ioutil.ReadFile:", err)
	}

	indexFilename := path.Join(sourcePath, settingsFilename)
	indexBody, err := ioutil.ReadFile(indexFilename)
	if err != nil {
		log.Fatalf("error reading %q in source: %v", settingsFilename, err)
	}

	indexFrontmatter := parseFrontmatter(&indexBody)

	index := Index{
		Title:     indexFrontmatter["title"].(string),
		URL:       indexFrontmatter["url"].(string),
		XMLURL:    indexFrontmatter["xmlurl"].(string),
		UpdatedAt: time.Now(),
	}

	for _, filename := range files {
		// skip the settings file
		if filepath.Base(filename) == settingsFilename {
			continue
		}

		body, err = ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal("ioutil.ReadFile:", err)
		}

		log.Println(filename)
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

		outfile, err = os.Create(path.Join(outputPath, outputFilename(filename, ".html")))
		if err != nil {
			log.Fatalln("os.Create:", err)
		}

		var descBuf, titleBuf bytes.Buffer
		xml.EscapeText(&descBuf, bytes.Trim(body[:200], " \n\r"))
		xml.EscapeText(&titleBuf, []byte(title))

		index.Posts = append(index.Posts, &Post{
			Body:         string(blackfriday.MarkdownOptions(body, renderer, blackfriday.Options{Extensions: commonExtensions})),
			Date:         date,
			Link:         index.URL + outputFilename(filename, ".html"),
			RelativeLink: "/" + outputFilename(filename, ".html"),
			Title:        title,
			XMLDesc:      descBuf.String(),
			XMLTitle:     titleBuf.String(),
			Draft:        draft,
		})

		err = tmpl.ExecuteTemplate(outfile, postTmplFilename,
			&struct {
				*Post
				Index *Index
			}{
				Post:  index.Posts[len(index.Posts)-1],
				Index: &index,
			})
		if err != nil {
			log.Fatalln("tmpl.ExecuteTemplate:", err)
		}
	}

	sort.Sort(sort.Reverse(index))

	// index.html
	if outfile, err = os.Create(path.Join(outputPath, "index.html")); err != nil {
		log.Fatalln("os.Create:", err)
	}
	if err := tmpl.ExecuteTemplate(outfile, indexTmplFilename, index); err != nil {
		log.Fatalln("tmpl.ExecuteTemplate:", err)
	}
	log.Println("index.html")

	// index.xml
	if outfile, err = os.Create(path.Join(outputPath, "index.xml")); err != nil {
		log.Fatalln("os.Create:", err)
	}
	if err := tmpl.ExecuteTemplate(outfile, feedTmplFilename, index); err != nil {
		log.Fatalln("tmpl.ExecuteTemplate:", err)
	}
	log.Println("index.xml")
}

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] sources\n", os.Args[0])
		flag.PrintDefaults()
	}

	watchFlag := flag.Bool("watch", false, "tries to rebuild the src on change")
	serveFlag := flag.String("serve", "", "listening address for serving the blog")
	outPathFlag := flag.String("output", "generated", "output path")
	assetsFlag := flag.String("assets", "", "path to the assets files for serving")
	templatesFlag := flag.String("templates", "", "path to the templates directory")

	flag.Parse()

	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(1)
	}

	cwd, _ := os.Getwd()

	// check output path
	if stat, err := os.Stat(path.Join(cwd, *outPathFlag)); err != nil && !os.IsExist(err) || !stat.IsDir() {
		err := os.Mkdir(*outPathFlag, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "specified path \"%s\" for output couldn't be created: %s\n", *outPathFlag, err)
			os.Exit(1)
		}
	}

	// check post in output path
	postPath := path.Join(cwd, *outPathFlag, "post")
	if stat, err := os.Stat(postPath); err != nil && !os.IsExist(err) || !stat.IsDir() {
		err := os.Mkdir(postPath, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "path \"%s\" couldn't be created: %s\n", postPath, err)
			os.Exit(1)
		}
	}

	// check assets path
	if assetsFlag != nil {
		// check out assetPath
		if stat, err := os.Stat(path.Join(cwd, *assetsFlag)); err != nil && !os.IsExist(err) || !stat.IsDir() {
			fmt.Fprintf(os.Stderr, "specified path \"%s\" for assets doesn't exists or is not a directory\n", *assetsFlag)
			os.Exit(1)
		}
	}

	sourcePath := flag.Arg(0)
	buildAll(*templatesFlag, *outPathFlag, sourcePath)

	if *watchFlag {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		files, err := sourceFiles(sourcePath)
		if err != nil {
			log.Fatal("ioutil.ReadFile:", err)
		}
		for _, filename := range files {
			log.Println("adding", filename)
			if err := watcher.Add(filename); err != nil {
				log.Fatal(err)
			}
		}
		for _, filename := range []string{indexTmplFilename, feedTmplFilename, postTmplFilename} {
			if err := watcher.Add(path.Join(*templatesFlag, filename)); err != nil {
				log.Fatal(err)
			}
		}

		go func() {
			for {
				select {
				case event := <-watcher.Events:
					log.Println(event.Op, event.Name)
					if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Write == fsnotify.Write {
						buildAll(*templatesFlag, *outPathFlag, sourcePath)
						watcher.Add(event.Name)
					}
				case err := <-watcher.Errors:
					log.Println(err)
				}
			}
		}()
	}

	if serveFlag != nil && *serveFlag != "" {
		if assetsFlag != nil && *assetsFlag != "" {
			fs := http.FileServer(http.Dir(*assetsFlag))
			http.Handle("/assets/", http.StripPrefix("/assets/", fs))
		}

		fs := http.FileServer(http.Dir(*outPathFlag))
		http.Handle("/", fs)

		fmt.Fprintf(os.Stderr, "Listening on http://%s\n", *serveFlag)
		if err := http.ListenAndServe(*serveFlag, nil); err != nil {
			panic(err)
		}
	} else if *watchFlag {
		// blocking for watch
		done := make(chan bool)
		<-done
	}
}
