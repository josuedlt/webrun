package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	VERSION = "0.2"
)

var (
	configFile = flag.String("config", "webrun.config", "path to config file")
	god        = flag.Bool("god", false, "enable god mode (default false)")
	logFile    = flag.String("log", "", "path to log file (default blank)")
	menuPath   = flag.String("menu", "/menu", "path to help menu")
	port       = flag.Int("port", 8080, "port number")
	silent     = flag.Bool("silent", false, "silent mode (default false)")
	showErrors = flag.Bool("showErrors", false, "show stderr output (default false)")
	routeMap   = make(map[string]string)
)

func init() {
	flag.Parse()
	{ // Override with environment variables
		if v := os.Getenv("WEBRUN_CONFIG"); v != "" {
			*configFile = v
		}
		if v := os.Getenv("WEBRUN_GOD"); v != "" {
			b, err := strconv.ParseBool(v)
			if err == nil {
				*god = b
			}
		}
		if v := os.Getenv("WEBRUN_LOGFILE"); v != "" {
			*logFile = v
		}
		if v := os.Getenv("WEBRUN_MENUPATH"); v != "" {
			*menuPath = v
		}
		if v := os.Getenv("WEBRUN_PORT"); v != "" {
			p, err := strconv.Atoi(v)
			if err == nil {
				*port = p
			}
		}
		if v := os.Getenv("WEBRUN_SILENT"); v != "" {
			b, err := strconv.ParseBool(v)
			if err == nil {
				*silent = b
			}
		}
		if v := os.Getenv("WEBRUN_SHOWERRORS"); v != "" {
			b, err := strconv.ParseBool(v)
			if err == nil {
				*showErrors = b
			}
		}
	}

	routeMap = LoadRoutes()
	{ // Override with command line args
		if len(flag.Args()) > 0 {
			routeMap["/"] = strings.Join(flag.Args(), " ")
		}
	}
}

func main() {
	logger := CreateLogger(*silent, *logFile)
	logger.Println("📧 Written by Josue de la Torre (josue@jdlt.com)")
	logger.Println("🔀", len(routeMap), "Routes loaded:", routeMap)
	if *god {
		logger.Println("❗ Warning: God mode enabled")
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
					routeMap = LoadRoutes()
					http.Redirect(w, r, *menuPath, http.StatusSeeOther)
					return
				}
			}

			{ // Display route menu...
				if r.URL.Path == *menuPath {
					logger.Println(r.URL.Path, "~~>", "showing help menu...")
					HelpMenuHandler(w, r, routeMap)
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

		hostname, _ := os.Hostname()
		logger.Printf("🌐 Server started at http://%s:%v", hostname, *port)
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
	stdOutScanner := bufio.NewReader(stdout)
	for {
		char, err := stdOutScanner.ReadByte()
		if err != nil {
			break
		}
		fmt.Fprint(w, string(char))
		w.(http.Flusher).Flush()
	}

	if *showErrors || *god {
		stdErrScanner := bufio.NewScanner(stderr)
		for stdErrScanner.Scan() {
			fmt.Fprintln(w, stdErrScanner.Text())
			w.(http.Flusher).Flush()
		}
	}
}

func HelpMenuHandler(w http.ResponseWriter, r *http.Request, routeMap map[string]string) {

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

func CreateLogger(silent bool, logFile string) *log.Logger {
	output := []io.Writer{}
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			output = append(output, file)
		}
	}
	if !silent {
		output = append(output, os.Stderr)
	}
	return log.New(io.MultiWriter(output...), "", log.LstdFlags)
}

func LoadRoutes() map[string]string {

	routeMap := make(map[string]string)

	{ // Load file fileRoutes...
		var routes []string
		readFile, err := os.Open(*configFile)
		if err != nil {
			// log.Println(err)
		} else {
			defer readFile.Close()
			fileScanner := bufio.NewScanner(readFile)
			fileScanner.Split(bufio.ScanLines)
			for fileScanner.Scan() {
				text := fileScanner.Text()
				if text != "" {
					routes = append(routes, text)
				}
			}
		}
		updateRouteMap(routeMap, routes)
	}

	{ // Override with environment variables...
		var routes []string
		pattern := regexp.MustCompile(`^WEBRUN_ROUTE_\d+$`)
		for _, e := range os.Environ() {
			keyvalue := strings.SplitN(e, "=", 2)
			key, value := keyvalue[0], keyvalue[1]
			if pattern.MatchString(key) {
				routes = append(routes, value)
			}
		}
		updateRouteMap(routeMap, routes)
	}

	return routeMap
}

func updateRouteMap(routeMap map[string]string, routes []string) {
	sort.Strings(routes)
	for _, route := range routes {
		parts := strings.Split(route, " ")
		path, command := parts[0], strings.Join(parts[1:], " ")
		routeMap[path] = command
	}
}
