package proxy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/httpspec"
	"github.com/google/martian/v3/parse"
	"github.com/imranismail/bff/bfflog"
	"github.com/imranismail/bff/healthcheck"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	// martian built-in modifiers
	_ "github.com/google/martian/v3/body"
	_ "github.com/google/martian/v3/cookie"
	_ "github.com/google/martian/v3/failure"
	_ "github.com/google/martian/v3/fifo"
	_ "github.com/google/martian/v3/header"
	_ "github.com/google/martian/v3/martianurl"
	_ "github.com/google/martian/v3/method"
	_ "github.com/google/martian/v3/pingback"
	_ "github.com/google/martian/v3/port"
	_ "github.com/google/martian/v3/priority"
	_ "github.com/google/martian/v3/querystring"
	_ "github.com/google/martian/v3/skip"
	_ "github.com/google/martian/v3/stash"
	_ "github.com/google/martian/v3/static"
	_ "github.com/google/martian/v3/status"
	_ "github.com/imranismail/bff/bffmethod"
	_ "github.com/imranismail/bff/bffquerystring"
	_ "github.com/imranismail/bff/bffstatus"
	_ "github.com/imranismail/bff/bffurl"
	_ "github.com/imranismail/bff/body"
	"github.com/imranismail/bff/config"
	"github.com/imranismail/bff/log"
)

var Proxy *martian.Proxy

// Serve start the webserver
func Serve(cmd *cobra.Command, args []string) {
	proxy := martian.NewProxy()
	defer proxy.Close()

	Proxy = proxy

	if url, err := url.Parse(viper.GetString("url")); err != nil {
		proxy.SetDownstreamProxy(url)
	}

	proxy.SetRoundTripper(&http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: viper.GetBool("insecure"),
		},
	})

	configureProxy(proxy)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", viper.GetString("port")))

	if err != nil {
		log.Errorf("%s", err)
		os.Exit(1)
	}

	log.Infof("bff: starting proxy %s on %s", config.Version, listener.Addr().String())

	go proxy.Serve(listener)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)

	<-sigc

	log.Infof("bff: shutting down")
}

func Configure() {
	configureProxy(Proxy)
}

func configureProxy(*martian.Proxy) {
	outer, inner := httpspec.NewStack("bff")

	main := NewErrorBoundary()
	Proxy.SetRequestModifier(main)
	Proxy.SetResponseModifier(main)

	main.SetRequestModifier(outer)
	main.SetResponseModifier(outer)
	main.SetRequestVerifier(outer)
	main.SetResponseVerifier(outer)

	hcm := healthcheck.NewModifier(200)
	outer.AddRequestModifier(hcm)
	outer.AddResponseModifier(hcm)

	var modifiers []json.RawMessage

	err := yaml.Unmarshal([]byte(viper.GetString("modifiers")), &modifiers)

	if err != nil {
		log.Errorf("%s", err)
		os.Exit(1)
	}

	for _, mod := range modifiers {
		res, err := parse.FromJSON(mod)

		if err != nil {
			log.Errorf("%s", err)
			os.Exit(1)
		}

		reqmod := res.RequestModifier()

		if reqmod != nil {
			inner.AddRequestModifier(reqmod)
		}

		resmod := res.ResponseModifier()

		if resmod != nil {
			inner.AddResponseModifier(resmod)
		}
	}

	ml := bfflog.NewLogger()
	outer.AddRequestModifier(ml)
	outer.AddResponseModifier(ml)
}
