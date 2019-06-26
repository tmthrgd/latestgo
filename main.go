package main

// We can't just use golang.org/x/build/maintner/maintnerd/apipb directly as it
// imports the experimental franken package grpc.go4.org that isn't compatible
// with google.golang.org/grpc. Instead we generate our own gRPC code using the
// same api.proto file.
//
//go:generate bash -c "protoc -I$(go list -e -f '{{.Dir}}' golang.org/x/build/maintner/maintnerd/apipb) api.proto --go_out=plugins=grpc:internal/proto"

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	pb "go.tmthrgd.dev/latestgo/internal/proto"
	"golang.org/x/mod/semver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// unpackedOkay is a sentinel zero-byte file to indicate that the Go version
// was downloaded and unpacked successfully. It is copied from
// golang.org/dl/internal/version.
const unpackedOkay = ".unpacked-success"

func main() {
	log.SetFlags(0)
	log.SetPrefix("latestgo: ")

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the Golang maintner server with standard TLS validation.
	cc, err := grpc.Dial("dns:///maintner.golang.org",
		grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	if err != nil {
		log.Fatal(err)
	}

	pc := pb.NewMaintnerServiceClient(cc)

	// Retrieve the list of latest supported releases.
	//
	// "By default, ListGoReleases returns only the latest patches
	//  of releases that are considered supported per policy."
	resp, err := pc.ListGoReleases(context.Background(), new(pb.ListGoReleasesRequest))
	if err != nil {
		log.Fatal(err)
	}

	var (
		downloaded bool
		latest     string
	)
	for _, release := range resp.Releases {
		if !validRelease(release.TagName) {
			continue
		}

		latest = maxVersion(latest, release.TagName)

		// Determine if this go version has already been installed.
		unpackedOkayPath := filepath.Join(home, "sdk", release.TagName, unpackedOkay)
		if _, err := os.Stat(unpackedOkayPath); err == nil {
			continue
		}

		downloaded = true
		fmt.Printf("Downloading %s\n", release.TagName)

		if err := downloadRelease(release.TagName); err != nil {
			log.Fatal(err)
		}
	}

	// If we found the latest version, write it to $HOME/sdk/latest for
	// scripts to use.
	if latest != "" {
		latestPath := filepath.Join(home, "sdk", "latest")
		if err := ioutil.WriteFile(latestPath, []byte(latest), 0644); err != nil {
			log.Fatal(err)
		}
	}

	if !downloaded {
		fmt.Println("Already up to date.")
	}
}

// downloadRelease installs go version v using golang.org/dl.
func downloadRelease(v string) error {
	goToolPath := filepath.Join(runtime.GOROOT(), "bin", "go")
	goWrapperPath := filepath.Join(gobin(), v)

	// TODO(tmthrgd): module aware global go install command.
	cmd := exec.Command(goToolPath, "get", "golang.org/dl/"+v)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), "GO111MODULE=off")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command(goWrapperPath, "download")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// validRelease reports whether v is a valid go1.X[.Y] version number.
func validRelease(v string) bool {
	if !strings.HasPrefix(v, "go1.") {
		return false
	}

	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}

	if _, err := strconv.ParseUint(parts[1], 10, 64); err != nil {
		return false
	}

	if len(parts) == 2 {
		return true
	}

	patch, err := strconv.ParseUint(parts[2], 10, 64)
	return err == nil && patch > 0
}

// maxVersion returns the version string that compares greatest.
//
// v1 and v2 must be valid go versions or empty.
func maxVersion(v1, v2 string) string {
	if v1 != "" {
		v1 = "v" + strings.TrimPrefix(v1, "go")
	}
	if v2 != "" {
		v2 = "v" + strings.TrimPrefix(v2, "go")
	}

	v := semver.Max(v1, v2)
	v = strings.TrimSuffix(v, ".0") // go doesn't use .0 patch versions.
	return "go" + strings.TrimPrefix(v, "v")
}

// gobin returns the directory go get installs binaries into. It uses $GOBIN if
// set, or $GOPATH/bin if $GOPATH is set, or $HOME/go/bin if $HOME is set. If
// the directory cannot be determined, it returns an empty string.
func gobin() string {
	if s := os.Getenv("GOBIN"); s != "" {
		return s
	}

	if s := os.Getenv("GOPATH"); s != "" {
		return filepath.Join(filepath.SplitList(s)[0], "bin")
	}

	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, "go", "bin")
	}

	return ""
}
