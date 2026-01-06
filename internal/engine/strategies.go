package engine

// Helper: Generates filter/new block for a specific list
func makeTcpRule(p *PipelineState, listName string, args []string) []string {
	prefix := []string{"--filter-tcp=443", "--hostlist=" + p.GetList(listName)}
	suffix := []string{"--new"}
	return append(append(prefix, args...), suffix...)
}

func makeUdpRule(p *PipelineState, listName string, args []string) []string {
	prefix := []string{"--filter-udp=443", "--hostlist=" + p.GetList(listName)}
	suffix := []string{"--new"}
	return append(append(prefix, args...), suffix...)
}

// === TCP STRATEGIES (Modular) ===

func Tcp_Split2(p *PipelineState, list string) []string {
	return makeTcpRule(p, list, []string{
		"--dpi-desync=split2", "--dpi-desync-split-seqovl=1", "--dpi-desync-split-pos=1",
	})
}

func Tcp_Fake(p *PipelineState, list string) []string {
	return makeTcpRule(p, list, []string{
		"--dpi-desync=fake", "--dpi-desync-repeats=6", "--dpi-desync-fooling=ts", "--dpi-desync-autottl=2", "--ip-id=zero",
		"--dpi-desync-fake-tls=" + p.GetBin("tls_clienthello_www_google_com.bin"),
	})
}

func Tcp_Multisplit_652(p *PipelineState, list string) []string {
	return makeTcpRule(p, list, []string{
		"--dpi-desync=multisplit", "--dpi-desync-split-seqovl=652", "--dpi-desync-split-pos=2", "--ip-id=zero",
		"--dpi-desync-split-seqovl-pattern=" + p.GetBin("tls_clienthello_www_google_com.bin"),
	})
}

func Tcp_Disorder(p *PipelineState, list string) []string {
	return makeTcpRule(p, list, []string{
		"--dpi-desync=fake,multidisorder", "--dpi-desync-fooling=badseq", "--dpi-desync-badseq-increment=1000",
		"--dpi-desync-repeats=6", "--dpi-desync-autottl=2",
	})
}

// === NEW STRATEGIES FROM @Untitled-1.txt (Exact Ports) ===

func Untitled_FakeDSplit(p *PipelineState, list string) []string {
	return makeTcpRule(p, list, []string{
		"--dpi-desync=fake,fakedsplit",
		"--dpi-desync-repeats=6",
		"--dpi-desync-fooling=ts",
		"--dpi-desync-fakedsplit-pattern=0x00",
		"--dpi-desync-fake-tls=" + p.GetBin("tls_clienthello_www_google_com.bin"),
	})
}

func Untitled_Media_SplitPorts(p *PipelineState, list string) []string {
	// Rule 1: Media Ports (FakeDSplit) - Extreme bypass
	rule1 := []string{
		"--filter-tcp=443,2053,2083,2087,2096,8443", "--hostlist=" + p.GetList(list),
		"--dpi-desync=fake,fakedsplit",
		"--dpi-desync-repeats=10",
		"--dpi-desync-fooling=ts",
		"--dpi-desync-fakedsplit-pattern=0x00",
		"--dpi-desync-fake-tls=" + p.GetBin("tls_clienthello_www_google_com.bin"),
		"--new",
	}
	return rule1
}

func Untitled_Media_SplitPorts_Variant2(p *PipelineState, list string) []string {
	// Rule 1: High Ports (FakeDSplit)
	rule1 := []string{
		"--filter-tcp=2053,2083,2087,2096,8443", "--hostlist=" + p.GetList(list),
		"--dpi-desync=fake,fakedsplit",
		"--dpi-desync-repeats=6",
		"--dpi-desync-fooling=ts",
		"--dpi-desync-fakedsplit-pattern=0x00",
		"--dpi-desync-fake-tls=" + p.GetBin("tls_clienthello_www_google_com.bin"),
		"--new",
	}

	// Rule 2: Port 443 (Use Disorder - stronger)
	rule2 := []string{
		"--filter-tcp=443", "--hostlist=" + p.GetList(list),
		"--dpi-desync=fake,multidisorder", 
		"--dpi-desync-fooling=badseq", 
		"--dpi-desync-badseq-increment=1000",
		"--dpi-desync-repeats=6", 
		"--dpi-desync-autottl=2",
		"--new",
	}

	return append(rule1, rule2...)
}

