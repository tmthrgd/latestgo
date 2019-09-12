package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

// unpackedOkay is a sentinel zero-byte file to indicate that the Go version
// was downloaded and unpacked successfully. It is copied from
// golang.org/dl/internal/version.
const unpackedOkay = ".unpacked-success"

// These are the golang.org URLs we fetch the release JSON from.
const (
	dlFeedURL    = "https://golang.org/dl/?mode=json"
	dlFeedAllURL = "https://golang.org/dl/?mode=json&include=all"
)

var allFlag = flag.Bool("all", false, "download all go releases since go1.8\n    includes all patch releases")

func main() {
	log.SetFlags(0)
	log.SetPrefix("latestgo: ")

	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the list of releases to download.
	releases, err := listReleases()
	if err != nil {
		log.Fatal(err)
	}

	var (
		downloaded bool
		latest     string
	)
	for _, release := range releases {
		if !validRelease(release.Version) || versionTooOld(release.Version) {
			continue
		}

		// Only consider stable releases when determining the latest
		// version.
		if release.Stable {
			latest = maxVersion(latest, release.Version)
		}

		// Determine if this go version has already been installed.
		unpackedOkayPath := filepath.Join(home, "sdk", release.Version, unpackedOkay)
		if _, err := os.Stat(unpackedOkayPath); err == nil {
			continue
		}

		downloaded = true
		fmt.Printf("Downloading %s\n", release.Version)

		if err := downloadRelease(release.Version); err != nil {
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

type releaseJSON struct {
	Version string
	Stable  bool
}

// listReleases returns a list of go releases from golang.org.
//
// If the -all flag was provided, it will return all go releases, otherwise
// only the latest supported go releases will be returned.
func listReleases() ([]releaseJSON, error) {
	url := dlFeedURL
	if *allFlag {
		url = dlFeedAllURL
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned non-200 OK status code: %s",
			resp.Request.URL.Hostname(), resp.Status)
	}

	r := io.LimitReader(resp.Body, 128<<20+1) // 128MiB

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if r.(*io.LimitedReader).N <= 0 {
		return nil, fmt.Errorf("%s returned an excessively large JSON response",
			resp.Request.URL.Hostname())
	}

	var releases []releaseJSON
	if err := json.Unmarshal(data, &releases); err != nil {
		return nil, err
	}

	return releases, nil
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

// versionTooOld returns whether the go version is too old to be downloaded
// with the golang.org/dl/go1.X[.Y] installers.
//
// v must be a valid go version.
func versionTooOld(v string) bool {
	v = "v" + strings.TrimPrefix(v, "go")
	return semver.Compare(v, "v1.8") < 0
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
