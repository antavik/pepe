package source

import (
	"github.com/antibantique/pepe/src/config"
)

type S struct {
	Id     string
	Ip     string
	Name   string
	Config *config.C
}

func (s S) String() string {
	var src string

	switch {
	case s.Name != "":
		src = s.Name
	case s.Id != "":
		src = s.Id
	default:
		src = s.Ip
	}

	return src
}

func (s *S) Map() map[string]interface{} {
	return map[string]interface{}{
		"id":     s.Id,
		"ip":     s.Ip,
		"name":   s.Name,
		"config": s.Config.Map(),
	}
}