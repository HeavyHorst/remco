[![Build Status](https://travis-ci.org/HeavyHorst/remco.svg?branch=master)](https://travis-ci.org/HeavyHorst/remco) [![Go Report Card](https://goreportcard.com/badge/github.com/HeavyHorst/remco)](https://goreportcard.com/report/github.com/HeavyHorst/remco) [![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/HeavyHorst/remco/master/LICENSE)
#Remco
remco is a lightweight configuration management tool. It's highly influenced by [confd](https://github.com/kelseyhightower/confd).
Remcos main purposes are (like confd's):

  - keeping local configuration files up-to-date using data stored in a key/value store like etcd or consul and processing template resources.
  - reloading applications to pick up new config file changes

## Differences between remco and confd
   - Multiple source/destination pairs per template resource - useful for programs that need more than one config file
   - Multiple backends per template resource - get normal config values from etcd and secrets from vault
   - [Pongo2](https://github.com/flosch/pongo2) template engine instead of go's text/template
   - Zombie reaping support (if remco runs as pid 1)
   - Additional backends can be provided as plugins.
   - [Exec](https://heavyhorst.github.io/remco/details/exec-mode/) mode similar to consul-template.

## Documentation 
See: https://heavyhorst.github.io/remco/


## Building from source
  - go get github.com/HeavyHorst/remco
  - go install github.com/HeavyHorst/remco

  You should now have remco in your $GOPATH/bin directory

## Project Status

Remco is under active development. We do not recommend its use in production, but we encourage you to try out Remco and provide feedback via issues and pull requests.
