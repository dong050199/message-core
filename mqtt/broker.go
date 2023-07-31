package mqtt

import (
	"log"
	hook "message-core/custom-hook"
	"os"
	"os/signal"
	"syscall"

	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/listeners"
)

func InstanceMQTTBroker() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	server := mqtt.New(nil)
	server.Options.Capabilities.Compatibilities.ObscureNotAuthorized = true
	server.Options.Capabilities.Compatibilities.PassiveClientDisconnect = true

	// _ = server.AddHook(new(auth.AllowHook), nil)

	err := server.AddHook(new(hook.CustomHook), map[string]any{})
	if err != nil {
		log.Fatal(err)
	}

	tcp := listeners.NewTCP("t1", ":1883", nil)
	err = server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	// Start the server
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-done
	server.Log.Warn().Msg("caught signal, stopping...")
	server.Close()
	server.Log.Info().Msg("main.go finished")
}
