package docker

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
)

type Client struct {
	Client *client.Client
}

func NewEnvClient() (*Client, error) {
	client, err := client.NewEnvClient()
	return &Client{Client: client}, err
}

func (c *Client) ContainerCreate(config *container.Config, hostconfig *container.HostConfig) (string, error) {
	body, err := c.Client.ContainerCreate(context.Background(), config, hostconfig, nil, "")
	return body.ID, err
}

func (c *Client) ContainerExec(id string, cmd []string) (int, []byte, error) {
	exec, err := c.Client.ContainerExecCreate(context.Background(), id, types.ExecConfig{
		AttachStdout: true,
		Cmd:          cmd,
	})
	if err != nil {
		return 0, nil, err
	}

	attach, err := c.Client.ContainerExecAttach(context.Background(), exec.ID, types.ExecConfig{})
	if err != nil {
		return 0, nil, err
	}
	defer attach.Close()

	bytes, err := ioutil.ReadAll(attach.Reader)
	if err != nil {
		return 0, nil, err
	}

	inspect, err := c.Client.ContainerExecInspect(context.Background(), exec.ID)
	if err != nil {
		return 0, nil, err
	}

	return inspect.ExitCode, bytes, nil
}

func (c *Client) ContainerInspect(id string) (string, error) {
	json, err := c.Client.ContainerInspect(context.Background(), id)
	if err != nil {
		return "", err
	}

	return json.NetworkSettings.IPAddress, nil
}

func (c *Client) ContainerStart(id string) error {
	return c.Client.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
}

func (c *Client) ContainerLogs(id string) ([]byte, error) {
	r, err := c.Client.ContainerLogs(context.Background(), id, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(r)
}

func (c *Client) ContainerRemove(id string) error {
	return c.Client.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{})
}

func (c *Client) ContainerStop(id string, timeout *time.Duration) error {
	return c.Client.ContainerStop(context.Background(), id, timeout)
}

func (c *Client) ContainerStopAndRemove(id string, timeout *time.Duration) error {
	err := c.ContainerStop(id, timeout)
	if err != nil {
		return err
	}
	return c.ContainerRemove(id)
}

func (c *Client) ContainerWait(id string) (int, error) {
	return c.Client.ContainerWait(context.Background(), id)
}

func (c *Client) ImageRemove(name string) error {
	_, err := c.Client.ImageRemove(context.Background(), name, types.ImageRemoveOptions{})
	return err
}

func (c *Client) VolumeCreate() (string, error) {
	vol, err := c.Client.VolumeCreate(context.Background(), types.VolumeCreateRequest{})
	return vol.Name, err
}

func (c *Client) VolumeRemove(name string) error {
	return c.Client.VolumeRemove(context.Background(), name)
}

func Duration(d time.Duration) *time.Duration {
	return &d
}
