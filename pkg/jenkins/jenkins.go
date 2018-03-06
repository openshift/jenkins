package jenkins

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/docker/engine-api/types/container"
	"github.com/openshift/jenkins/pkg/docker"
)

type Jenkins struct {
	ID     string
	ip     string
	Volume string
	Client *docker.Client
}

func NewJenkins(client *docker.Client) *Jenkins {
	return &Jenkins{Client: client}
}

func (j *Jenkins) CreateJob(name, password, filename string) (*http.Response, error) {
	xml, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer xml.Close()

	req, err := http.NewRequest("POST", "http://"+j.ip+":8080/createItem?name="+name, xml)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("admin", password)
	req.Header.Set("Content-Type", "application/xml")

	return http.DefaultClient.Do(req)
}

func (j *Jenkins) GetJob(name, password string) (*http.Response, error) {
	req, err := http.NewRequest("GET", "http://"+j.ip+":8080/job/"+name, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("admin", password)

	return http.DefaultClient.Do(req)
}

func (j *Jenkins) Start(image string, env []string) error {
	var err error
	j.ID, err = j.Client.ContainerCreate(
		&container.Config{
			Image: image,
			Env:   env,
			Tty:   true,
		},
		&container.HostConfig{
			Binds: []string{
				j.Volume + ":/var/lib/jenkins",
			},
		})
	if err != nil {
		return err
	}

	err = j.Client.ContainerStart(j.ID)
	if err != nil {
		return err
	}

	j.ip, err = j.Client.ContainerInspect(j.ID)
	if err != nil {
		return err
	}

	return j.wait()
}

func (j *Jenkins) wait() error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	for {
		reqctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		req, err := http.NewRequest("GET", "http://"+j.ip+":8080/login", nil)
		if err != nil {
			return err
		}
		req = req.WithContext(reqctx)

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return errors.New("timeout")
		default:
		}

		<-reqctx.Done()
	}
}
