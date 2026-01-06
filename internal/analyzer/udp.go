package analyzer

import (
	"encoding/binary"
	"net"
	"time"
)

// CheckDiscordUDP проверяет прохождение UDP пакетов до голосовых серверов Discord.
// Требует отправки специфичного пакета IP Discovery (70 байт).
// targetIP - один из известных IP голосовых серверов (например, из Роттердама или России).
func CheckDiscordUDP(targetIP string, port int, timeoutMs int) (bool, time.Duration) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(targetIP),
		Port: port,
	}

	// 1. Создаем подключение
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		return false, 0
	}
	defer conn.Close()

	// 2. Формируем пакет IP Discovery (70 байт)
	// Документация Discord (и реверс-инжиниринг):
	// Offset 0-1: Type (0x0001 Request)
	// Offset 2-3: Length (70)
	// Offset 4-8: SSRC (Любое число, обычно timestamp)
	// Остальное: Нули
	packet := make([]byte, 70)
	binary.BigEndian.PutUint16(packet[0:2], 1)  // Type 1
	binary.BigEndian.PutUint16(packet[2:4], 70) // Length 70
	binary.BigEndian.PutUint32(packet[4:8], uint32(time.Now().UnixNano())) // Random SSRC

	// 3. Отправка и ожидание
	start := time.Now()
	conn.SetDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))

	_, err = conn.Write(packet)
	if err != nil {
		return false, 0
	}

	// 4. Чтение ответа
	// Если DPI работает (дропает пакеты), мы получим timeout.
	// Если все ок, сервер пришлет ответ (тоже 70 байт с нашим внешним IP).
	recvBuf := make([]byte, 1024)
	_, err = conn.Read(recvBuf)
	duration := time.Since(start)

	if err != nil {
		// Timeout или ICMP Unreachable
		return false, duration
	}

	return true, duration
}
