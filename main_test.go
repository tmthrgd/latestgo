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

func TestMaxVersion(t *testing.T) {
	assert.Equal(t, "go1.12", maxVersion("go1.12", ""))
	assert.Equal(t, "go1.12.6", maxVersion("go1.12.6", ""))
	assert.Equal(t, "go1.12.6", maxVersion("go1.12.6", "go1.11.11"))
	assert.Equal(t, "go1.12.6", maxVersion("go1.11.11", "go1.12.6"))
}
