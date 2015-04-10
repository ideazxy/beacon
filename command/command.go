package command

import (
	"encoding/json"
	"log"
)

type Command struct {
	Type   string // add|update|remove
	Proto  string // tcp|http
	Image  string
	Cmd    []string
	Env    []string
	Vol    []string
	Listen string // only one port will be registerd
}

func Unmarshal(raw string, cmd *Command) error {
	return json.Unmarshal([]byte(raw), cmd)
}

func (c *Command) Process() error {
	log.Println(c.Marshal())
	return nil
}

func (c *Command) run() error {
	return nil
}

func (c *Command) register() error {
	return nil
}

func (c *Command) Marshal() string {
	b, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(b)
}
