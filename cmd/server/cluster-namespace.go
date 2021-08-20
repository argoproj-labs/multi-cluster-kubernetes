package server

import (
	"fmt"
	"strings"
)

func join(clusterName, namespace string) string {
	return fmt.Sprintf("%s.%s", clusterName, namespace)
}

func split(s string) (string, string) {
	parts := strings.Split(s, ".")
	return parts[0], parts[1]
}
