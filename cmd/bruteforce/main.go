package main

import (
	"fmt"
	"liberty/internal/analyzer"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ConfigStrategy описывает набор аргументов для winws
type ConfigStrategy struct {
	Name string
	Args []string
}

func main() {
	// 1. Setup Environment
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	rootDir := filepath.Dir(ex)
	// Коррекция для go run (когда запускаемся из tmp)
	if _, err := os.Stat(filepath.Join(rootDir, "bin")); os.IsNotExist(err) {
		wd, _ := os.Getwd()
		rootDir = wd
	}

	binDir := filepath.Join(rootDir, "bin")
	listsDir := filepath.Join(rootDir, "lists")
	winwsPath := filepath.Join(binDir, "winws.exe")

	// Пути к ресурсам
	quicGoogle := filepath.Join(binDir, "quic_initial_www_google_com.bin")
	tlsGoogle := filepath.Join(binDir, "tls_clienthello_www_google_com.bin")
	tlsMaxRu := filepath.Join(binDir, "tls_clienthello_max_ru.bin")
	tls4pda := filepath.Join(binDir, "tls_clienthello_4pda_to.bin") // Некоторые конфиги используют это

	listGeneral := filepath.Join(listsDir, "list-general.txt")
	listGoogle := filepath.Join(listsDir, "list-google.txt")
	listExclude := filepath.Join(listsDir, "list-exclude.txt")
	ipsetExclude := filepath.Join(listsDir, "ipset-exclude.txt")

	// Очистка от старых процессов
	killWinws()

	// 2. Define Strategies (Парсинг из @Untitled-1.txt)
	strategies := []ConfigStrategy{
		{
			Name: "Strategy 1: Fake (Simple)",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=443", "--hostlist=" + listGeneral, "--hostlist-exclude=" + listExclude, "--ipset-exclude=" + ipsetExclude, "--dpi-desync=fake", "--dpi-desync-repeats=6", "--dpi-desync-fake-quic=" + quicGoogle, "--new",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=2053,2083,2087,2096,8443", "--hostlist-domains=discord.media", "--dpi-desync=fake,fakedsplit", "--dpi-desync-repeats=6", "--dpi-desync-fooling=ts", "--dpi-desync-fakedsplit-pattern=0x00", "--dpi-desync-fake-tls=" + tlsGoogle, "--new",
				"--filter-tcp=443", "--hostlist=" + listGoogle, "--ip-id=zero", "--dpi-desync=fake,fakedsplit", "--dpi-desync-repeats=6", "--dpi-desync-fooling=ts", "--dpi-desync-fakedsplit-pattern=0x00", "--dpi-desync-fake-tls=" + tlsGoogle, "--new",
				"--filter-tcp=80,443", "--hostlist=" + listGeneral, "--hostlist-exclude=" + listExclude, "--ipset-exclude=" + ipsetExclude, "--dpi-desync=fake,fakedsplit", "--dpi-desync-repeats=6", "--dpi-desync-fooling=ts", "--dpi-desync-fakedsplit-pattern=0x00", "--dpi-desync-fake-tls=" + tlsGoogle, "--new",
			},
		},
		{
			Name: "Strategy 2: Multisplit 652",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=443", "--hostlist=" + listGoogle, "--ip-id=zero", "--dpi-desync=multisplit", "--dpi-desync-split-seqovl=652", "--dpi-desync-split-pos=2", "--dpi-desync-split-seqovl-pattern=" + tlsGoogle, "--new",
				"--filter-tcp=80,443", "--hostlist=" + listGeneral, "--dpi-desync=multisplit", "--dpi-desync-split-seqovl=652", "--dpi-desync-split-pos=2", "--dpi-desync-split-seqovl-pattern=" + tlsGoogle, "--new",
			},
		},
		{
			Name: "Strategy 3: HostFakeSplit (ya.ru)",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=443", "--hostlist=" + listGoogle, "--ip-id=zero", "--dpi-desync=fake,hostfakesplit", "--dpi-desync-fake-tls-mod=rnd,dupsid,sni=www.google.com", "--dpi-desync-hostfakesplit-mod=host=www.google.com,altorder=1", "--dpi-desync-fooling=ts", "--new",
				"--filter-tcp=80,443", "--hostlist=" + listGeneral, "--dpi-desync=fake,hostfakesplit", "--dpi-desync-fake-tls-mod=rnd,dupsid,sni=ya.ru", "--dpi-desync-hostfakesplit-mod=host=ya.ru,altorder=1", "--dpi-desync-fooling=ts", "--new",
			},
		},
		{
			Name: "Strategy 4: Multisplit + Badseq (Increment 1000)",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=443", "--hostlist=" + listGoogle, "--dpi-desync=fake,multisplit", "--dpi-desync-repeats=6", "--dpi-desync-fooling=badseq", "--dpi-desync-badseq-increment=1000", "--dpi-desync-fake-tls=" + tlsGoogle, "--new",
			},
		},
		{
			Name: "Strategy 5: Syndata + Multidisorder (Aggressive)",
			Args: []string{
				"--wf-tcp=443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-l3=ipv4", "--filter-tcp=443,2053,2083,2087,2096,8443", "--dpi-desync=syndata,multidisorder", "--new",
			},
		},
		{
			Name: "Strategy 6: Multisplit 681 (Standard)",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=443", "--hostlist=" + listGoogle, "--dpi-desync=multisplit", "--dpi-desync-split-seqovl=681", "--dpi-desync-split-pos=1", "--dpi-desync-split-seqovl-pattern=" + tlsGoogle, "--new",
			},
		},
		{
			Name: "Strategy 7: Multisplit SNI Ext",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=443", "--hostlist=" + listGoogle, "--dpi-desync=multisplit", "--dpi-desync-split-pos=2,sniext+1", "--dpi-desync-split-seqovl=679", "--dpi-desync-split-seqovl-pattern=" + tlsGoogle, "--new",
			},
		},
		{
			Name: "Strategy 8: Fake TLS Mod None + Badseq 2",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=443", "--hostlist=" + listGoogle, "--dpi-desync=fake", "--dpi-desync-fake-tls-mod=none", "--dpi-desync-repeats=6", "--dpi-desync-fooling=badseq", "--dpi-desync-badseq-increment=2", "--new",
			},
		},
		{
			Name: "Strategy 9: HostFakeSplit Ozon (RU)",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=80,443", "--hostlist=" + listGeneral, "--dpi-desync=hostfakesplit", "--dpi-desync-repeats=4", "--dpi-desync-fooling=ts,md5sig", "--dpi-desync-hostfakesplit-mod=host=ozon.ru", "--new",
			},
		},
		{
			Name: "Strategy 10: Fake + 4pda TLS",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=80,443", "--hostlist=" + listGeneral, "--dpi-desync=fake", "--dpi-desync-repeats=6", "--dpi-desync-fooling=ts", "--dpi-desync-fake-tls=" + tls4pda, "--dpi-desync-fake-tls-mod=none", "--new",
			},
		},
		{
			Name: "Strategy 11: Multisplit + MaxRU TLS",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=80,443", "--hostlist=" + listGeneral, "--dpi-desync=fake,multisplit", "--dpi-desync-split-seqovl=654", "--dpi-desync-split-pos=1", "--dpi-desync-fooling=ts", "--dpi-desync-repeats=8", "--dpi-desync-split-seqovl-pattern=" + tlsMaxRu, "--dpi-desync-fake-tls=" + tlsMaxRu, "--new",
			},
		},
		{
			Name: "Strategy 12: Fakedsplit + Badseq",
			Args: []string{
				"--wf-tcp=80,443,2053,2083,2087,2096,8443", "--wf-udp=443,19294-19344,50000-50100",
				"--filter-udp=19294-19344,50000-50100", "--filter-l7=discord,stun", "--dpi-desync=fake", "--dpi-desync-fake-discord=" + quicGoogle, "--dpi-desync-fake-stun=" + quicGoogle, "--dpi-desync-repeats=6", "--new",
				"--filter-tcp=80,443", "--hostlist=" + listGeneral, "--dpi-desync=fake,fakedsplit", "--dpi-desync-split-pos=1", "--dpi-desync-fooling=badseq", "--dpi-desync-badseq-increment=2", "--dpi-desync-repeats=8", "--dpi-desync-fake-tls-mod=rnd,dupsid,sni=www.google.com", "--new",
			},
		},
	}

	// 3. Execution Loop
	fmt.Println("=== Liberty: Automatic Strategy Finder ===")
	fmt.Printf("Loaded %d strategies. Starting brute-force...\n", len(strategies))

	for i, s := range strategies {
		fmt.Printf("\n>>> Testing [%d/%d]: %s\n", i+1, len(strategies), s.Name)
		
		// Start winws
		cmd := exec.Command(winwsPath, s.Args...)
		// Redirect output to avoid clutter, or maybe log to file
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr
		
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error starting winws: %v\n", err)
			continue
		}

		// Wait for driver to init
		time.Sleep(3 * time.Second)

		// Test Connectivity
		success := runTests()

		// Cleanup
		if err := cmd.Process.Kill(); err != nil {
			fmt.Printf("Failed to kill winws: %v\n", err)
		}
		// Ждем освобождения драйвера
		time.Sleep(1 * time.Second)

		if success {
			fmt.Println("\n!!! SUCCESS !!!")
			fmt.Printf("Strategy '%s' works!\n", s.Name)
			saveWorkingConfig(s, rootDir)
			return
		}
	}

	fmt.Println("\nNo working strategy found. Try checking ipset or updating winws.")
}

func runTests() bool {
	// 1. YouTube TCP
	okY, _ := analyzer.CheckYouTube()
	statusY := "FAIL"
	if okY { statusY = "OK" }
	fmt.Printf("   YouTube TCP: %s\n", statusY)

	// 2. Discord UDP (Voice)
	// Using Rotterdam IP for raw UDP test
	okD, _ := analyzer.ProbeDiscordUDP("66.22.196.2", 50000, 50100)
	statusD := "FAIL"
	if okD { statusD = "OK" }
	fmt.Printf("   Discord UDP: %s\n", statusD)

	// Критерий успеха: Должен работать хотя бы YouTube TCP ИЛИ Discord UDP.
	// В идеале оба, но UDP часто блочится сильнее.
	return okY && okD
}

func killWinws() {
	exec.Command("taskkill", "/F", "/IM", "winws.exe").Run()
}

func saveWorkingConfig(s ConfigStrategy, dir string) {
	// TODO: Сохранить в файл для последующего запуска
	fmt.Println("Config saved (conceptually). You can hardcode Strategy #" + s.Name)
}
