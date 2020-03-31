package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode/utf8"
)

type customHandler struct {
	Scenarios map[string]scenario
}

type option struct {
	Text string `json:"text"`
	Arc  string `json:"arc"`
}

type scenario struct {
	Title   string   `json:"title"`
	Story   []string `json: "story"`
	Options []option `json:"options"`
}

func createHtmlResponseForScenario(scene scenario) string {
	tmpl, err := template.ParseFiles("./base.html")
	if err != nil {
		panic(err)
	}
	var textTemplate bytes.Buffer
	tmpl.Execute(&textTemplate, scene)
	return textTemplate.String()
}

func retrieveScenarioFromMapOfScenarios(scenarioTitle string, mapOfScenarios map[string]scenario) (scenario, error) {
	var err error
	scene, foundScenario := mapOfScenarios[scenarioTitle]
	if !foundScenario {
		err = errors.New(fmt.Sprintf("Unable to find a scenario of title \"%s\"", scenarioTitle))
	}
	return scene, err
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func (h customHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	// Path always contains a leading / , even when request is made without any
	var sceneTitle = trimFirstRune(req.URL.Path)
	sceneTitle = strings.ToLower(sceneTitle)
	scene, errFindingScene := retrieveScenarioFromMapOfScenarios(sceneTitle, h.Scenarios)
	var textResponse string
	if errFindingScene == nil {
		textResponse = createHtmlResponseForScenario(scene)
	} else {
		textResponse = "<p>You are doing really well.</p>"
	}
	response := []byte(textResponse)
	w.Write(response)
}

func scenarioKeyNeedsToBeLowerCased(originalKey string, lowerCasedKey string) bool {
	return lowerCasedKey != originalKey
}

func lowerCaseScenarioKeys(m map[string]scenario) {
	for key, value := range m {
		lowerCasedKey := strings.ToLower(key)
		if scenarioKeyNeedsToBeLowerCased(key, lowerCasedKey) {
			m[strings.ToLower(key)] = value
			delete(m, key)
		}
	}
}

func parseScenariosFromFile() map[string]scenario {
	jsonStories, err := ioutil.ReadFile("./stories.json")
	if err != nil {
		panic(err)
	}
	var scenarios map[string]scenario
	unmarshallingError := json.Unmarshal(jsonStories, &scenarios)
	if unmarshallingError != nil {
		panic(unmarshallingError)
	}
	lowerCaseScenarioKeys(scenarios)
	return scenarios
}

func main() {
	scenarios := parseScenariosFromFile()
	fmt.Println(scenarios)

	mux := http.NewServeMux()
	defaultHandler := customHandler{
		Scenarios: scenarios,
	}
	mux.Handle("/", defaultHandler)

	fmt.Println("Launching server on port 3645")
	err := http.ListenAndServe(":3645", mux)
	if err != nil {
		panic(err)
	}
}
