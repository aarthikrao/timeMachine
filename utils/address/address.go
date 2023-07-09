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
