# latestgo

**latestgo** is a tool to automatically install the latest supported golang
versions.

## Installation

Installation is simple and no different to any Go tool. The only requirement is
a working [Go](https://golang.org/) install.

```bash
go get go.tmthrgd.dev/latestgo
```

## Usage

```bash
latestgo
```

That's it. That's all there is to it.

**latestgo** will install the `golang.org/dl` wrappers for the latest patch
release for each supported version of go. This means you can run a specific go
version by running `go1.X.Y`. The go distribution will be installed into
`$HOME/sdk/go1.X.Y`.

As a convenience, **latestgo** writes the latest supported version of go to
`$HOME/sdk/latest`. This file can be used by shell scripts to find the latest
go binary. For example, in my `.bash_profile` I add
`$HOME/sdk/$(cat ~/sdk/latest)/bin` to `$PATH` so `go` will always be the
latest version of go installed.

## License

[BSD 3-Clause License](LICENSE)

## Note

**latestgo** uses the Golang maintner service at maintner.golang.org to
retrieve the latest supported versions.
