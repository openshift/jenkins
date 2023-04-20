package podman

// https://github.com/containers/podman/tree/main/pkg/bindings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/containers/podman/v4/libpod/define"
	"github.com/containers/podman/v4/pkg/api/handlers"
	"github.com/containers/podman/v4/pkg/bindings"
	bcontainers "github.com/containers/podman/v4/pkg/bindings/containers"
	bimages "github.com/containers/podman/v4/pkg/bindings/images"
	bvolumes "github.com/containers/podman/v4/pkg/bindings/volumes"
	"github.com/containers/podman/v4/pkg/domain/entities"
	"github.com/containers/podman/v4/pkg/domain/entities/reports"
	"github.com/containers/podman/v4/pkg/specgen"
)

type Client struct {
	Client *context.Context
}

func NewEnvClient() (*Client, error) {
	client, err := bindings.NewConnection(context.Background(), "unix://run/user/1000/podman/podman.sock")
	if err != nil {
		return nil, err
	}
	return &Client{Client: &client}, err
}

func (c *Client) ExecInActiveContainers(w io.Writer, ctx context.Context, cmd []string) {
	for {
		time.Sleep(60 * time.Second)
		select {
		case <-ctx.Done():
			return
		default:
			containers, err := bcontainers.List(*c.Client, &bcontainers.ListOptions{})
			if err != nil {
				fmt.Fprintf(w, "container list error: %#v\n", err)
				continue
			}
			fmt.Fprintf(w, "found %d containers\n", len(containers))
			for _, container := range containers {
				fmt.Fprintf(w, "found container %s running command %s\n", container.ID, container.Command)
				createConfig := new(handlers.ExecCreateConfig)
				createConfig.Cmd = cmd
				id, err := bcontainers.ExecCreate(*c.Client, container.ID, createConfig)
				if err != nil {
					fmt.Fprintf(w, "container ExecCreate error: %#v\n", err)
					continue
				}
				if err := bcontainers.ExecStart(*c.Client, id, &bcontainers.ExecStartOptions{}); err != nil {
					fmt.Fprintf(w, "container ExecStart error: %#v\n", err)
					continue
				}
				rPipe, wPipe, err := os.Pipe()
				if err != nil {
					fmt.Fprintf(w, "error creating pipes: %#v\n", err)
					continue
				}
				defer wPipe.Close()
				defer rPipe.Close()

				rErrPipe, wErrPipe, err := os.Pipe()
				if err != nil {
					fmt.Fprintf(w, "error creating pipes: %#v\n", err)
					continue
				}
				defer wErrPipe.Close()
				defer rErrPipe.Close()

				streams := define.AttachStreams{
					OutputStream: wPipe,
					ErrorStream:  wErrPipe,
					InputStream:  nil,
					AttachOutput: true,
					AttachError:  true,
					AttachInput:  false,
				}
				startAndAttachOptions := new(bcontainers.ExecStartAndAttachOptions)
				startAndAttachOptions.WithOutputStream(streams.OutputStream).WithErrorStream(streams.ErrorStream)
				if streams.InputStream != nil {
					startAndAttachOptions.WithInputStream(*streams.InputStream)
				}
				startAndAttachOptions.WithAttachError(streams.AttachError).WithAttachOutput(streams.AttachOutput).WithAttachInput(streams.AttachInput)
				if err := bcontainers.ExecStartAndAttach(*c.Client, id, startAndAttachOptions); err != nil {
					fmt.Fprintf(w, "container exec error: %#v\n", err)
					continue
				}

				bytes, err := io.ReadAll(rPipe)
				if err != nil {
					fmt.Fprintf(w, "error creating pipes: %#v\n", err)
					continue
				}

				fmt.Fprintf(w, "exec of command %#v into %s had text %s\n", cmd, container.ID, string(bytes))
			}
		}
	}
}

func (c *Client) InspectActiveContainers(w io.Writer, ctx context.Context) {
	for {
		time.Sleep(60 * time.Second)
		select {
		case <-ctx.Done():
			return
		default:
			containers, err := bcontainers.List(*c.Client, &bcontainers.ListOptions{})
			if err != nil {
				fmt.Fprintf(w, "container list error: %#v\n", err)
				continue
			}
			fmt.Fprintf(w, "found %d containers\n", len(containers))
			for _, container := range containers {
				fmt.Fprintf(w, "found container %s running command %s\n", container.ID, container.Command)
				data, err := bcontainers.Inspect(*c.Client, container.ID, &bcontainers.InspectOptions{})
				if err != nil {
					fmt.Fprintf(w, "container inspect error: %#v\n", err)
					continue
				}
				body, err := json.Marshal(data)
				if err != nil {
					fmt.Fprintf(w, "inspect of %s returned raw json %s\n", container.ID, string(body))
					continue
				}
				var prettyJSON bytes.Buffer
				error := json.Indent(&prettyJSON, body, "", "\t")
				if error != nil {
					fmt.Fprintf(w, "inspect of %s returned raw json %s\n", container.ID, string(body))
					continue
				}
				fmt.Fprintf(w, "inspect of %s returned formatted json:\n%s\n", container.ID, prettyJSON.String())
			}
		}
	}
}

