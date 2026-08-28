package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func LookupEnvString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func LookupEnvBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if val == "" || val == "0" || val == "false" {
			return false
		}
		return true
	}
	return defaultVal
}

func LookupEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal().Msgf("### Can't convert string '%s' to int: %v", val, err)
		}
		return v
	}
	return defaultVal
}

func LookupEnvFloat64(key string, defaultVal float64) float64 {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			log.Fatal().Msgf("### Can't convert string '%s' to float64: %v", val, err)
		}
		return v
	}
	return defaultVal
}

func LookupEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		v, err := time.ParseDuration(val)
		if err != nil {
			log.Fatal().Msgf("### Can't convert string '%s' to time.Duration: %v", val, err)
		}
		return v
	}
	return defaultVal
}

// StringList is a comma separated flag value, e.g. "https://a.example,https://b.example".
type StringList []string

func (l *StringList) String() string {
	if l == nil {
		return ""
	}
	return strings.Join(*l, ",")
}

// Set appends to the current value, so a repeated flag accumulates:
// -flag=a,b -flag=c gives [a b c].
func (l *StringList) Set(value string) error {
	*l = append(*l, splitList(value)...)
	return nil
}

// stringListValue drops the env default the first time the flag appears on the command line.
// Without it the env value and the command line one would merge into a list wider than either.
type stringListValue struct {
	list *StringList
	seen bool
}

func (v *stringListValue) String() string {
	return v.list.String()
}

func (v *stringListValue) Set(value string) error {
	if !v.seen {
		*v.list, v.seen = nil, true
	}

	return v.list.Set(value)
}

// StringListVar defines a comma separated list flag with a default taken from env.
// flag.Var accepts no default of its own, hence the helper.
func StringListVar(p *StringList, name, key, defaultVal, usage string) {
	*p = splitList(LookupEnvString(key, defaultVal))
	flag.Var(&stringListValue{list: p}, name, usage)
}

// splitList turns a comma separated setting into a list, dropping empty entries and whitespace.
func splitList(value string) []string {
	res := []string{}

	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			res = append(res, item)
		}
	}

	return res
}
