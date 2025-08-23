package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/joho/godotenv"

	"twitchgo/commands"
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

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := commands.SavePointsData(); err != nil {
				log.Printf("Error saving points data: %v", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := client.Connect(); err != nil {
			log.Fatal("Erro ao conectar:", err)
		}
	}()

	<-quit
	log.Println("🛑 Finalizando conexão com a Twitch...")

	if err := commands.SavePointsData(); err != nil {
		log.Printf("Error saving points data on shutdown: %v", err)
	}

	client.Disconnect()
}
