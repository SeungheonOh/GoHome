package main

import (
	"bufio"
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
	EntrieGroups []EntrieGroup
	Labels       []Label
}

var (
	CSSFile      string
	HTMLTemplate string
	EntrieFile   string
)

func init() {
	log.SetFlags(log.Ltime | log.Llongfile)

	flag.StringVar(&CSSFile, "c", "", "Set CSS")
	flag.StringVar(&CSSFile, "css", "./templates/default.css", "Set CSS")

	flag.StringVar(&HTMLTemplate, "t", "", "Set Tamplate")
	flag.StringVar(&HTMLTemplate, "tamplate", "./templates/default.html", "Set Tamplate")

	flag.StringVar(&EntrieFile, "e", "", "Set Tamplate")
	flag.StringVar(&EntrieFile, "entrie", "./templates/entrie.txt", "Set Entries")
}

func main() {
	css, err := ioutil.ReadFile(CSSFile)
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
		}
	}

	tmpl, err := template.New("default.html").ParseFiles("default.html")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		tmpl.Execute(w, HomePage)
		log.Printf("requested to home")
	})
	http.HandleFunc("/run/", func(w http.ResponseWriter, r *http.Request) {
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
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
