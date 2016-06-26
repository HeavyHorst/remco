package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

type tomlConf struct {
	Config []struct {
		FileMode string
		Cmd      struct {
			Check  string
			Reload string
		}
		Template struct {
			Src string
			Dst string
		}
		Backend struct {
			Name       string
			Prefix     string
			Interval   int
			Keys       []string
			Etcdconfig struct {
				Nodes   []string
				Version int
			}
		}
	}
}

func main() {
	f, err := os.Open("sample.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var c tomlConf
	if err := toml.Unmarshal(buf, &c); err != nil {
		panic(err)
	}

	for _, v := range c.Config {
		fmt.Println(v.Backend.Etcdconfig.Nodes)
	}
}
