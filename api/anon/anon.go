package anon

import (
	"bytes"
	"github.com/alice-ws/alice/resources"
	"html/template"
	"log"
	"math/rand"
)

type generatedName struct {
	Name  string
	Title string
}

func GenerateUsername(seed int64) string {
	files, e := template.New("name generator").Parse(resources.GetNameGenTemplate())
	if e != nil {
		log.Fatal(e)
	}

	names := resources.GetNames()
	titles := resources.GetTitles()

	rand.Seed(seed)
	generatedName := generatedName{
		Name:  names[rand.Intn(len(names))],
		Title: titles[rand.Intn(len(titles))],
	}

	buf := new(bytes.Buffer)
	e = files.Execute(buf, generatedName)

	return buf.String()
}

func Defaults() (string, string) {
	return "anon@example.com", "password"
}
