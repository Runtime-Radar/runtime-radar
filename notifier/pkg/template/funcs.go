package template

import (
	"encoding/json"
	"strconv"
	"strings"
	"text/template"

	enforcer "github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
	"github.com/tidwall/gjson"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var funcs = template.FuncMap{
	"translate":    translate,
	"title":        title,
	"toLower":      strings.ToLower,
	"toTitle":      strings.ToTitle,
	"toUpper":      strings.ToUpper,
	"trim":         strings.Trim,
	"trimLeft":     strings.TrimLeft,
	"trimRight":    strings.TrimRight,
	"trimSpace":    strings.TrimSpace,
	"trimPrefix":   strings.TrimPrefix,
	"trimSuffix":   strings.TrimSuffix,
	"hasPrefix":    strings.HasPrefix,
	"hasSuffix":    strings.HasSuffix,
	"join":         strings.Join,
	"split":        strings.Split,
	"parseInt":     strconv.Atoi,
	"parseFloat64": parseFloat64,
	"sub":          sub,
	"queryJSON":    queryJSON,
	"escapeJSON":    escapeJSON,
	"sum":           sum,
	"tacticName":    tacticName,
	"techniqueName": techniqueName,
}

var translations = map[string]string{
	enforcer.NoneSeverity.String():     "undefined",
	enforcer.LowSeverity.String():      "low",
	enforcer.MediumSeverity.String():   "medium",
	enforcer.HighSeverity.String():     "high",
	enforcer.CriticalSeverity.String(): "critical",
}

func translate(s string) string {
	if t, ok := translations[s]; ok {
		return t
	}
	return s
}

// title is a replacement for strings.Title as it's deprecated.
func title(s string) string {
	return cases.Title(language.Und, cases.NoLower).String(s)
}

func sub(a, b int) int {
	return a - b
}

func sum(a, b int) int {
	return a + b
}

func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func queryJSON(path, json string) gjson.Result {
	return gjson.Get(json, path)
}

func escapeJSON(s string) (string, error) {
	marshalled, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	s2 := string(marshalled)

	// trim quotes added by json.Marshal,
	// so that result string can be surrounded with quotes in template just as any other string
	s2 = strings.TrimPrefix(s2, `"`)
	s2 = strings.TrimSuffix(s2, `"`)

	return s2, nil
}
