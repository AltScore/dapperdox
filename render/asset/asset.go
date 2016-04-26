package asset

// This package scans a default template directory *and* an override template directory,
// building an "go generate" compliant Assets structure. Templates in under the override
// directory replace of suppliment those in the default directory.
// This allows default themes to be provided, and then changed on a per-use basis by
// dropping files in the override directory.

import (
	"bufio"
	"bytes"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/shurcooL/github_flavored_markdown"
	"github.com/zxchris/swaggerly/config"
	"github.com/zxchris/swaggerly/logger"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var _bindata = map[string][]byte{}
var _metadata = map[string]map[string]string{}
var guideReplacer *strings.Replacer

func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if a, ok := _bindata[cannonicalName]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

func MetaData(filename string, name string) string {
	if md, ok := _metadata[filename]; ok {
		if val, ok := md[strings.ToLower(name)]; ok {
			return val
		}
	}
	return ""
}

func MetaDataFileList() []string {
	files := make([]string, len(_metadata))
	ix := 0
	for key := range _metadata {
		files[ix] = key
		ix++
	}
	return files
}

func Compile(dir string, prefix string) {

	cfg, _ := config.Get()

	// Build a replacer to search/replace Document URLs in the documents.
	if guideReplacer == nil {
		var replacements []string

		// Configure the replacer with key=value pairs
		for i := range cfg.DocumentRewriteURL {

			slice := strings.Split(cfg.DocumentRewriteURL[i], "=")

			if len(slice) != 2 {
				panic("Invalid DocumentWriteUrl - does not contain an = delimited from=to pair")
			}
			replacements = append(replacements, slice...)
		}
		guideReplacer = strings.NewReplacer(replacements...)
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		logger.Errorf(nil, "Error forming absolute path: %s", err)
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() {
			// Skip hidden directories TODO this should be applied to files also.
			_, node := filepath.Split(path)
			if node[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}

		ext := ""
		if strings.Index(path, ".") != -1 {
			ext = filepath.Ext(path)
		}

		//if ext == ".tmpl" { // Removed as may be too restrictive. What about images, css?
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			panic(err)
		}

		var meta map[string]string

		// The file may be in GFM, so convert to HTML
		if ext == ".md" {
			buf, meta = ProcessMarkdown(buf)

			// Now change extension to be .tmpl
			md := strings.TrimSuffix(rel, ext)
			rel = md + ".tmpl"
		}

		newname := prefix + "/" + rel

		logger.Tracef(nil, "Import file as '%s'\n", newname)

		if _, ok := _bindata[newname]; !ok {
			// Store the template, doing and search/replaces on the way
			_bindata[newname] = []byte(guideReplacer.Replace(string(buf)))
			if len(meta) > 0 {
				logger.Printf(nil, "Adding metadata for file %s\n", newname)
				_metadata[newname] = meta
			}
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// Returns rendered metadata and mpa of metadata key/value pairs
//
func ProcessMarkdown(doc []byte) ([]byte, map[string]string) {

	// Inspect the markdown src doc to see if it contains metadata
	reader := bytes.NewReader(doc)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	var newdoc string
	metaData := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		splitLine := strings.Split(line, ":")

		trimmed := strings.TrimSpace(splitLine[0])
		if (len(splitLine) < 2) || (!unicode.IsLetter(rune(trimmed[0]))) { // Have we reached a non KEY: line? If so, we're done with the metadata.
			if len(line) > 0 { // If the line is not empty, keep the contents
				newdoc = newdoc + line + "\n"
			}
			// Gather up all remainging lines
			for scanner.Scan() {
				// TODO Make this more efficient!
				newdoc = newdoc + scanner.Text() + "\n"
			}
			break
		}

		// Else, deal with meta-data
		metaValue := ""
		if len(splitLine) > 1 {
			metaValue = strings.TrimSpace(splitLine[1])
		}

		metaKey := strings.ToLower(splitLine[0])
		metaData[metaKey] = metaValue
	}

	doc = github_flavored_markdown.Markdown([]byte(newdoc))

	return doc, metaData
}

// ---------------------------------------------------------------------------
// end