func Untitled_Voice_Specific(p *PipelineState) []string {
	return []string{
		"--filter-udp=50000-65535,19294-19344", "--filter-l7=discord,stun",
		"--dpi-desync=fake",
		"--dpi-desync-repeats=6",
		"--dpi-desync-fake-discord=" + p.GetBin("quic_initial_www_google_com.bin"),
		"--dpi-desync-fake-stun=" + p.GetBin("quic_initial_www_google_com.bin"),
		"--new",
	}
}

// === YOUTUBE STRATEGIES ===

func YT_Untitled_FakeDSplit(p *PipelineState, list string) []string {
	tcp := makeTcpRule(p, list, []string{
		"--ip-id=zero",
		"--dpi-desync=fake,fakedsplit",
		"--dpi-desync-repeats=6",
		"--dpi-desync-fooling=ts",
		"--dpi-desync-fakedsplit-pattern=0x00",
		"--dpi-desync-fake-tls=" + p.GetBin("tls_clienthello_www_google_com.bin"),
	})
	udp := makeUdpRule(p, list, []string{"--dpi-desync=fake", "--dpi-desync-repeats=6", "--dpi-desync-fake-quic=" + p.GetBin("quic_initial_www_google_com.bin")})
	return append(tcp, udp...)
}

func YT_Untitled_HostFakeSplit_Google(p *PipelineState, list string) []string {
	tcp := makeTcpRule(p, list, []string{
		"--ip-id=zero",
		"--dpi-desync=fake,hostfakesplit",
		"--dpi-desync-fake-tls-mod=rnd,dupsid,sni=www.google.com",
		"--dpi-desync-hostfakesplit-mod=host=www.google.com,altorder=1",
		"--dpi-desync-fooling=ts",
	})
	udp := makeUdpRule(p, list, []string{"--dpi-desync=fake", "--dpi-desync-repeats=6", "--dpi-desync-fake-quic=" + p.GetBin("quic_initial_www_google_com.bin")})
	return append(tcp, udp...)
}

func YT_L2_Standard(p *PipelineState, list string) []string {
	tcp := makeTcpRule(p, list, []string{"--dpi-desync=multisplit", "--dpi-desync-split-seqovl=681", "--dpi-desync-split-pos=1", "--ip-id=zero", "--dpi-desync-split-seqovl-pattern=" + p.GetBin("tls_clienthello_www_google_com.bin")})
	udp := makeUdpRule(p, list, []string{"--dpi-desync=fake", "--dpi-desync-repeats=11", "--ip-id=zero", "--dpi-desync-fake-quic=" + p.GetBin("quic_initial_www_google_com.bin")})
	return append(tcp, udp...)
}

func Media_Soft(p *PipelineState, list string) []string {
	return makeTcpRule(p, list, []string{
		"--dpi-desync=fake", 
		"--dpi-desync-repeats=6", 
		"--dpi-desync-fooling=badseq", 
		"--dpi-desync-badseq-increment=10",
		"--dpi-desync-autottl=2",
		"--ip-id=zero",
	})
}

func Media_SniSpoof_Google(p *PipelineState, list string) []string {
	return []string{
		"--filter-tcp=443,2053,2083,2087,2096,8443", "--hostlist=" + p.GetList(list),
		"--dpi-desync=fake,hostfakesplit",
		"--dpi-desync-fake-tls-mod=rnd,dupsid,sni=www.google.com",
		"--dpi-desync-hostfakesplit-mod=host=www.google.com,altorder=1",
		"--dpi-desync-fooling=ts",
		"--new",
	}
}

func Voice_Combo(p *PipelineState) []string {
	return []string{
		"--filter-udp=50000-65535,19294-19344", "--filter-l7=discord",
		"--dpi-desync=fake", "--dpi-desync-repeats=10", 
		"--dpi-desync-any-protocol=1",
		"--dpi-desync-udplen-increment=10",
		"--dpi-desync-fake-quic=" + p.GetBin("quic_initial_www_google_com.bin"),
		"--new",
	}
}
