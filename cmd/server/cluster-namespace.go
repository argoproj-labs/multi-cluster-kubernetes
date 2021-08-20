package server

import (
	"fmt"
	"strings"
)

func join(clusterName, namespace string) string {
	if clusterName == "" {
		return namespace
	}
	return fmt.Sprintf("%s.%s", clusterName, namespace)
}

func split(s string) (string, string) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return "default", s
	}
	return parts[0], parts[1]
}
