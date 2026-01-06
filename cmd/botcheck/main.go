package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	mode := flag.String("mode", "voice", "Check mode: 'gateway', 'voice' or 'leave'")
	channelID := flag.String("channel", "", "Override Channel ID")
	token := flag.String("token", "", "Discord Bot Token")
	guildID := flag.String("guild", "", "Discord Guild ID")
	flag.Parse()

	if *token == "" {
		fmt.Println("[ERR] Token is required")
		os.Exit(1)
	}

	dg, err := discordgo.New("Bot " + *token)
	if err != nil {
		fmt.Printf("[ERR] Session creation failed: %v\n", err)
		os.Exit(1)
	}
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates
	dg.LogLevel = discordgo.LogError

	fmt.Println("[INF] Connecting to Discord API...")
	err = dg.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR] API Error: %v\n", err)
		os.Exit(1)
	}
	defer dg.Close()

	if *mode == "leave" {
		if *guildID != "" {
			dg.ChannelVoiceJoin(*guildID, "", false, false)
		}
		time.Sleep(1 * time.Second)
		fmt.Println("[INF] Bot session terminated")
		os.Exit(0)
	}

	if *mode == "gateway" {
		os.Exit(0)
	}

	if *channelID == "" || *guildID == "" {
		fmt.Println("[ERR] Channel and Guild IDs are required for voice mode")
		os.Exit(1)
	}

	// Fetch Channel Name
	channel, err := dg.Channel(*channelID)
	if err == nil {
		fmt.Printf("[INF] Identified Channel: #%s\n", channel.Name)
	}

	success := verifyChannel(dg, *guildID, *channelID)
	if success {
		fmt.Println("[OK] Connection Established")
		os.Exit(0)
	} else {
		fmt.Println("[ERR] Handshake Failed")
		os.Exit(1)
	}
}

func verifyChannel(dg *discordgo.Session, gID, cID string) bool {
	fmt.Printf("[INF] Joining Voice: %s\n", cID)
	voiceConn, err := dg.ChannelVoiceJoin(gID, cID, false, true)
	if err != nil {
		fmt.Printf("[ERR] Join Error: %v\n", err)
		return false
	}
	
	ready := false
	for i := 0; i < 100; i++ {
		if voiceConn.Ready {
			fmt.Println("[OK] Linked to Voice Server")
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	voiceConn.Disconnect()
	time.Sleep(500 * time.Millisecond) 
	return ready
}
