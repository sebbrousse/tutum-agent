package agent

import (
	"github.com/getsentry/raven-go"
)

var sentryClient *raven.Client = nil
var DSN string = "https://af8f9402046b434781f9e09ab6da4b92:81139d3c30b64df8ae1601374e2ed698@app.getsentry.com/40515"

func getSentryClient() *raven.Client {
	if sentryClient == nil {
		client, err := raven.NewClient(DSN, nil)
		if err != nil {
			Logger.Println(err)
		}
		sentryClient = client
	}
	return sentryClient
}

func SendError(err error, msg string, extra map[string]interface{}) {
	go func() {
		client := getSentryClient()
		packet := &raven.Packet{Message: msg, Interfaces: []raven.Interface{raven.NewException(err, raven.NewStacktrace(0, 5, nil))}}
		if extra != nil {
			packet.Extra = extra
		}
		_, ch := client.Capture(packet, nil)
		if sendErr := <-ch; sendErr != nil {
			Logger.Println(sendErr)
		} else {
			Logger.Println("sent error to sentry successfully")
		}
	}()
}
