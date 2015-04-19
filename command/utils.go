package command

import (
	"strings"
)

func parseImageName(name string) (string, string) {
	n := strings.Index(name, "@")
	if n >= 0 {
		parts := strings.Split(name, "@")
		return parts[0], parts[1]
	}
	n = strings.LastIndex(name, ":")
	if n < 0 {
		return name, ""
	}
	if tag := name[n+1:]; !strings.Contains(tag, "/") {
		return name[:n], tag
	}
	return name, ""
}

// splitReposName breaks a reposName into an index name and remote name
func splitReposName(reposName string) (string, string) {
	nameParts := strings.SplitN(reposName, "/", 2)
	var indexName, remoteName string
	if len(nameParts) == 1 || (!strings.Contains(nameParts[0], ".") &&
		!strings.Contains(nameParts[0], ":") && nameParts[0] != "localhost") {
		// This is a Docker Index repos (ex: samalba/hipache or ubuntu)
		indexName = "docker.io"
		remoteName = reposName
	} else {
		indexName = nameParts[0]
		remoteName = nameParts[1]
	}
	return indexName, remoteName
}
