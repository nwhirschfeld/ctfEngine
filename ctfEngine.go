package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/russross/blackfriday"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func handleError(c *fiber.Ctx, err error) error {
	fmt.Printf("[ERROR] an error occurred: %s!\n", err)
	return c.Redirect("/")
}

//go:embed views/*
var viewsFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	var ctfLocation string
	var signupTokenToAdd string
	flag.StringVar(&ctfLocation, "l", "/ctf", "load ctf from this directory")
	flag.StringVar(&signupTokenToAdd, "a", "", "add a token for signup to the database")
	flag.Parse()
	ctf, err := initCTF(ctfLocation)
	if err != nil {
		fmt.Printf("Could not load CTF from path \"%s\".\n", ctfLocation)
		return
	}

	engine := html.NewFileSystem(http.FS(viewsFS), ".html")

	engine.AddFunc(
		"renderMarkdown", func(s string) template.HTML {
			return template.HTML(blackfriday.MarkdownCommon([]byte(s)))
		},
	)
	engine.AddFunc(
		"ppFilesize", func(size int64) template.HTML {
			var suffixes [5]string
			suffixes[0] = "B"
			suffixes[1] = "KB"
			suffixes[2] = "MB"
			suffixes[3] = "GB"
			suffixes[4] = "TB"

			base := math.Log(float64(size)) / math.Log(1024)
			getSize := math.Ceil(10*math.Pow(1024, base-math.Floor(base))) / 10
			getSuffix := suffixes[int(math.Floor(base))]

			return template.HTML(strconv.FormatFloat(getSize, 'f', -1, 64) + " " + getSuffix)
		},
	)
	engine.AddFunc(
		"inList", func(element string, list []string) bool {
			for _, listElement := range list {
				if strings.Compare(element, listElement) == 0 {
					return true
				}
			}
			return false
		},
	)

	app := fiber.New(fiber.Config{
		Views:        engine,
		ServerHeader: "ctfEngine",
		AppName:      "ctfEngine",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// TODO: add logging here
			return c.Render("views/error", fiber.Map{}, "views/layouts/main")

		},
	})

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(staticFS),
		PathPrefix: "static",
		Browse:     true,
	}))

	if signupTokenToAdd != "" {
		err = ctf.addSignupToken(signupTokenToAdd)
		if err == nil {
			fmt.Println("added signup token to DB")
		} else {
			fmt.Println("could not add signup token to DB")
		}
		return
	}

	addRoutes(app, &ctf)

	log.Fatal(app.Listen(":3000"))
}
