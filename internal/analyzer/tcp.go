package analyzer

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CheckHTTP выполняет запрос и возвращает статус + описание ошибки (если есть)
func CheckHTTP(url string, timeoutMs int) (bool, string, time.Duration) {
	client := http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}

	start := time.Now()
	resp, err := client.Get(url)
	duration := time.Since(start)

	if err != nil {
		errStr := err.Error()
		// Анализируем ошибку на предмет сброса соединения (RST)
		if strings.Contains(errStr, "connection reset") || strings.Contains(errStr, "forcibly closed") {
			return false, "RST (Blocked)", duration
		}
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
			return false, "Timeout", duration
		}
		return false, "Error", duration
	}
	defer resp.Body.Close()

	if resp.StatusCode > 0 {
		return true, "OK", duration
	}
	return false, fmt.Sprintf("Code %d", resp.StatusCode), duration
}

func CheckYouTube() (bool, time.Duration) {
	ok, status, d := CheckHTTP("https://www.youtube.com", 2000)
	if !ok {
		// Если сайт недоступен, нет смысла проверять редиректор
		fmt.Printf("(Web: %s) ", status)
		return false, d
	}
	okVid, statusVid, dVid := CheckHTTP("https://redirector.googlevideo.com/generate_204", 2000)
	if !okVid {
		fmt.Printf("(Vid: %s) ", statusVid)
	}
	return ok && okVid, d + dVid
}

func CheckCloudflare() (bool, time.Duration) {
	ok, _, d := CheckHTTP("https://www.cloudflare.com", 2000)
	return ok, d
}

// Специализированная проверка для Discord Media (WSS)
func CheckDiscordMediaWSS() (bool, string, time.Duration) {
	url := "https://finland10164.discord.media"
	
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.9219 Chrome/138.0.7204.251 Electron/37.6.0 Safari/537.36")
	req.Header.Set("Origin", "https://discord.com")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	client := http.Client{
		Timeout: 2000 * time.Millisecond,
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "connection reset") {
			return false, "RST", duration
		}
		return false, "Err", duration
	}
	defer resp.Body.Close()
	return resp.StatusCode > 0, "OK", duration
}
