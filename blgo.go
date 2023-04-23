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
	"github.com/russross/blackfriday"
	yaml "gopkg.in/yaml.v2"
)

const (
	shortTimeFormat = "2006-01-02"

	postTmplFilename  = "post.tmpl.html"
	indexTmplFilename = "index.tmpl.html"
	feedTmplFilename  = "index.tmpl.xml"

	settingsFilename = "_index.md"
)

// Post represents a single blog post
type Post struct {
	Index          *Index
	Slug           string
	OutputFilename string
	Body           string
	Date           time.Time
	Description    string
	GUID           string
	Link           string
	RelativeLink   string
	Title          string
	XMLDesc        string
	XMLTitle       string
	Draft          bool
	Short          bool
}

// ReadFile will fill the post from given filename
func (p *Post) ReadFile(filename string) error {
	body, err := ioutil.ReadFile(filename)
	if err == nil {
		return p.Read(filename, body)
	}
	return err
}

// Read will fill the post from given byte string
func (p *Post) Read(filename string, body []byte) error {
	var title string
	var draft bool
	var short bool
	var date time.Time
	var err error

	frontmatter, err := parseFrontmatter(&body)
	if err != nil {
		return err
	}

	if v, ok := frontmatter["title"]; ok {
		title = v.(string)
	} else {
		return fmt.Errorf("could not read the title from post")
	}

	if v, ok := frontmatter["draft"]; ok {
		draft = v.(bool)
	}

	if v, ok := frontmatter["short"]; ok {
		short = v.(bool)
	}

	if v, ok := frontmatter["date"]; ok {
		if date, err = time.Parse(shortTimeFormat, v.(string)); err != nil {
			return err
		}
	}

	var descBuf, titleBuf bytes.Buffer
	var desc = body
	if len(desc) > 200 {
		desc = desc[:200]
	}
	xml.EscapeText(&descBuf, bytes.Trim(desc, " \n\r"))
	xml.EscapeText(&titleBuf, []byte(title))

	p.Slug = strings.TrimSuffix(filepath.Base(filename), ".md")
	p.OutputFilename = path.Join("post", p.Slug+".html")
	p.Body = string(blackfriday.MarkdownOptions(body, renderer, blackfriday.Options{Extensions: commonExtensions}))
	p.Title = title
	p.Date = date
	p.Link = strings.TrimSuffix(p.Index.URL, "/") + "/" + path.Join("post", p.Slug)
	p.RelativeLink = path.Join("post", p.Slug)
	p.XMLDesc = descBuf.String()
	p.XMLTitle = titleBuf.String()
	p.Draft = draft
	p.Short = short

	return nil
}

// Index represents global settings/variables and the index of the posts
// the index.html will generated from Index
type Index struct {
	Title     string
	Posts     []*Post
	URL       string
	XMLURL    string
	UpdatedAt time.Time
}

func (index *Index) Len() int           { return len(index.Posts) }
func (index *Index) Swap(i, j int)      { index.Posts[i], index.Posts[j] = index.Posts[j], index.Posts[i] }
func (index *Index) Less(i, j int) bool { return index.Posts[i].Date.Before(index.Posts[j].Date) }

// ReadFrontmatterFile will fill the index frontmatter from given filename
func (index *Index) ReadFrontmatterFile(filename string) error {
	body, err := ioutil.ReadFile(filename)
	if err == nil {
		return index.ReadFrontmatter(body)
	}
	return err
}

// ReadFrontmatter will fill the index frontmatter from given data
func (index *Index) ReadFrontmatter(body []byte) error {
	indexFrontmatter, err := parseFrontmatter(&body)
	if err != nil {
		return err
	}

	index.Title = indexFrontmatter["title"].(string)
	index.URL = indexFrontmatter["url"].(string)
	index.XMLURL = indexFrontmatter["xmlurl"].(string)
	index.UpdatedAt = time.Now()
	return nil
}

func parseFrontmatter(body *[]byte) (map[string]interface{}, error) {
	var frontmatterBuf bytes.Buffer
	buf := bytes.NewBuffer(*body)
	started := false
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if line == "---\n" {
			if started {
				break
			}
			started = true
		}
		if started {
			frontmatterBuf.Write([]byte(line))
		}
	}

	*body = buf.Bytes() // rest of the bytes
	frontmatter := make(map[string]interface{})
	return frontmatter, yaml.Unmarshal(frontmatterBuf.Bytes(), &frontmatter)
}

