package main

import (
	"context"
	"encoding/json"
	"fmt"
	"liberty/internal/app"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App { return &App{} }
func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// GetISPInfo возвращает детальную информацию о соединении в формате JSON
func (a *App) GetISPInfo() string {
	client := &http.Client{Timeout: 4 * time.Second}
	// Запрашиваем провайдера, AS, город, код страны и IP (query)
	resp, err := client.Get("http://ip-api.com/json/?fields=isp,as,city,countryCode,query")
	if err != nil {
		return "{}"
	}
	defer resp.Body.Close()

	var res struct {
		ISP         string `json:"isp"`
		AS          string `json:"as"`
		City        string `json:"city"`
		CountryCode string `json:"countryCode"`
		Query       string `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "{}"
	}
	
	// Возвращаем как JSON-строку для фронтенда
	out, _ := json.Marshal(res)
	return string(out)
}

// SetFinlandFix прописывает или удаляет финские сервера в hosts
func (a *App) SetFinlandFix(active bool) string {
	hostsPath := `C:\Windows\System32\drivers\etc\hosts`
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		return "Error reading hosts"
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inBlock := false

	// Очищаем старые записи Liberty
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "# LIBERTY FINLAND FIX START" {
			inBlock = true
			continue
		}
		if trimmed == "# LIBERTY FINLAND FIX END" {
			inBlock = false
			continue
		}
		if !inBlock {
			newLines = append(newLines, line)
		}
	}

	if active {
		newLines = append(newLines, "# LIBERTY FINLAND FIX START")
		for i := 10000; i <= 10199; i++ {
			newLines = append(newLines, fmt.Sprintf("104.25.158.178 finland%d.discord.media", i))
		}
		newLines = append(newLines, "# LIBERTY FINLAND FIX END")
	}

	output := strings.Join(newLines, "\n")
	err = os.WriteFile(hostsPath, []byte(output), 0644)
	if err != nil {
		return "Error writing hosts: " + err.Error()
	}

	return "OK"
}

// OpenAppData открывает папку приложения в проводнике Windows
func (a *App) OpenAppData() {
	// Путь по умолчанию для Wails приложений на Windows
	home, _ := os.UserHomeDir()
	appPath := filepath.Join(home, "AppData", "Roaming", "liberty")
	
	// Создаем папку, если её нет (на всякий случай)
	os.MkdirAll(appPath, 0755)
	
	// Открываем через проводник
	exec.Command("explorer", appPath).Run()
}

type AppSettings struct {
	DiscordToken   string `json:"discord_token"`
	DiscordGuild   string `json:"discord_guild"`
	DiscordChannel string `json:"discord_channel"`
	WorkDir        string `json:"work_dir"`
}

func (a *App) getSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "AppData", "Roaming", "liberty", "settings.json")
}

func (a *App) GetSettings() AppSettings {
	path := a.getSettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return AppSettings{WorkDir: ""}
	}
	var settings AppSettings
	json.Unmarshal(data, &settings)
	return settings
}

func (a *App) SaveSettings(settings AppSettings) string {
	path := a.getSettingsPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.MarshalIndent(settings, "", "  ")
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return err.Error()
	}
	return "OK"
}

func (a *App) SelectFolder() string {
	res, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Working Directory",
	})
	if err != nil {
		return ""
	}
	return res
}

func (a *App) Start(fullTournament bool) {
	settings := a.GetSettings()
	go func() {
		workDir := settings.WorkDir
		if workDir == "" {
			workDir, _ = os.MkdirTemp("", "liberty_gui_")
		} else {
			os.MkdirAll(workDir, 0755)
		}

		if err := unpack(workDir); err != nil {
			runtime.EventsEmit(a.ctx, "log", "[ERR] Unpack failed: "+err.Error())
			return
		}

		guiLogger := func(msg string) { runtime.EventsEmit(a.ctx, "log", msg) }
		app.Run(workDir, guiLogger, func(name string, data ...interface{}) {
			runtime.EventsEmit(a.ctx, name, data...)
		}, fullTournament, settings.DiscordToken, settings.DiscordGuild, settings.DiscordChannel)
	}()
}

// GetSessionData возвращает всю сохраненную информацию
func (a *App) GetSessionData() string {
	settings := a.GetSettings()
	workDir := settings.WorkDir
	if workDir == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(workDir, "session.json"))
	if err != nil {
		return ""
	}
	return string(data)
}

func (a *App) Reset() {
	settings := a.GetSettings()
	if settings.WorkDir != "" {
		os.Remove(filepath.Join(settings.WorkDir, "session.json"))
		os.Remove(filepath.Join(settings.WorkDir, "last_config.txt"))
	}
	a.Stop()
	runtime.EventsEmit(a.ctx, "log", "[SYSTEM] Engine reset.\n")
}

func (a *App) IsWinwsRunning() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq winws.exe", "/NH")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, _ := cmd.Output()
	if strings.Contains(strings.ToLower(string(out)), "winws.exe") { return true }
	cmdSvc := exec.Command("sc", "query", "zapret")
	cmdSvc.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	outSvc, _ := cmdSvc.Output()
	return strings.Contains(string(outSvc), "RUNNING")
}

func (a *App) Stop() {
	cmd1 := exec.Command("taskkill", "/F", "/IM", "winws.exe")
	cmd1.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd1.Run()

	cmd2 := exec.Command("net", "stop", "zapret")
	cmd2.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd2.Run()
	
	runtime.EventsEmit(a.ctx, "status", "IDLE")
}

func (a *App) GetBuildDetails(rawConfig string) map[string]interface{} {
	details := make(map[string]interface{})
	if strings.Contains(rawConfig, "--dpi-desync=split2") {
		details["Method"] = "Split2"
		details["Frag"] = "Seq Ovl"
	} else if strings.Contains(rawConfig, "fakedsplit") {
		details["Method"] = "FakeDSplit"
		details["Payload"] = "TLS Client"
	}
	if strings.Contains(rawConfig, "ts") { details["Fooling"] = "Timestamps" }
	return details
}
