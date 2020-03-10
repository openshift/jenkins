# go-init

**go-init** is a minimal init system with simple *lifecycle management* heavily inspired by [dumb-init](https://github.com/Yelp/dumb-init).

It is designed to run as the first process (PID 1) inside a container.

It is lightweight (less than 500KB after UPX compression) and statically linked so you don't need to install any dependency.

## Download

You can download the latest version on [releases page](https://github.com/pablo-ruth/go-init/releases)

## Why you need an init system

I can't explain it better than Yelp in *dumb-init* repo, so please [read their explanation](https://github.com/Yelp/dumb-init/blob/v1.2.0/README.md#why-you-need-an-init-system)

Summary:
- Proper signal forwarding
- Orphaned zombies reaping

## Why another minimal init system

In addition to *init* problematic, **go-init** tries to solve another Docker flaw by adding *hooks* on start and stop of the main process.

If you want to launch a command before the main process of your container and another one after the main process exit, you can't with Docker, see [issue 6982](https://github.com/moby/moby/issues/6982)

With **go-init** you can do that with "pre" and "post" hooks.

## Usage

### one command

```
$ go-init -main "my_command param1 param2"
```

### pre-start and post-stop hooks

```
$ go-init -pre "my_pre_command param1" -main "my_command param1 param2" -post "my_post_command param1"
```

## docker

Example of Dockerfile using *go-init*:
```
FROM alpine:latest

COPY go-init /bin/go-init

RUN chmod +x /bin/go-init

ENTRYPOINT ["go-init"]

CMD ["-pre", "echo hello world", "-main", "sleep 5", "-post", "echo I finished my sleep bye"]
```

Build it:
```
docker build -t go-init-example
```

Run it:
```
docker run go-init-example
```
