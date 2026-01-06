package engine

import (
	"fmt"
	"liberty/internal/analyzer"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type LogFunc func(string)
type ProgressFunc func(step int, msg string)

type PipelineState struct {
	AccumulatedArgs []string
	WinwsPath       string
	BinDir          string
	ListsDir        string
	Logger          LogFunc
	OnProgress      ProgressFunc // Новый callback для прогресса
}

type StrategyFunc func(p *PipelineState) []string

type ServiceDefinition struct {
	ID         int
	Name       string
	ListName   string
	CheckFunc  func() (bool, string)
	Strategies []func(p *PipelineState, list string) []string
	UdpStrategies []func(p *PipelineState) []string
}

func NewPipeline(baseDir string, logger LogFunc, progress ProgressFunc) *PipelineState {
	binDir := baseDir
	listsDir := filepath.Join(baseDir, "lists")
	winws := filepath.Join(binDir, "winws.exe")

	return &PipelineState{
		AccumulatedArgs: []string{
			"--wf-tcp=80,443,2053,2083,2087,2096,8443",
			"--wf-udp=443,19294-19344,50000-50100",
		},
		WinwsPath: winws,
		BinDir:    binDir,
		ListsDir:  listsDir,
		Logger:    logger,
		OnProgress: progress,
	}
}

func (p *PipelineState) Log(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if p.Logger != nil { p.Logger(msg) }
}

func (p *PipelineState) Run(services []ServiceDefinition) {
	killWinws()
	for _, service := range services {
		p.Log("\n>>> Unlocking: %d. %s\n", service.ID, service.Name)
		success := false
		
		if service.ListName != "" {
			for i, strat := range service.Strategies {
				p.OnProgress(service.ID, fmt.Sprintf("Testing Strategy %d/%d...", i+1, len(service.Strategies)))
				
				extraArgs := strat(p, service.ListName)
				if p.TryConfig(extraArgs) {
					ok, status := service.CheckFunc()
					if ok {
						p.Log("[SUCCESS]\n")
						p.OnProgress(service.ID, "Optimized")
						p.AccumulatedArgs = append(p.AccumulatedArgs, extraArgs...)
						success = true
						break
					} else {
						p.Log("[FAIL: %s]\n", status)
						p.OnProgress(service.ID, fmt.Sprintf("Fail: %s", status))
					}
				}
			}
		} else if len(service.UdpStrategies) > 0 {
			for i, strat := range service.UdpStrategies {
				p.OnProgress(service.ID, fmt.Sprintf("Applying UDP Strategy %d/%d...", i+1, len(service.UdpStrategies)))
				extraArgs := strat(p)
				if p.TryConfig(extraArgs) {
					p.OnProgress(service.ID, "Applied")
					p.AccumulatedArgs = append(p.AccumulatedArgs, extraArgs...)
					success = true
					break
				}
			}
		}

		if !success {
			p.OnProgress(service.ID, "Failed to Unlock")
		}
	}
}

func (p *PipelineState) TryConfig(extraArgs []string) bool {
	killWinws()
	fullArgs := append([]string{}, p.AccumulatedArgs...)
	fullArgs = append(fullArgs, extraArgs...)

	cmd := exec.Command(p.WinwsPath, fullArgs...)
	cmd.Dir = p.BinDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	if err := cmd.Start(); err != nil { return false }
	time.Sleep(1500 * time.Millisecond)
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() { return false }
	return true
}

func killWinws() { 
	cmd := exec.Command("taskkill", "/F", "/IM", "winws.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run() 
}
func (p *PipelineState) GetList(name string) string { return filepath.Join(p.ListsDir, name) }
func (p *PipelineState) GetBin(name string) string { return filepath.Join(p.BinDir, name) }

func CheckYouTube() (bool, string) { 
	ok, _ := analyzer.CheckYouTube()
	return ok, "YouTube Check"
}
func CheckCloudflare() (bool, string) { 
	ok, _ := analyzer.CheckCloudflare() 
	return ok, "Cloudflare Check"
}
func CheckDiscordTCP() (bool, string) { 
	ok, _ := analyzer.CheckDiscordTCP()
	return ok, "Gateway Check"
}
func CheckDiscordUpdates() (bool, string) {
	ok, _ := analyzer.CheckDiscordUpdates()
	return ok, "CDN Check"
}
func CheckDiscordUDP() (bool, string) { return true, "UDP Ready" }
