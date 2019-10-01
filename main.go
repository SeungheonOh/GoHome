package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
	"crypto/rand"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

type Cmd struct {
	Output []byte
	Error  error
}

func RunCommand(ch chan<- Cmd, s []string) {
	output, err := exec.Command(s[0], s[1:]...).Output()
	ch <- Cmd{Output: output, Error: err}
}

type Entrie struct {
	Name    string
	Command string
}

type EntrieGroup struct {
	SubTitle string
	Entries  []Entrie
}

type Label struct {
	Prompt string
}

type HomeMenu struct {
	Title        string
	CSS          string
	CSRFToken	string
	EntrieGroups []EntrieGroup
	Labels       []Label
}

var (
	CSSFile      string
	HTMLTemplate string
	EntrieFile   string
	CSRFKey string
)

func init() {
	log.SetFlags(log.Ltime | log.Llongfile)

	flag.StringVar(&CSSFile, "c", "", "Set CSS")
	flag.StringVar(&CSSFile, "css", "./templates/default.css", "Set CSS")

	flag.StringVar(&HTMLTemplate, "t", "", "Set Tamplate")
	flag.StringVar(&HTMLTemplate, "tamplate", "./templates/default.html", "Set Tamplate")

	flag.StringVar(&EntrieFile, "e", "", "Set Tamplate")
	flag.StringVar(&EntrieFile, "entrie", "./templates/entrie.txt", "Set Entries")

	flag.StringVar(&CSRFKey, "k", "", "base64-encoded key for encrypting CSRF cookies")
	flag.StringVar(&CSRFKey, "csrfkey", "", "base64-encoded key for encrypting CSRF cookies")
}

func main() {
	css, err := ioutil.ReadFile(CSSFile)
	if err != nil {
		log.Fatal(err)
	}
	html, err := ioutil.ReadFile(HTMLTemplate)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(EntrieFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)



	HomePage := HomeMenu{
		Title: "",
		CSS:   string(css),
	}

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) <= 0 {
			continue
		}

		var group EntrieGroup
		if line[0] == '!' {
			group.SubTitle = line[1:]
			for scanner.Scan() {
				if len(scanner.Text()) <= 0 || scanner.Text()[0] == '!' {
					HomePage.EntrieGroups = append(HomePage.EntrieGroups, group)
					break
				}
				divided := strings.Fields(scanner.Text())
				group.Entries = append(group.Entries, Entrie{divided[0], strings.Join(divided[1:], " ")})
			}
		} else if line[0] == '+' {
			HomePage.Labels = append(HomePage.Labels, Label{line[1:]})
		} else if line[0] == '^' {
			if(CSRFKey != "") {
				log.Println("A CSRF key was specified both in the command line and in the entries file.")
				log.Println("Using the one from the command line.")
			} else {
				CSRFKey = line[1:]
			}
		}
	}
	csrf_key := make([]byte, 32)
	if(CSRFKey == "") {
		rand.Read(csrf_key)
		encoded_csrf_key := base64.StdEncoding.EncodeToString(csrf_key)
		log.Printf("Using newly generated CSRF key: %s", encoded_csrf_key)
		log.Println("Specify it on the command line using the -k flag or add")
		log.Printf("^%s", encoded_csrf_key)
		log.Println("to the entrie file if you want to be able to reuse it.")
	} else {
		csrf_key, err = base64.StdEncoding.DecodeString(CSRFKey)
		if(err != nil) {
			log.Println("A CSRF key was supplied on the command line but could not be decoded.")
			log.Fatal(err)
		}
		if(len(csrf_key) != 32) {
			log.Fatal("A CSRF key was specified on the command line, but it decoded to %d bytes. gorilla/csrf requires a 32 byte key.", len(csrf_key))
		}
	}
	router := mux.NewRouter()

	tmpl, err := template.New("").Parse(string(html))
	if err != nil {
		log.Fatal(err)
	}

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		HomePage.CSRFToken = csrf.Token(r)
		tmpl.Execute(w, HomePage)
		log.Printf("requested to home")
	})
	router.PathPrefix("/run/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		log.Printf("Command: %s\n", path[len("/run/"):len(path)])

		c := make(chan Cmd)

		app := strings.Fields(path[len("/run/"):len(path)])
		go RunCommand(c, app)

		var out Cmd
		select {
		case o := (<-c):
			out = o
		case <-time.After(300 * time.Millisecond):
			return
		}
		if out.Error != nil {
			log.Println(err)
		}

		log.Printf("\n%s", string(out.Output))

		fmt.Fprintf(w, string(out.Output))
	}).Methods("POST")
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(
		// only listen on the loopback so that others on the LAN can't send commands to execute
		"localhost:8080",
		// use the gorilla CSRF package to prevent links in other web pages from firing off commands
		// csrf.Secure(false) is because this isn't going over https,
		// and csrf.MaxAge(0) advises the browser to discard any cookies when it's restarted.
		csrf.Protect(csrf_key, csrf.Secure(false), csrf.MaxAge(0))(router)))
}
