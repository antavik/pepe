package config

import (
	"text/template"
	"regexp"
	"strings"
	"fmt"
	"strconv"

	log "github.com/go-pkgz/lgr"
)

type C struct {
	Template  *template.Template
	TgEnabled bool
	SlEnabled bool
	Stdout    bool
	Stderr    bool
	Re        *regexp.Regexp
}

func MakeContainerConfig(name string, commonConf C, labels map[string]string) *C {
	c := &C{
		Template:  commonConf.Template,
		TgEnabled: commonConf.TgEnabled,
		SlEnabled: commonConf.SlEnabled,
		Stdout:    commonConf.Stdout,
		Stderr:    commonConf.Stderr,
		Re:        commonConf.Re,
	}

	for k, v := range labels {
		if !strings.HasPrefix(k, "pepe") {
			continue
		}

		parts := strings.Split(k, ".")
		conf := parts[1]
		switch conf {
		case "format":
			template, err := template.New(fmt.Sprintf("%s_template", name)).Parse(v)
			if err != nil {
				log.Printf("[WARNING] template parse error for %s: %v", name, err)
				continue
			}
			c.Template = template

		case "telegram":
			enabled, err := strconv.ParseBool(v)
			if err != nil {
				log.Printf("[WARNING] telegram enabled opt parse error for %s: %v", name, err)
				continue
			}
			c.TgEnabled = enabled

		case "slack":
			enabled, err := strconv.ParseBool(v)
			if err != nil {
				log.Printf("[WARNING] slack enabled opt parse error for %s: %v", name, err)
				continue
			}
			c.SlEnabled = enabled

		case "stdout":
			enabled, err := strconv.ParseBool(v)
			if err != nil {
				log.Printf("[WARNING] stdout opt parse error for %s: %v", name, err)
				continue
			}
			c.Stdout = enabled

		case "stderr":
			enabled, err := strconv.ParseBool(v)
			if err != nil {
				log.Printf("[WARNING] stderr opt parse error for %s: %v", name, err)
				continue
			}
			c.Stderr = enabled

		case "regexp":
			re, err := regexp.Compile(v)
			if err != nil {
				log.Printf("[WARNING] stream_regexp opt parse error for %s: %v", name, err)
				continue
			}
			c.Re = re
		}
	}

	return c
}