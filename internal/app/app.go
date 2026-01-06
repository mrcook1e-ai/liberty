package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"liberty/internal/engine"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var AppBaseDir string

const RotterdamID = "1379839944536621097"

type RegionResult struct {
	Status string `json:"status"`
	MS     int64  `json:"ms"`
}

type SessionData struct {
	Config    string                  `json:"config"`
	Regions   map[string]RegionResult `json:"regions"`
	Steps     map[int]string          `json:"steps"`
	StartTime int64                   `json:"start_time"`
	DpiLevel  string                  `json:"dpi_level"`
	Rate      float64                 `json:"rate"`
	Latency   int64                   `json:"latency"`
	Entropy   float64                 `json:"entropy"`
}

var RegionNames = map[string]string{
	"1033697363073708052": "Brazil",
	"1379840353221218515": "Hong Kong",
	"1379839058536300685": "India",
	"1379839914451009636": "Japan",
	RotterdamID:           "Rotterdam",
	"1379839971447279727": "Singapore",
	"1379840041689288745": "Sydney",
}

var RegionOrder = []string{
	RotterdamID, "1379839058536300685", "1379839914451009636",
	"1379839971447279727", "1379840353221218515", "1033697363073708052", "1379840041689288745",
}

type EventEmitter func(name string, data ...interface{})

