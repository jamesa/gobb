package utils

import (
	"fmt"
	"go/build"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/russross/blackfriday"
	"github.com/stevenleeg/gobb/config"
	"github.com/stevenleeg/gobb/models"
)

// ListTemplates returns a list of all available themes
func ListTemplates() []string {
	names := []string{"default"}

	staticPath, _ := config.Config.GetString("gobb", "base_path")
	files, _ := ioutil.ReadDir(path.Join(staticPath, "templates"))

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		names = append(names, f.Name())
	}

	fmt.Println(names)
	return names
}

func tplAdd(first, second int) int {
	return first + second
}

func tplParseMarkdown(input string) template.HTML {
	byte_slice := []byte(input)
	return template.HTML(string(blackfriday.MarkdownBasic(byte_slice)))
}

func tplGetCurrentUser(r *http.Request) func() *models.User {
	return func() *models.User {
		return GetCurrentUser(r)
	}
}

func tplGetStringSetting(key string) string {
	val, _ := models.GetStringSetting(key)
	return val
}

func tplIsValidTime(in time.Time) bool {
	return in.Year() > 1
}

func tplParseFaviconType(url string) string {
	split := strings.Split(url, ".")
	if len(split) == 0 {
		return ""
	}
	return split[len(split)-1]
}

var defaultFuncmap = template.FuncMap{
	"TimeRelativeToNow": TimeRelativeToNow,
	"Add":               tplAdd,
	"ParseMarkdown":     tplParseMarkdown,
	"IsValidTime":       tplIsValidTime,
	"GetStringSetting":  tplGetStringSetting,
	"ParseFaviconType":  tplParseFaviconType,
}

func RenderTemplate(
	out http.ResponseWriter,
	r *http.Request,
	tplFile string,
	context map[string]interface{},
	funcs template.FuncMap) {

	currentUser := GetCurrentUser(r)
	siteName, _ := config.Config.GetString("gobb", "site_name")
	baseURL, _ := config.Config.GetString("gobb", "base_url")
	gaTrackingID, _ := config.Config.GetString("googleanalytics", "tracking_id")
	gaAccount, _ := config.Config.GetString("googleanalytics", "account")

	stylesheet := ""
	if (currentUser != nil) && currentUser.StylesheetUrl.Valid && currentUser.StylesheetUrl.String != "" {
		stylesheet = currentUser.StylesheetUrl.String
	} else if currentUser == nil || !currentUser.StylesheetUrl.Valid || currentUser.StylesheetUrl.String == "" {
		globalTheme, _ := models.GetStringSetting("theme_stylesheet")
		if globalTheme != "" {
			stylesheet = globalTheme
		}
	}

	faviconURL, _ := models.GetStringSetting("favicon_url")

	send := map[string]interface{}{
		"currentUser":    currentUser,
		"request":        r,
		"site_name":      siteName,
		"ga_tracking_id": gaTrackingID,
		"ga_account":     gaAccount,
		"stylesheet":     stylesheet,
		"favicon_url":    faviconURL,
		"base_url":       baseURL,
	}

	// Merge the global template variables with the local context
	for key, val := range context {
		send[key] = val
	}

	// Same with the function map
	funcMap := defaultFuncmap
	funcMap["GetCurrentUser"] = tplGetCurrentUser(r)
	for key, val := range funcs {
		funcMap[key] = val
	}

	// Get the base template path
	selected_template, _ := models.GetStringSetting("template")
	var base_path string
	if selected_template == "default" {
		pkg, _ := build.Import("github.com/stevenleeg/gobb/gobb", ".", build.FindOnly)
		base_path = filepath.Join(pkg.SrcRoot, pkg.ImportPath, "../templates/")
	} else {
		base_path, _ = config.Config.GetString("gobb", "base_path")
		base_path = filepath.Join(base_path, "templates", selected_template)
	}

	baseTpl := filepath.Join(base_path, "base.html")
	rendTpl := filepath.Join(base_path, tplFile)

	tpl, err := template.New("tpl").Funcs(funcMap).ParseFiles(baseTpl, rendTpl)
	if err != nil {
		fmt.Printf("[error] Could not parse template (%s)\n", err.Error())
	}

	// Attempt to execute the template we're on
	err = tpl.ExecuteTemplate(out, tplFile, send)
	if err != nil {
		fmt.Printf("[error] Could not parse template (%s)\n", err.Error())
	}

	// And now the base template
	err = tpl.ExecuteTemplate(out, "base.html", send)
	if err != nil {
		fmt.Printf("[error] Could not parse template (%s)\n", err.Error())
	}
}
