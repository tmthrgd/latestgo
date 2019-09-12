package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidRelease(t *testing.T) {
	for _, v := range []string{
		"go1.12",
		"go1.12.6",
	} {
		assert.Truef(t, validRelease(v), "validRelease(%q)", v)
	}

	for _, v := range []string{
		"go1",
		"go1.12.0",
		"go1.-12.6",
		"go1.12.-6",
		"go2.1",
		"GO1",
		"v1.12",
		"v1.12.6",
		"1.12",
		"1.12.6",
		"evil/../version",
		"gorandom",
		"go1.random",
		"go1.12.random",
	} {
		assert.Falsef(t, validRelease(v), "validRelease(%q)", v)
	}
}

func TestVersionTooOld(t *testing.T) {
	for _, v := range []string{
		"go1",
		"go1.2",
		"go1.2.2",
		"go1.4",
		"go1.4.1",
		"go1.7",
		"go1.7.1",
		"go1.7.99",
		"go1.7.999999",
	} {
		assert.Truef(t, versionTooOld(v), "versionTooOld(%q)", v)
	}

	for _, v := range []string{
		"go1.8",
		"go1.8.1",
		"go1.12",
		"go1.12.6",
		"go1.13",
		"go1.13.1",
		"go2",
		"go2.1",
	} {
		assert.Falsef(t, versionTooOld(v), "versionTooOld(%q)", v)
	}
}

func TestMaxVersion(t *testing.T) {
	assert.Equal(t, "go1.12", maxVersion("go1.12", ""))
	assert.Equal(t, "go1.12.6", maxVersion("go1.12.6", ""))
	assert.Equal(t, "go1.12.6", maxVersion("go1.12.6", "go1.11.11"))
	assert.Equal(t, "go1.12.6", maxVersion("go1.11.11", "go1.12.6"))
}