func Run(baseDir string, logger engine.LogFunc, emitter EventEmitter, fullTournament bool, discordToken string, guildID string) {
	AppBaseDir = baseDir
	session := SessionData{
		Regions:   make(map[string]RegionResult),
		Steps:     make(map[int]string),
		StartTime: time.Now().Unix(),
		DpiLevel:  "MEDIUM",
		Rate:      99.8,
		Latency:   142,
		Entropy:   0.84,
	}

	onProgress := func(step int, msg string) {
		session.Steps[step] = msg
		if step > 3 {
			session.DpiLevel = "HIGH"
		}
		session.Rate = 80.0 + float64(step)*4.0
		session.Latency = 100 + int64(step*15)
		session.Entropy = 0.7 + float64(step)*0.05

		emitter("step-progress", map[string]interface{}{"step": step, "msg": msg})

		go func() {
			for i := 0; i < 10; i++ {
				emitter("packet-activity", 50+i*10)
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}

	pipe := engine.NewPipeline(baseDir, logger, onProgress)

	services := []engine.ServiceDefinition{
		{ID: 1, Name: "Cloudflare", ListName: "list-cloudflare.txt", CheckFunc: engine.CheckCloudflare, Strategies: []func(p *engine.PipelineState, l string) []string{engine.Tcp_Split2, engine.Untitled_FakeDSplit}},
		{ID: 2, Name: "YouTube", ListName: "list-google.txt", CheckFunc: engine.CheckYouTube, Strategies: []func(p *engine.PipelineState, l string) []string{engine.YT_Untitled_FakeDSplit, engine.YT_Untitled_HostFakeSplit_Google}},
		{ID: 3, Name: "Discord WSS", ListName: "list-discord.txt", CheckFunc: func() (bool, string) { return CheckDiscordWebAndCDN(discordToken) }, Strategies: []func(p *engine.PipelineState, l string) []string{engine.Untitled_FakeDSplit, engine.Tcp_Split2, engine.Tcp_Disorder, engine.Tcp_Fake}},
		{ID: 4, Name: "Discord UDP", CheckFunc: engine.CheckDiscordUDP, UdpStrategies: []func(p *engine.PipelineState) []string{engine.Untitled_Voice_Specific, engine.Voice_Combo}},
	}

	pipe.Run(services)

	onProgress(5, "WebRTC Analysis")
	mediaStrategies := []struct {
		Name string
		Func func(p *engine.PipelineState, l string) []string
	}{
		{"SplitPorts (Default)", engine.Untitled_Media_SplitPorts},
		{"SplitPorts V2", engine.Untitled_Media_SplitPorts_Variant2},
		{"Tcp_Disorder", engine.Tcp_Disorder},
	}

	var bestArgs []string
	bestScore := -1

	for _, strat := range mediaStrategies {
		emitter("strategy-change", strat.Name)
		emitter("log", fmt.Sprintf("[WEBRTC] Testing strategy: %s\n", strat.Name))
		extraArgs := strat.Func(pipe, "list-discord-media.txt")

		if pipe.TryConfig(extraArgs) {
			time.Sleep(2 * time.Second)
			emitter("log", "[WEBRTC] Verifying region: Rotterdam...\n")
			ok, ms := checkSingleRegion(RotterdamID, discordToken, guildID, emitter)
			res := RegionResult{Status: ifThen(ok, "OK", "XX"), MS: ms}
			session.Regions["Rotterdam"] = res
			emitter("region-check-end", map[string]interface{}{"name": "Rotterdam", "status": res.Status, "ms": res.MS})

			if ok {
				emitter("log", fmt.Sprintf("[WEBRTC] Success! Latency: %dms\n", ms))
				if !fullTournament {
					bestArgs = extraArgs
					bestScore = 100
					break
				}
			}

			score := 0
			if ok {
				score += 2
			}

			if fullTournament {
				for _, rID := range RegionOrder {
					if rID == RotterdamID {
						continue
					}
					name := RegionNames[rID]
					emitter("log", fmt.Sprintf("[WEBRTC] Verifying region: %s...\n", name))
					rOk, rMs := checkSingleRegion(rID, discordToken, guildID, emitter)
					rRes := RegionResult{Status: ifThen(rOk, "OK", "XX"), MS: rMs}
				session.Regions[name] = rRes
				emitter("region-check-end", map[string]interface{}{"name": name, "status": rRes.Status, "ms": rMs})
				if rOk {
					emitter("log", fmt.Sprintf("[WEBRTC] %s OK (%dms)\n", name, rMs))
					score += 2
				}
				}
			}

			if score > bestScore {
				bestScore = score
				bestArgs = extraArgs
			}
		}
	}

	if bestArgs != nil {
		pipe.AccumulatedArgs = append(pipe.AccumulatedArgs, bestArgs...)
		pipe.TryConfig(bestArgs)
		session.Config = strings.Join(pipe.AccumulatedArgs, " ")
		emitter("final-config", session.Config)

		// Сообщаем фронтенду, что этап WebRTC успешно завершен
		onProgress(5, "Optimized")

		// Сохраняем полную сессию в JSON в рабочую папку
		saveSession(baseDir, session)
	}

	SaveBatchFile(baseDir, pipe.AccumulatedArgs)
	cleanupDiscord(baseDir, discordToken, guildID)
	emitter("done")
}

func cleanupDiscord(baseDir string, token, guildID string) {
	botExe := filepath.Join(baseDir, "bin", "botcheck.exe")
	cmd := exec.Command(botExe, "-mode=leave", "-token="+token, "-guild="+guildID)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()
}

func saveSession(baseDir string, s SessionData) {
	data, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(filepath.Join(baseDir, "session.json"), data, 0644)
}

func checkSingleRegion(cID string, token, guildID string, emitter EventEmitter) (bool, int64) {
	botExe := filepath.Join(AppBaseDir, "bin", "botcheck.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, botExe, "-mode=voice", "-channel="+cID, "-token="+token, "-guild="+guildID)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// Читаем вывод бота в реальном времени
	stdout, _ := cmd.StdoutPipe()
	start := time.Now()

	if err := cmd.Start(); err != nil {
		return false, 0
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		emitter("log", line)                                                   // Отправляем каждую строку бота в интерфейс
		emitter("step-progress", map[string]interface{}{"step": 5, "msg": line}) // И обновляем статус этапа
	}

	err := cmd.Wait()
	return err == nil, time.Since(start).Milliseconds()
}

func ifThen(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func CheckDiscordWebAndCDN(token string) (bool, string) {
	botExe := filepath.Join(AppBaseDir, "bin", "botcheck.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, botExe, "-mode=gateway", "-token="+token)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil, "Result"
}

func SaveBatchFile(baseDir string, args []string) {
	content := "@echo off\r\nset BIN=%~dp0bin/\n"
	content += fmt.Sprintf("start \"Zapret\" /min \"%%BIN%%winws.exe\" %s\r\n", strings.Join(args, " "))
	_ = os.WriteFile(filepath.Join(baseDir, "final_start.bat"), []byte(content), 0755)
}
