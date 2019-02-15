package scripts

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/i18n/translations"
	"github.com/hexya-erp/hexya/src/tools/strutils"
)

// A customMessage holds the custom translation string for the given ID
type customMessage struct {
	ID     string `json:"id"`
	String string `json:"string"`
}

// A moduleCustomMessageList is the list of all custom translations of a module
type moduleCustomMessageList struct {
	Messages []customMessage `json:"messages"`
}

// A LangCustomMaps holds custom translations for all modules for a given language
type LangCustomMaps map[string]moduleCustomMessageList

// A customTranslationsMaps holds all custom translations for all modules and all languages
type customTranslationsMaps map[string]LangCustomMaps

// langModuleTranslationsMap is the memory cache for custom translations
var langModuleTranslationsMap customTranslationsMaps

// loadCustomTranslationsMap populates a customTranslationsMaps for this application.
func loadCustomTranslationsMap() customTranslationsMaps {
	out := make(customTranslationsMaps)
	allTrans := i18n.GetAllCustomTranslations()
	for lang, entry := range allTrans {
		if out[lang] == nil {
			out[lang] = make(LangCustomMaps)
		}
		for module, trans := range entry {
			for k, v := range trans {
				msg := customMessage{
					ID:     k,
					String: v,
				}
				list := out[lang][module]
				list.Messages = append(list.Messages, msg)
				out[lang][module] = list
			}
		}
	}
	return out
}

// ListModuleTranslations returns a map containing all custom translations of a module for the given language
func ListModuleTranslations(lang string) LangCustomMaps {
	if langModuleTranslationsMap == nil {
		langModuleTranslationsMap = loadCustomTranslationsMap()
	}
	return langModuleTranslationsMap[lang]
}

func addToTranslationMap(messages translations.MessageMap, lang, moduleName, value, refFile string, refLine int) translations.MessageMap {
	translated := i18n.TranslateCustom(lang, value, moduleName)
	if translated == value {
		translated = ""
	}
	msgRef := translations.MessageRef{MsgId: value}
	msg := translations.GetOrCreateMessage(messages, msgRef, translated)
	msg.ExtractedComment = fmt.Sprintf("custom: %s", moduleName)
	if refFile != "" {
		msg.ReferenceFile = append(msg.ReferenceFile, refFile)
		msg.ReferenceLine = append(msg.ReferenceLine, refLine)
	}
	messages[msgRef] = msg
	return messages
}

func updateFuncJS(messages translations.MessageMap, lang, filePath, moduleName string) translations.MessageMap {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return messages
	}
	var tVar = `_t`
	var coreVar = `core`
	rxT := regexp.MustCompile(tVar + `\("(.*?)"\)`)
	rxTVar := regexp.MustCompile(`var (.*?) = ` + coreVar + `\._t`)
	rxCoreVar := regexp.MustCompile(`var (.*?) = require\(web\.core\)`)
	for i, line := range strings.Split(string(content), "\n") {
		switch {
		case rxT.MatchString(line):
			matches := rxT.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				addToTranslationMap(messages, lang, moduleName, match[1], filepath.Base(filePath), i)
			}
		case rxTVar.MatchString(line):
			matches := rxTVar.FindStringSubmatch(line)
			tVar = matches[1]
			rxT = regexp.MustCompile(tVar + `\("(.*?)"\)`)
		case rxCoreVar.MatchString(line):
			matches := rxCoreVar.FindStringSubmatch(line)
			coreVar = matches[1]
			rxTVar = regexp.MustCompile(`var (.*?) = ` + coreVar + `\._t`)
		}
	}
	return messages
}

// A Node is an XML Node used for walking down the tree.
type Node struct {
	XMLName xml.Name
	Content []byte     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
	Attrs   []xml.Attr `xml:",attr"`
}

func walk(nodes []Node, f func(Node, string) (bool, string), str string) {
	for _, n := range nodes {
		if ok, strNew := f(n, str); ok {
			walk(n.Nodes, f, strNew)
		}
	}
}

func updateFuncXML(messages translations.MessageMap, lang, filePath, moduleName string) translations.MessageMap {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(fmt.Errorf("unable to read file %s: %s", filePath, err))
	}
	buf := bytes.NewBuffer(data)
	dec := xml.NewDecoder(buf)
	var n Node
	err = dec.Decode(&n)
	if err != nil {
		panic(fmt.Errorf("unable to read file %s: %s", filePath, err))
	}
	walk([]Node{n}, func(n Node, xmlPath string) (bool, string) {
		content := strings.TrimSpace(string(n.Content))
		for _, attr := range n.Attrs {
			if strutils.IsIn(attr.Name.Local, `title`, `alt`, `label`, `placeholder`) && len(attr.Value) > 0 {
				addToTranslationMap(messages, lang, moduleName, attr.Value, filepath.Base(filePath), 0)
			}
		}
		if len(content) > 0 && !strings.HasPrefix(content, "<") {
			addToTranslationMap(messages, lang, moduleName, content, filepath.Base(filePath), 0)
		}
		return true, path.Join(xmlPath, n.XMLName.Local)
	}, ".")
	return messages
}

// UpdateFunc is the function that extracts strings to translate from XML and JS files.
func UpdateFunc(messages translations.MessageMap, lang, path, moduleName string) translations.MessageMap {
	if filepath.Ext(path) == ".js" {
		return updateFuncJS(messages, lang, path, moduleName)
	}
	return updateFuncXML(messages, lang, path, moduleName)
}
