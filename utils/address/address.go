package address

import (
	"fmt"
	"strconv"
	"strings"
)

func GetGRPCAddress(hostandport string) string {
	// host:port
	ss := strings.Split(hostandport, ":")
	if len(ss) != 2 {
		return ""
	}

	port, _ := strconv.Atoi(ss[1])

	// Add 200 to get the grpc port
	port += 200

	return fmt.Sprintf("%s:%d", ss[0], port)
}
