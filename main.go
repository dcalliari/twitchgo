package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/joho/godotenv"

	"twitchgo/handlers"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := godotenv.Load(); err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	nick := os.Getenv("TWITCH_NICK")
	oauth := os.Getenv("TWITCH_OAUTH")
	channel := os.Getenv("TWITCH_CHANNEL")
	prefix := os.Getenv("PREFIX")

	if nick == "" || oauth == "" || channel == "" {
		log.Fatal("Variáveis de ambiente estão faltando")
	}

	client := twitch.NewClient(nick, oauth)

	client.OnConnect(func() {
		log.Printf("✅ Conectado como %s ao canal %s", nick, channel)
	})

	client.OnPrivateMessage(func(msg twitch.PrivateMessage) {
		handlers.OnMessage(client, msg, prefix)
	})

	client.Join(channel)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := client.Connect(); err != nil {
			log.Fatal("Erro ao conectar:", err)
		}
	}()

	<-quit
	log.Println("🛑 Finalizando conexão com a Twitch...")
	client.Disconnect()
}
