package main

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
)

func GetRoutes() map[string]string {

	routeMap := make(map[string]string)

	{ // Load file fileRoutes...
		var fileRoutes []string
		readFile, err := os.Open(*configFile)
		if err != nil {
			// log.Println(err)
		} else {
			defer readFile.Close()
			fileScanner := bufio.NewScanner(readFile)
			fileScanner.Split(bufio.ScanLines)
			for fileScanner.Scan() {
				text := fileScanner.Text()
				fileRoutes = append(fileRoutes, text)
			}
		}
		updateRoutes(routeMap, fileRoutes)
	}

	{ // Load environment envRoutes...
		var envRoutes []string
		pattern := regexp.MustCompile(`^WEBRUN_ROUTE_\d+$`)
		for _, e := range os.Environ() {
			keyvalue := strings.SplitN(e, "=", 2)
			key, value := keyvalue[0], keyvalue[1]
			if pattern.MatchString(key) {
				envRoutes = append(envRoutes, value)
			}
		}
		updateRoutes(routeMap, envRoutes)
	}

	return routeMap
}

func updateRoutes(routeMap map[string]string, routes []string) {
	sort.Strings(routes)
	for _, route := range routes {
		parts := strings.Split(route, " ")
		path, command := parts[0], strings.Join(parts[1:], " ")
		routeMap[path] = command
	}
}
