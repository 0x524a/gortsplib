package base

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustParseURL(s string) *URL {
	u, err := ParseURL(s)
	if err != nil {
		panic(err)
	}
	return u
}

func TestParseURL(t *testing.T) {
	for _, ca := range []struct {
		name string
		enc  string
		u    *URL
	}{
		{
			"ipv6 stateless",
			`rtsp://user:pa%23ss@[fe80::a8f4:3219:f33e:a072%wl0]:8554/prox%23ied`,
			&URL{
				Scheme: "rtsp",
				Host:   "[fe80::a8f4:3219:f33e:a072%wl0]:8554",
				Path:   "/prox#ied",
				User:   url.UserPassword("user", "pa#ss"),
			},
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			u, err := ParseURL(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.u, u)
		})
	}
}

func TestURLParseErrors(t *testing.T) {
	for _, ca := range []struct {
		name string
		enc  string
		err  string
	}{
		{
			"invalid",
			":testing",
			"parse \":testing\": missing protocol scheme",
		},
		{
			"unsupported scheme",
			"http://testing",
			"unsupported scheme 'http'",
		},
		{
			"with opaque data",
			"rtsp:opaque?query",
			"URLs with opaque data are not supported",
		},
		{
			"with fragment",
			"rtsp://localhost:8554/teststream#fragment",
			"URLs with fragments are not supported",
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			_, err := ParseURL(ca.enc)
			require.EqualError(t, err, ca.err)
		})
	}
}

func TestURLStringUserinfoSpecialChars(t *testing.T) {
	// ! is a valid sub-delimiter (RFC 3986 §3.2.1) and must not be percent-encoded.
	u := mustParseURL("rtsp://rtspuser:CorpSec123$!@host:554/path")
	require.Equal(t, "rtsp://rtspuser:CorpSec123$!@host:554/path", u.String())
}

func TestURLStringKeepsEncodedAtHashColon(t *testing.T) {
	// Passwords with @, #, $ must remain percent-encoded in the URL string
	// so that re-parsing the string does not break the URL structure.
	u := mustParseURL("rtsp://service:Service.!%40%23%2434@192.168.1.153:8554/stream")
	s := u.String()

	// ! and $ are unescaped (valid sub-delimiters in userinfo)
	// %40 (@), %23 (#) must stay encoded to avoid breaking re-parse
	require.Equal(t, "rtsp://service:Service.!%40%23$34@192.168.1.153:8554/stream", s)

	// The string must be re-parseable without error
	u2, err := ParseURL(s)
	require.NoError(t, err)
	require.Equal(t, "192.168.1.153:8554", u2.Host)
}

func TestURLClone(t *testing.T) {
	u := mustParseURL("rtsp://localhost:8554/test/stream")
	u2 := u.Clone()
	u.Host = "otherhost"

	require.Equal(t, &URL{
		Scheme: "rtsp",
		Host:   "otherhost",
		Path:   "/test/stream",
	}, u)

	require.Equal(t, &URL{
		Scheme: "rtsp",
		Host:   "localhost:8554",
		Path:   "/test/stream",
	}, u2)
}

func TestURLCloneWithoutCredentials(t *testing.T) {
	u := mustParseURL("rtsp://user:pass@localhost:8554/test/stream")
	u2 := u.CloneWithoutCredentials()
	u.Host = "otherhost"

	require.Equal(t, &URL{
		Scheme: "rtsp",
		Host:   "otherhost",
		Path:   "/test/stream",
		User:   url.UserPassword("user", "pass"),
	}, u)

	require.Equal(t, &URL{
		Scheme: "rtsp",
		Host:   "localhost:8554",
		Path:   "/test/stream",
	}, u2)
}
