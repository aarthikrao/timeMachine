package address

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aarthikrao/timeMachine/utils/constants"
)

func GetGRPCAddress(hostandport string) string {
	// host:port
	ss := strings.Split(hostandport, ":")
	if len(ss) != 2 {
		return ""
	}

	port, _ := strconv.Atoi(ss[1])

	port += constants.GRPCPortAdd

	return fmt.Sprintf("%s:%d", ss[0], port)
}

// Returns the HTTP address from the raft address.
// HTTP port = Raft port - 100
func GetHTTPAddressFromRaft(hostandport string) string {
	// host:port
	ss := strings.Split(hostandport, ":")
	if len(ss) != 2 {
		return ""
	}
	port, _ := strconv.Atoi(ss[1])
	port -= 100
	return fmt.Sprintf("%s:%d", ss[0], port)
}