func (c *Client) ContainerList() ([]entities.ListContainer, error) {
	return bcontainers.List(*c.Client, &bcontainers.ListOptions{})
}

func (c *Client) ContainerCreate(config *specgen.SpecGenerator) (string, error) {
	resp, err := bcontainers.CreateWithSpec(*c.Client, config, &bcontainers.CreateOptions{})
	return resp.ID, err
}

func (c *Client) ContainerExec(id string, cmd []string) (int, []byte, error) {
	createConfig := new(handlers.ExecCreateConfig)
	createConfig.Cmd = cmd
	createConfig.AttachStdout = true
	id, err := bcontainers.ExecCreate(*c.Client, id, createConfig)

	if err != nil {
		return 0, nil, err
	}
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		return 0, nil, err
	}
	defer wPipe.Close()
	defer rPipe.Close()

	rErrPipe, wErrPipe, err := os.Pipe()
	if err != nil {
		return 0, nil, err
	}
	defer wErrPipe.Close()
	defer rErrPipe.Close()

	streams := define.AttachStreams{
		OutputStream: wPipe,
		ErrorStream:  wErrPipe,
		InputStream:  nil,
		AttachOutput: true,
		AttachError:  true,
		AttachInput:  false,
	}
	startAndAttachOptions := new(bcontainers.ExecStartAndAttachOptions)
	startAndAttachOptions.WithOutputStream(streams.OutputStream).WithErrorStream(streams.ErrorStream)
	if streams.InputStream != nil {
		startAndAttachOptions.WithInputStream(*streams.InputStream)
	}
	startAndAttachOptions.WithAttachError(streams.AttachError).WithAttachOutput(streams.AttachOutput).WithAttachInput(streams.AttachInput)
	if err := bcontainers.ExecStartAndAttach(*c.Client, id, startAndAttachOptions); err != nil {
		return 0, nil, err
	}

	bytes, err := io.ReadAll(rPipe)
	if err != nil {
		return 0, nil, err
	}

	inspect, err := bcontainers.ExecInspect(*c.Client, id, &bcontainers.ExecInspectOptions{})
	if err != nil {
		return 0, nil, err
	}

	return inspect.ExitCode, bytes, nil
}

func (c *Client) ContainerInspect(id string) (string, error) {
	data, err := bcontainers.Inspect(*c.Client, id, &bcontainers.InspectOptions{})
	if err != nil {
		return "", err
	}
	return data.NetworkSettings.IPAddress, nil
}

func (c *Client) ContainerStart(id string) error {
	return bcontainers.Start(*c.Client, id, &bcontainers.StartOptions{})
}

func (c *Client) ContainerLogs(id string) ([]byte, error) {
	fmt.Printf("Container ID: %s \n", id)
	data, err := bcontainers.Inspect(*c.Client, id, &bcontainers.InspectOptions{})
	if err != nil {
		return []byte{}, err
	}
	fmt.Printf("DATA: %s", data.State.Status)
	truePtr := true
	stdOutChan := make(chan string, 100)
	stdErrChan := make(chan string, 100)

	if err := bcontainers.Logs(*c.Client, id, &bcontainers.LogOptions{
		Stdout: &truePtr,
		Stderr: &truePtr,
	}, stdOutChan, stdErrChan); err != nil {
		return nil, err
	}

	close(stdOutChan)
	close(stdErrChan)

	allLogs := []string{}
	for e := range stdOutChan {
		allLogs = append(allLogs, e)
		fmt.Printf("stdOutChan: %s", e)
	}
	for e := range stdErrChan {
		allLogs = append(allLogs, e)
		fmt.Printf("stdErrChan: %s", e)
	}

	return []byte(strings.Join(allLogs, "\n")), nil
}

func (c *Client) ContainerRemove(id string) ([]*reports.RmReport, error) {
	return bcontainers.Remove(*c.Client, id, &bcontainers.RemoveOptions{})
}

func (c *Client) ContainerStop(id string, timeout int) error {
	stopOptions := new(bcontainers.StopOptions)
	stopOptions.WithTimeout(uint(30))
	return bcontainers.Stop(*c.Client, id, stopOptions)
}

func (c *Client) ContainerStopAndRemove(id string, timeout int) ([]*reports.RmReport, error) {
	err := c.ContainerStop(id, timeout)
	if err != nil {
		return nil, err
	}
	return c.ContainerRemove(id)
}

func (c *Client) ContainerWait(id string) (int32, error) {
	return bcontainers.Wait(*c.Client, id, &bcontainers.WaitOptions{})
}

func (c *Client) ImagesRemove(names []string) (*entities.ImageRemoveReport, []error) {
	return bimages.Remove(*c.Client, names, &bimages.RemoveOptions{})
}

func (c *Client) VolumeCreate() (*entities.VolumeConfigResponse, error) {
	return bvolumes.Create(*c.Client, entities.VolumeCreateOptions{}, &bvolumes.CreateOptions{})
}

func (c *Client) VolumeRemove(name string) error {
	return bvolumes.Remove(*c.Client, name, &bvolumes.RemoveOptions{})
}

func Duration(d time.Duration) *time.Duration {
	return &d
}