// listSourceFiles lists files that has ".md" extension in specified path
func listSourceFiles(sourcePath string) (filenames []string, err error) {
	filenames, err = filepath.Glob(path.Join(sourcePath, "*.md"))
	return
}

// buildAll builds the whole blog
func buildAll(seedPath string, outputPath string) {
	templatesPath := filepath.Join(seedPath, "templates")
	sourcePath := filepath.Join(seedPath, "src")

	log.SetFlags(log.LstdFlags)
	tmpl := template.Must(template.ParseFiles(
		path.Join(templatesPath, postTmplFilename),
		path.Join(templatesPath, indexTmplFilename),
		path.Join(templatesPath, feedTmplFilename),
	))

	files, err := listSourceFiles(sourcePath)
	if err != nil {
		log.Fatal("ioutil.ReadFile:", err)
	}

	indexFilename := path.Join(sourcePath, settingsFilename)
	index := &Index{}
	if err := index.ReadFrontmatterFile(indexFilename); err != nil {
		log.Fatalf("error in reading frontmatter of %q: %v", settingsFilename, err)
	}

	var outfile *os.File

	for _, filename := range files {
		// skip the settings file
		if filepath.Base(filename) == settingsFilename {
			continue
		}
		post := &Post{Index: index}
		if err := post.ReadFile(filename); err != nil {
			log.Fatalln("post.ReadFile:", err)
		}
		index.Posts = append(index.Posts, post)

		if outfile, err = os.Create(path.Join(outputPath, post.OutputFilename)); err != nil {
			log.Fatalln("os.Create:", err)
		}
		if tmpl.ExecuteTemplate(outfile, postTmplFilename, post); err != nil {
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

	// index.xml
	if outfile, err = os.Create(path.Join(outputPath, "index.xml")); err != nil {
		log.Fatalln("os.Create:", err)
	}
	if err := tmpl.ExecuteTemplate(outfile, feedTmplFilename, index); err != nil {
		log.Fatalln("tmpl.ExecuteTemplate:", err)
	}
}

type fileServer struct {
	h          http.Handler
	suffix     string
	defaultExt string
}

func (n *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.String())
	if strings.HasSuffix(r.URL.Path, n.suffix) {
		http.NotFound(w, r)
		return
	}
	if n.defaultExt != "" {
		if r.URL.Path != "/" && !strings.HasSuffix(r.URL.Path, n.defaultExt) {
			r.URL.Path = r.URL.Path + n.defaultExt
		}
	}
	n.h.ServeHTTP(w, r)
}

// FileServer will return 404 when requested url ends with suffix
func FileServer(suffix string, defaultExt string, h http.Handler) http.Handler {
	return &fileServer{suffix: suffix, defaultExt: defaultExt, h: h}
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
	seedFlag := flag.String("seed", "", "path to seed directory")

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

	// check assets in output path
	assetsPath := path.Join(cwd, *outPathFlag, "assets")
	if stat, err := os.Stat(assetsPath); err != nil && !os.IsExist(err) || !stat.IsDir() {
		err := os.Mkdir(assetsPath, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "specified path \"%s\" for assets doesn't exists or is not a directory\n", assetsPath)
			os.Exit(1)
		}
	}

	buildAll(*seedFlag, *outPathFlag)

	if *watchFlag {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		sourcePath := filepath.Join(*seedFlag, "src")
		files, err := listSourceFiles(sourcePath)
		if err != nil {
			log.Fatal("ioutil.ReadFile:", err)
		}
		for _, filename := range files {
			log.Println("adding", filename)
			if err := watcher.Add(filename); err != nil {
				log.Fatal(err)
			}
		}

		templatesFlag := filepath.Join(*seedFlag, "templates")
		for _, filename := range []string{indexTmplFilename, feedTmplFilename, postTmplFilename} {
			if err := watcher.Add(path.Join(templatesFlag, filename)); err != nil {
				log.Fatal(err)
			}
		}

		go func() {
			for {
				select {
				case event := <-watcher.Events:
					log.Println(event)
					if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Write == fsnotify.Write {
						buildAll(*seedFlag, *outPathFlag)
						watcher.Add(event.Name)
					}
				case err := <-watcher.Errors:
					log.Println(err)
				}
			}
		}()
	}

	if serveFlag != nil && *serveFlag != "" {
		fs := FileServer("/", "", http.FileServer(http.Dir(assetsPath)))
		http.Handle("/assets/", http.StripPrefix("/assets", fs))

		fs = FileServer("/post/", ".html", http.FileServer(http.Dir(*outPathFlag)))
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
