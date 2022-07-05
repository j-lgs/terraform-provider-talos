package talos

import (
	"net"
	"os"
	"strconv"
	"time"
)

var (
	actionsTimeBuffer = 60 * time.Second
	isFullyResetWait  = 50 * time.Second
)

func lookupEnvBool(key string) (result bool, err error) {
	val, set := os.LookupEnv("TF_ACC")
	if set {
		result, err = strconv.ParseBool(val)
		if err != nil {
			return
		}
	}

	return
}

func qemuReset(conn net.Conn) error {
	// This approach is not ideal is it might take much more or much less time for a talos host to
	// reset. Ideally there would be an insecure endpoint that can be checked to determine if a host
	// is up. Likely it would return 200 if up. This is handy too as it can help the provider determine
	// whether the host's networking stack is up.
	time.Sleep(isFullyResetWait)

	// Require more time if inside a Github Action
	isGithubAction, err := lookupEnvBool("GITHUB_ACTIONS")
	if err != nil {
		return err
	}

	if isGithubAction {
		time.Sleep(actionsTimeBuffer)
	}

	buf := make([]byte, 256)
	if n, err := conn.Read(buf); n <= 0 || err != nil {
		return err
	}

	conn.Write([]byte(`{"execute": "qmp_capabilities"}`))
	if n, err := conn.Read(buf); n <= 0 || err != nil {
		return err
	}

	conn.Write([]byte(`{"execute": "system_reset"}`))
	if n, err := conn.Read(buf); n <= 0 || err != nil {
		return err
	}

	return nil
}
