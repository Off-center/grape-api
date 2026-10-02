package pkg

import "net"

func RemoteIPAddress(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)

	if err != nil {
		return remoteAddr
	}

	return host
}
