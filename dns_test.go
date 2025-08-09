package uri

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestDNSvsHost(t *testing.T) {
	for _, scheme := range schemesWithDNS() {
		require.Truef(t, UsesDNSHostValidation(scheme), "expected scheme %q to use Internet Domain Names", scheme)
	}

	require.False(t, UsesDNSHostValidation("phone"))
}

func TestValidateHostForScheme(t *testing.T) {
	for _, host := range []string{
		"a.b.c",
		"a",
		"a.b1b",
		"a.b2",
		"a.1b",
		"a.b.c.d",
		"a-b.c-d",
		"www.詹姆斯.org",
		"www.詹-姆斯.org",
		fmt.Sprintf("a.%s.c", strings.Repeat("b", 63)),
		"a%2Eb%2ec.d",
		"a.b.c.d%30",
		"a.b.c.%30d",
		"a.b.c.%55",
	} {
		require.NoErrorf(t, validateHostForScheme(host, "http"),
			"expected host %q to validate",
			host,
		)
	}

	for _, host := range []string{
		"a.b.c|",
		"a.b.c-",
		"a-",
		"a.",
		"a.b.",
		"a.2",
		"a.b.c..",
		".",
		"",
		"www.詹姆斯.org/",
		"www.詹{姆}斯.org/",
		fmt.Sprintf("a.%s.c", strings.Repeat("b", 64)),
		fmt.Sprintf("a.%sb.c", string([]rune{utf8.RuneError})),
		fmt.Sprintf("%sa.b.c", string([]rune{utf8.RuneError})), //nolint:perfsprint // we prefer consistency with the above lines over performance
		".a.b.c",
		"a.b.c.d%2b",
		"a.b.c.%",
		"a.b.c.%X",
		"%",
		"%X",
	} {
		require.Errorf(t, validateHostForScheme(host, "http"),
			"expected host %q NOT to validate",
			host,
		)
	}
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
		"postgresql",
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
