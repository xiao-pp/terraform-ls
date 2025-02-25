package main

import (
	goversion "github.com/hashicorp/go-version"
)

// The main version number that is being run at the moment.
var version = "0.0.0"

// A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
var versionPrerelease = "dev"

// Version (build) metadata, such as homebrew
var versionMetadata = ""

func init() {
	// Verify that the version is proper semantic version, which should always be the case.
	_, err := goversion.NewVersion(version)
	if err != nil {
		panic(err.Error())
	}
}

// VersionString returns the complete version string, including prerelease
func VersionString() string {
	v := version
	if versionPrerelease != "" {
		v += "-" + versionPrerelease
	}

	if versionMetadata != "" {
		v += "+" + versionMetadata
	}

	return v
}
