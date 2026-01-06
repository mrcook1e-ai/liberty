package main

import (
	"fmt"
	"liberty/internal/analyzer"
)

func main() {
	// IP адрес одного из голосовых серверов Discord (Rotterdam).
	// В идеале нужно резолвить 'russia.discord.gg' или подобные, но они за балансировщиком.
	// Прямой IP надежнее для теста протокола.
	targetIP := "66.22.196.2" 

	fmt.Println(analyzer.AnalyzeDiscordFull(targetIP))
	
	fmt.Println("\n=== General Checks ===")
	// YouTube Check
	ok, lat := analyzer.CheckYouTube()
	if ok {
		fmt.Printf("YouTube (GGC): OK (%v)\n", lat)
	} else {
		fmt.Printf("YouTube (GGC): FAIL\n")
	}
}