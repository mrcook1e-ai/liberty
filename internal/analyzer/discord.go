package analyzer

import (
	"encoding/binary"
	"net"
	"net/http"
	"time"
)

// CheckDiscordTCP проверяет доступность API/Gateway (Интерфейс)
func CheckDiscordTCP() (bool, time.Duration) {
	client := http.Client{Timeout: 3 * time.Second}
	start := time.Now()
	resp, err := client.Get("https://gateway.discord.gg")
	duration := time.Since(start)

	if err != nil {
		return false, duration
	}
	defer resp.Body.Close()
	return true, duration
}

// CheckDiscordUpdates проверяет доступность CDN (для картинок и обновлений)
func CheckDiscordUpdates() (bool, time.Duration) {
	// Пробуем загрузить маленькую иконку с CDN
	url := "https://cdn.discordapp.com/embed/avatars/0.png"
	client := http.Client{Timeout: 5 * time.Second}
	
	start := time.Now()
	resp, err := client.Get(url)
	duration := time.Since(start)

	if err != nil {
		return false, duration
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, duration
}

// ProbeDiscordUDP (Оставляем для ручных тестов, но в пайплайне отключим)
func ProbeDiscordUDP(targetIP string, startPort, endPort int) (bool, time.Duration) {
	ips := []string{"66.22.196.2", "51.159.165.214", "162.159.129.233"} 
	if targetIP != "" { ips = []string{targetIP} }

	for _, ip := range ips {
		ports := []int{50000, 50005, 55000}
		for _, port := range ports {
			success, lat := sendDiscordDiscoveryPacket(ip, port, 1000)
			if success { return true, lat }
		}
	}
	return false, 0
}

func sendDiscordDiscoveryPacket(ip string, port int, timeoutMs int) (bool, time.Duration) {
	addr := net.UDPAddr{IP: net.ParseIP(ip), Port: port}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil { return false, 0 }
	defer conn.Close()

	packet := make([]byte, 70)
	binary.BigEndian.PutUint16(packet[0:2], 1)
	binary.BigEndian.PutUint16(packet[2:4], 70)
	binary.BigEndian.PutUint32(packet[4:8], uint32(time.Now().UnixNano()))

	start := time.Now()
	conn.SetDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	if _, err := conn.Write(packet); err != nil { return false, 0 }

	recvBuf := make([]byte, 1024)
	if _, err := conn.Read(recvBuf); err != nil { return false, time.Since(start) }
	return true, time.Since(start)
}

func AnalyzeDiscordFull(voiceIP string) string { return "Analysis disabled" }