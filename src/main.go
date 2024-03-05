package main

import (
	"bufio"
	"flag"
	"fmt"
	"josuedlt/webrun/basicLogger"
	"net/http"
	"os/exec"
	"sort"
	"strings"
)

const (
	VERSION = "0.1"
)

var (
	configFile *string
	god        *bool
	logFile    *string
	menuPath   *string
	port       *int
	routeMap   map[string]string
	silent     *bool
)

func init() {
	configFile = flag.String("config", "webrun.config", "path to config file")
	god = flag.Bool("god", false, "enable god mode (default false)")
	logFile = flag.String("log", "", "path to log file (default blank)")
	menuPath = flag.String("menu", "/menu", "path to help menu")
	port = flag.Int("port", 8080, "port number")
	silent = flag.Bool("silent", false, "silent mode (default false)")

	flag.Parse()
}

func main() {
	routeMap = GetRoutes()
	logger := basicLogger.CreateLogger(*silent, *logFile)
	logger.Println("\n📧 Written by Josue de la Torre (josue@jdlt.com)")
	logger.Println("🔀", len(routeMap), "Routes loaded:", routeMap)
	if *god {
		logger.Println("❗ God mode: enabled")
	}

	{ // Start the web server
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			{ // Ignore favicon
				if r.URL.Path == "/favicon.ico" {
					return
				}
			}

			{ // Match to routes
				if command := routeMap[r.URL.Path]; command != "" {
					logger.Println(r.URL.Path, "-->", strings.Split(command, " "))
					CommandHandler(w, r, command)
					return
				}
			}

			{ // Reload routes...
				if r.URL.Path == "/reload" {
					logger.Println(r.URL.Path, "~~>", "reloading routes...")
					routeMap = GetRoutes()
					http.Redirect(w, r, *menuPath, http.StatusSeeOther)
					return
				}
			}

			{ // Display route menu...
				if r.URL.Path == *menuPath {
					logger.Println(r.URL.Path, "~~>", "showing help menu...")
					BuildHelpMenu(w, r, routeMap)
					return
				}
			}

			{ // God mode...
				if *god {
					if command := r.URL.Path[1:]; command != "" {
						logger.Println(r.URL.Path, "==>", strings.Split(command, " "))
						CommandHandler(w, r, command)
						return
					}
				}
			}

			logger.Println(r.URL.Path, "~~>", *menuPath)
			http.Redirect(w, r, *menuPath, http.StatusSeeOther)
		})
		logger.Printf("🌐 Server started on port %v", *port)
		logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), nil))
	}
}

func CommandHandler(w http.ResponseWriter, r *http.Request, command string) {
	parts := strings.Split(command, " ")
	process := exec.Command(parts[0], parts[1:]...)
	stdout, _ := process.StdoutPipe()
	stderr, _ := process.StderrPipe()

	if err := process.Start(); err != nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, err)
		w.(http.Flusher).Flush()
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	stdOutScanner := bufio.NewScanner(stdout)
	for stdOutScanner.Scan() {
		fmt.Fprintln(w, stdOutScanner.Text())
		w.(http.Flusher).Flush()
	}
	stdErrScanner := bufio.NewScanner(stderr)
	for stdErrScanner.Scan() {
		fmt.Fprintln(w, stdErrScanner.Text())
		w.(http.Flusher).Flush()
	}
}

func BuildHelpMenu(w http.ResponseWriter, r *http.Request, routeMap map[string]string) {

	fmt.Fprintln(w, "<p><a href='/reload'><button>Reload routes</button></a></p>")
	if len(routeMap) == 0 {
		fmt.Fprintln(w, "<code>No routes loaded</code>")
		return
	}

	items := []string{}
	for k, v := range routeMap {
		items = append(items, fmt.Sprintf("<div><code><a href='%s'>%s</a> --> %s</code></div>", k, k, v))
	}
	sort.Strings(items)

	fmt.Fprintln(w, strings.Join(items, "\n"))
}
