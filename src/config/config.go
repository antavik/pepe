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
	TemplateRaw string
	Template    *template.Template
	TgEnabled   bool
	SlEnabled   bool
	Stdout      bool
	Stderr      bool
	ReRaw       string
	Re          *regexp.Regexp
}

func MakeContainerConfig(name string, commonConf C, labels map[string]string) *C {
	c := C{
		TemplateRaw: commonConf.TemplateRaw,
		Template:    commonConf.Template,
		TgEnabled:   commonConf.TgEnabled,
		SlEnabled:   commonConf.SlEnabled,
		Stdout:      commonConf.Stdout,
		Stderr:      commonConf.Stderr,
		ReRaw:       commonConf.ReRaw,
		Re:          commonConf.Re,
	}

	for k, v := range labels {
		if !strings.HasPrefix(k, "pepe") {
			continue
		}

		parts := strings.Split(k, ".")
		conf := parts[1]
		switch conf {
		case "template":
			template, err := template.New(fmt.Sprintf("%s_template", name)).Parse(v)
			if err != nil {
				log.Printf("[WARNING] template parse error for %s: %v", name, err)
				continue
			}
			c.TemplateRaw = v
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

		case "regex":
			re, err := regexp.Compile(v)
			if err != nil {
				log.Printf("[WARNING] stream_regexp opt parse error for %s: %v", name, err)
				continue
			}
			c.ReRaw = v
			c.Re = re
		}
	}

	return &c
}

func (c *C) Map() map[string]interface{} {
	return map[string]interface{}{
		"template": c.TemplateRaw,
		"telegram": c.TgEnabled,
		"slack":    c.SlEnabled,
		"stdout":   c.Stdout,
		"stderr":   c.Stderr,
		"regex":    c.ReRaw,
	}
}