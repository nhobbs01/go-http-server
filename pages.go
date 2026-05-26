package main

import "os"

func InitiPages() {
	RegisterRoute("/", handleIndexRoute)
	RegisterRoute("/about", handleAboutRoute)
}

func handleIndexRoute(req HTTPRequest) string {
	fileBytes, err := os.ReadFile("app/index.html")
	if err != nil {
		panic(err)
	}
	name := req.queryParams["name"]
	return replaceKeyInString(string(fileBytes), "name", name)
}

func handleAboutRoute(req HTTPRequest) string {
	fileBytes, err := os.ReadFile("app/about.html")
	if err != nil {
		panic(err)
	}
	return string(fileBytes)
}
