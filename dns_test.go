package uri

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDNSvsHost(t *testing.T) {
	for _, scheme := range schemesWithDNS() {
		require.True(t, UsesDNSHostValidation(scheme))
	}

	require.False(t, UsesDNSHostValidation("phone"))
}

func schemesWithDNS() []string {
	return []string{
		"dns",
		"dntp",
		"finger",
		"ftp",
		"git",
		"http",
		"https",
		"imap",
		"irc",
		"jms",
		"mailto",
		"nfs",
		"nntp",
		"ntp",
		"postgres",
		"redis",
		"rmi",
		"rtsp",
		"rsync",
		"sftp",
		"skype",
		"smtp",
		"snmp",
		"soap",
		"ssh",
		"steam",
		"svn",
		"tcp",
		"telnet",
		"udp",
		"vnc",
		"wais",
		"ws",
		"wss",
	}
}
