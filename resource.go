package main

import (
	"context"
	"time"

	"github.com/HeavyHorst/remco/backends"
	backendErrors "github.com/HeavyHorst/remco/backends/error"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
)

type resource struct {
	Template []*template.ProcessConfig
	Backend  backends.Config

	// defaults to the filename of the resource
	Name string
}

func (r *resource) run(ctx context.Context) {
	var backendList template.Backends

	// try to connect to all backends
	// connection to all backends must succeed to continue
	for _, config := range r.Backend.GetBackends() {
	retryloop:
		for {
			select {
			case <-ctx.Done():
				for _, b := range backendList {
					b.Close()
				}
				return
			default:
				b, err := config.Connect()
				if err == nil {
					backendList = append(backendList, b)
				} else if err != backendErrors.ErrNilConfig {

					log.WithFields(logrus.Fields{
						"backend":  b.Name,
						"resource": r.Name,
					}).Error(err)

					// try again every 2 seconds
					time.Sleep(2 * time.Second)
					continue retryloop
				}
				// break out of the loop on success or if the backend is nil
				break retryloop
			}
		}
	}

	t, err := template.NewResource(backendList, r.Template, r.Name)
	// make sure that all backend clients are closed cleanly
	defer t.Close()
	if err != nil {
		log.Error(err.Error())
		return
	}

	t.Monitor(ctx)
}
