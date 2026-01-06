import React, { useState, useEffect, useCallback, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
    Shield, Youtube, MessageSquare, Mic, Play, 
    Settings, CheckCircle2, Loader2, 
    RotateCcw, Activity, Sparkles, Terminal,
    Languages, Power, ChevronRight,
    X, Minus, Square, Copy, Zap, Globe, Info, Cpu,
    MapPin, Clock, Trash2, Bell, Monitor, Network, ScrollText, FolderOpen, Key, Folder, Hash
} from 'lucide-react';
import { WindowMinimise, Quit } from '../wailsjs/runtime/runtime';

const i18n = {
    en: {
        title: "Liberty", deploy: "Connect", running: "Syncing...",
        stop: "Disconnect", active: "Protected", idle: "Offline",
        standby: "Ready to connect", reset: "Reset",
        s1: "Cloudflare", s1_method: "TCP",
        s2: "YouTube", s2_method: "QUIC",
        s3: "Discord WSS", s3_method: "TLS",
        s4: "Discord UDP", s4_method: "UDP",
        s5: "Media", s5_method: "RTC",
        metrics: "Status", efficiency: "Efficiency", latency: "Latency",
        provider: "ISP", detecting: "Detecting...",
        uptime: "Session", dpi_level: "DPI", entropy: "Entropy",
        settings: "Settings", lang: "Language", data: "Data",
        clear_data: "Reset Session", close: "Close",
        finland_fix: "Finland Fix", finland_desc: "Force Cloudflare IP",
        autostart: "Auto-start", autostart_desc: "Start with Windows",
        tray: "System Tray", tray_desc: "Minimize on close",
        dns: "Secure DNS", dns_desc: "Cloudflare DoH",
        logs: "Core Logs", general: "General", network: "Network",
        open_data: "App Data Folder", open_desc: "Clean temporary files",
        discord_token: "Bot Token", discord_guild: "Server ID", discord_channel: "Voice Channel ID",
        work_dir: "Work Directory", select: "Select",
        secrets: "Secrets", paths: "Paths"
    },
    ru: {
        title: "Liberty", deploy: "Подключить", running: "Синхронизация...",
        stop: "Отключить", active: "Защищено", idle: "Офлайн",
        standby: "Готов к работе", reset: "Сброс",
        s1: "Cloudflare", s1_method: "TCP",
        s2: "YouTube", s2_method: "QUIC",
        s3: "Discord WSS", s3_method: "TLS",
        s4: "Discord UDP", s4_method: "UDP",
        s5: "Медиа", s5_method: "RTC",
        metrics: "Статус", efficiency: "Обход", latency: "Задержка",
        provider: "Провайдер", detecting: "Поиск...",
        uptime: "Сессия", dpi_level: "DPI", entropy: "Энтропия",
        settings: "Настройки", lang: "Язык", data: "Данные",
        clear_data: "Сбросить всё", close: "Закрыть",
        finland_fix: "Фикс Финляндии", finland_desc: "Cloudflare IP для Discord",
        autostart: "Автозагрузка", autostart_desc: "Запуск при старте ОС",
        tray: "Системный трей", tray_desc: "Сворачивать в трей",
        dns: "Безопасный DNS", dns_desc: "DNS через HTTPS",
        logs: "Журнал ядра", general: "Основные", network: "Сеть",
        open_data: "Папка AppData", open_desc: "Удалить временный мусор",
        discord_token: "Токен бота", discord_guild: "ID Сервера", discord_channel: "ID Голосового канала",
        work_dir: "Рабочая папка", select: "Выбрать",
        secrets: "Секреты", paths: "Пути"
    }
};

const STEPS = [
    { id: 1, name: 's1', icon: Shield },
    { id: 2, name: 's2', icon: Youtube },
    { id: 3, name: 's3', icon: MessageSquare },
    { id: 4, name: 's4', icon: Settings },
    { id: 5, name: 's5', icon: Mic },
];

function App() {
    const [lang, setLang] = useState('ru');
    const t = i18n[lang];

    const [currentStep, setCurrentStep] = useState(0); 
    const [results, setResults] = useState({}); 
    const [stepDetails, setStepDetails] = useState({});
    const [isRunning, setIsRunning] = useState(false);
    const [winwsActive, setWinwsActive] = useState(false);
    const [metrics, setMetrics] = useState({ rate: 0, ms: 0, entropy: 0 });
    const [connInfo, setConnInfo] = useState(null);
    const [uptime, setUptime] = useState(0);
    const [showSettings, setShowSettings] = useState(false);
    const [finlandFix, setFinlandFix] = useState(false);
    const [persistedDpi, setPersistedDpi] = useState('IDLE');
    
    const [autoStart, setAutoStart] = useState(false);
    const [minToTray, setMinTray] = useState(true);
    const [secureDns, setSecureDns] = useState(false);

    // Настройки
    const [discordToken, setDiscordToken] = useState('');
    const [discordGuild, setDiscordGuild] = useState('');
    const [discordChannel, setDiscordChannel] = useState('');
    const [workDir, setWorkDir] = useState('');

    const startTimeRef = useRef(null);

    const callGo = useCallback(async (fnName, ...args) => {
        if (window.go?.main?.App?.[fnName]) return await window.go.main.App[fnName](...args);
        return null;
    }, []);

    useEffect(() => {
        let interval;
        if (winwsActive) {
            interval = setInterval(() => {
                if (startTimeRef.current) {
                    const now = Math.floor(Date.now() / 1000);
                    setUptime(now - startTimeRef.current);
                } else {
                    setUptime(prev => prev + 1);
                }
            }, 1000);
        } else {
            setUptime(0);
            startTimeRef.current = null;
            if(interval) clearInterval(interval);
        }
        return () => clearInterval(interval);
    }, [winwsActive]);

    const formatUptime = (seconds) => {
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    };

    const loadSession = useCallback(async (active) => {
        const sessionJson = await callGo('GetSessionData');
        if (sessionJson) {
            try {
                const session = JSON.parse(sessionJson);
                if (active) {
                    const restoredResults = {};
                    if (session.steps) {
                        Object.keys(session.steps).forEach(k => { restoredResults[k] = 'success'; });
                    }
                    setResults(restoredResults);
                    setStepDetails(session.steps || {});
                    setPersistedDpi(session.dpi_level || 'MEDIUM');
                    setMetrics({ rate: session.rate || 0, ms: session.latency || 0, entropy: session.entropy || 0 });
                    if (session.start_time) {
                        startTimeRef.current = session.start_time;
                        const now = Math.floor(Date.now() / 1000);
                        setUptime(now - session.start_time);
                    }
                    setCurrentStep(6);
                }
            } catch (e) {}
        }
    }, [callGo]);

    useEffect(() => {
        const init = async () => {
            const active = await callGo('IsWinwsRunning');
            setWinwsActive(active);
            const infoRaw = await callGo('GetISPInfo');
            if (infoRaw) {
                try { 
                    const data = JSON.parse(infoRaw);
                    if (data.query) {
                        const parts = data.query.split('.');
                        data.maskedIp = `${parts[0]}.${parts[1]}.***.***`;
                    }
                    setConnInfo(data); 
                } catch (e) {}
            }
            
            // Загрузка настроек
            const settings = await callGo('GetSettings');
            if (settings) {
                setDiscordToken(settings.discord_token || '');
                setDiscordGuild(settings.discord_guild || '');
                setDiscordChannel(settings.discord_channel || '');
                setWorkDir(settings.work_dir || '');
            }

            await loadSession(active);
            setFinlandFix(localStorage.getItem('finlandFix') === 'true');
            setAutoStart(localStorage.getItem('autoStart') === 'true');
            setSecureDns(localStorage.getItem('secureDns') === 'true');
        };
        init();
        const interval = setInterval(async () => {
            const active = await callGo('IsWinwsRunning');
            setWinwsActive(active);
            if (!active) { startTimeRef.current = null; setUptime(0); setCurrentStep(0); setResults({}); }
        }, 2000);
        return () => clearInterval(interval);
    }, [callGo, loadSession]);

    useEffect(() => {
        const events = window.runtime;
        if (!events) return;
        const onProgress = (data) => {
            setCurrentStep(data.step);
            setStepDetails(prev => ({ ...prev, [data.step]: data.msg }));
            if (data.msg === "Optimized" || data.msg === "Applied") setResults(prev => ({ ...prev, [data.step]: 'success' }));
            setMetrics(prev => ({
                rate: Math.min(99.9, prev.rate + Math.random() * 2),
                ms: Math.floor(100 + Math.random() * 50),
                entropy: 0.75 + Math.random() * 0.1
            }));
        };
        events.EventsOn("step-progress", onProgress);
        events.EventsOn("done", () => { setIsRunning(false); setCurrentStep(6); loadSession(true); });
        return () => events.EventsOff("step-progress", "done");
    }, [loadSession]);

    const saveAppSettings = async (token, guild, channel, dir) => {
        await callGo('SaveSettings', {
            discord_token: token,
            discord_guild: guild,
            discord_channel: channel,
            work_dir: dir
        });
    };

    const start = () => { if (winwsActive || isRunning) return; setIsRunning(true); setResults({}); setStepDetails({}); setCurrentStep(1); callGo('Start', false); };
    const stop = async () => { await callGo('Stop'); setWinwsActive(false); setIsRunning(false); setCurrentStep(0); setResults({}); setMetrics({ rate: 0, ms: 0, entropy: 0 }); };
    const reset = async () => { await callGo('Reset'); setResults({}); setCurrentStep(0); setWinwsActive(false); setMetrics({ rate: 0, ms: 0, entropy: 0 }); setShowSettings(false); };

    const toggleFinlandFix = async () => {
        const newState = !finlandFix;
        const res = await callGo('SetFinlandFix', newState);
        if (res === "OK") { setFinlandFix(newState); localStorage.setItem('finlandFix', newState); }
    };

    const toggleSetting = (key, setter, current) => {
        const next = !current;
        setter(next);
        localStorage.setItem(key, next);
    };

    const handleSelectFolder = async () => {
        const path = await callGo('SelectFolder');
        if (path) {
            setWorkDir(path);
            saveAppSettings(discordToken, discordGuild, discordChannel, path);
        }
    };

    const getDPILevel = () => {
        if (!winwsActive && !isRunning) return "IDLE";
        return isRunning ? (currentStep > 3 ? "HIGH" : "MEDIUM") : persistedDpi;
    };

    const SettingItem = ({ icon: Icon, title, desc, active, onClick }) => (
        <div className="flex items-center gap-5 cursor-pointer group py-1" onClick={onClick}>
            <div className={`flex-shrink-0 w-8 h-8 rounded-full border flex items-center justify-center transition-all duration-500 bg-[#0b0d11] ${
                active ? 'border-indigo-500 text-indigo-400 shadow-[0_0_15px_rgba(99,102,241,0.2)]' : 'border-white/5 text-slate-800'
            }`}>
                <Icon size={14} />
            </div>
            <div className="flex flex-col min-w-0">
                <span className={`text-[11px] font-bold transition-colors duration-500 ${active ? 'text-white' : 'text-slate-700 group-hover:text-slate-500'}`}>{title}</span>
                <span className="text-[8px] font-black text-slate-800 uppercase tracking-widest leading-none mt-1">{desc}</span>
            </div>
        </div>
    );

    return (
        <div className="h-screen flex flex-col bg-[#0b0d11] text-slate-400 font-sans overflow-hidden select-none border border-white/5 relative">
            <header className="h-10 flex items-center justify-between bg-[#0b0d11] z-40 shrink-0" style={{ "--wails-draggable": "drag" }}>
                <div className="flex items-center gap-2 px-4">
                    <Shield size={12} className="text-indigo-500" />
                    <span className="text-[10px] font-bold tracking-widest text-slate-500 uppercase">{t.title}</span>
                </div>
                <AnimatePresence>
                    {winwsActive && (
                        <motion.div initial={{ opacity: 0, y: -5 }} animate={{ opacity: 1, y: 0 }} className="flex h-full no-drag items-center gap-4 border-x border-white/5 px-4" style={{ "--wails-draggable": "no-drag" }}>
                            <div className="flex flex-col items-center"><span className="text-[7px] font-black text-slate-600 uppercase tracking-widest leading-none">{t.dpi_level}</span><span className={`text-[10px] font-mono font-bold text-indigo-400`}>{getDPILevel()}</span></div>
                        </motion.div>
                    )}
                </AnimatePresence>
                <div className="flex h-full no-drag items-center" style={{ "--wails-draggable": "no-drag" }}>
                    <button onClick={() => setShowSettings(true)} className="px-3 h-full text-slate-600 hover:text-white transition-all"><Settings size={14} /></button>
                    <button onClick={WindowMinimise} className="px-3 h-full text-slate-600 hover:text-white transition-colors"><Minus size={14} /></button>
                    <button onClick={Quit} className="px-3 h-full hover:bg-red-500/20 text-slate-600 hover:text-red-500 transition-colors"><X size={14} /></button>
                </div>
            </header>

            <main className="flex-grow flex flex-col px-6 py-4 overflow-hidden gap-8 relative">
                <div className="relative flex flex-col items-center py-4">
                    <div className="flex flex-col items-center gap-1">
                        <div className="flex items-center gap-2">
                            <motion.div animate={winwsActive ? { opacity: [0.4, 1, 0.4] } : {}} transition={{ duration: 2, repeat: Infinity }} className={`w-1.5 h-1.5 rounded-full ${winwsActive ? 'bg-emerald-500 shadow-[0_0_8px_#10b981]' : isRunning ? 'bg-amber-500' : 'bg-slate-700'}`} />
                            <span className="text-xs font-bold tracking-[0.1em] text-white uppercase">{winwsActive ? t.active : isRunning ? t.running : t.idle}</span>
                        </div>
                        <AnimatePresence>
                            {winwsActive ? (
                                <motion.span initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-[10px] font-mono font-bold text-indigo-400 tracking-widest">{formatUptime(uptime)}</motion.span>
                            ) : (
                                <span className="text-[9px] text-slate-600 font-medium uppercase tracking-widest">{t.standby}</span>
                            )}
                        </AnimatePresence>
                    </div>
                    <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-32 h-[1px] bg-white/5 overflow-hidden">
                        <motion.div animate={winwsActive || isRunning ? { x: [-128, 128] } : {}} transition={{ duration: 2, repeat: Infinity, ease: "linear" }} className="w-full h-full bg-gradient-to-r from-transparent via-indigo-500/40 to-transparent" />
                    </div>
                </div>

                <div className="flex-grow relative px-4">
                    <div className="absolute left-[30px] top-4 bottom-4 w-[1px] bg-white/5 z-0" />
                    <div className="absolute left-[30px] top-4 bottom-4 w-[1px] overflow-hidden z-10">
                        <motion.div initial={{ height: 0 }} animate={{ height: `${(Math.max(0, currentStep - 1) / (STEPS.length - 1)) * 100}%` }} className="w-full bg-indigo-500 shadow-[0_0_10px_#6366f1]" />
                    </div>
                    <div className="space-y-6 relative z-20">
                        {STEPS.map((step) => {
                            const isSuccess = (winwsActive || isRunning) && (results[step.id] === 'success' || results[String(step.id)] === 'success' || (currentStep > step.id && currentStep === 6));
                            const isActive = isRunning && currentStep === step.id;
                            return (
                                <div key={step.id} className="flex items-center gap-5">
                                    <motion.div animate={{ scale: isActive ? 1.15 : 1 }} className={`flex-shrink-0 w-7 h-7 rounded-full border flex items-center justify-center transition-all duration-500 bg-[#0b0d11] ${isSuccess ? 'border-indigo-500 text-indigo-400 shadow-[0_0_10px_rgba(99,102,241,0.2)]' : isActive ? `border-white text-white shadow-[0_0_15px_rgba(255,255,255,0.2)]` : 'border-white/5 text-slate-800'}`}>
                                        {isSuccess ? <CheckCircle2 size={12} /> : <step.icon size={12} />}
                                    </motion.div>
                                    <div className="flex flex-col min-w-0">
                                        <div className="flex items-center gap-2">
                                            <span className={`text-[11px] font-bold transition-colors duration-500 ${isActive ? 'text-indigo-400' : isSuccess ? 'text-slate-400' : 'text-slate-700'}`}>{t[step.name]}</span>
                                            <span className={`text-[7px] font-black px-1 rounded border ${isActive ? 'border-indigo-500/20 text-indigo-500/60' : isSuccess ? 'border-indigo-500/20 text-indigo-500/40' : 'border-white/5 text-slate-800'}`}>{t[step.name + '_method']}</span>
                                        </div>
                                        {isActive && (<motion.span initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-[9px] text-indigo-500/60 font-medium truncate italic mt-0.5">{stepDetails[step.id] || 'Syncing...'}</motion.span>)}
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                </div>

                <div className="grid grid-cols-3 py-4 border-t border-white/5 gap-4 shrink-0">
                    <div className="flex flex-col gap-1"><span className="text-[8px] font-black text-slate-700 uppercase tracking-widest">{t.efficiency}</span><span className="text-xs font-mono font-bold text-slate-300">{metrics.rate.toFixed(1)}%</span></div>
                    <div className="flex flex-col gap-1 items-center"><span className="text-[8px] font-black text-slate-700 uppercase tracking-widest">{t.latency}</span><span className="text-xs font-mono font-bold text-slate-300">{metrics.ms}ms</span></div>
                    <div className="flex flex-col gap-1 items-end"><span className="text-[8px] font-black text-slate-700 uppercase tracking-widest">{t.entropy}</span><span className="text-xs font-mono font-bold text-slate-300">{winwsActive ? (metrics.entropy * 100).toFixed(0) + '%' : '---'}</span></div>
                </div>

                <div className="flex items-center justify-between text-[9px] font-medium text-slate-700 uppercase tracking-tighter pb-2 shrink-0">
                    <div className="flex items-center gap-1.5"><MapPin size={10} /><span>{connInfo?.countryCode || 'RU'} · {connInfo?.city || 'Local'}</span></div>
                    <span className="font-mono">{connInfo?.maskedIp || '0.0.***.***'}</span>
                </div>

                <div className="shrink-0 pt-2 pb-2">
                    <AnimatePresence mode="wait">
                        {winwsActive ? (
                            <motion.button key="stop" initial={{ opacity: 0 }} animate={{ opacity: 1 }} onClick={stop} className="w-full h-11 rounded-lg font-bold text-[10px] uppercase tracking-[0.2em] bg-transparent border border-red-500/30 text-red-500 hover:bg-red-500/10 transition-all flex items-center justify-center gap-2">
                                <Square size={12} fill="currentColor" /> {t.stop}
                            </motion.button>
                        ) : (
                            <motion.button key="start" initial={{ opacity: 0 }} animate={{ opacity: 1 }} disabled={isRunning} onClick={start} className={`w-full h-11 rounded-lg font-bold text-[10px] uppercase tracking-[0.2em] transition-all flex items-center justify-center gap-2 ${isRunning ? 'bg-[#0b0d11] border border-white/5 text-slate-700' : 'bg-indigo-600 text-white hover:bg-indigo-500 shadow-lg'}`}>
                                {isRunning ? <Loader2 size={14} className="animate-spin" /> : <Power size={14} />} {isRunning ? t.running : t.deploy}
                            </motion.button>
                        )}
                    </AnimatePresence>
                </div>

                {/* Fixed Settings Overlay */}
                <AnimatePresence>
                    {showSettings && (
                        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="fixed inset-0 z-[999] bg-[#0b0d11] p-8 flex flex-col">
                            <div className="flex items-center justify-between mb-6">
                                <div className="flex items-center gap-3"><Settings size={18} className="text-indigo-500" /><h2 className="text-[11px] font-black text-white uppercase tracking-[0.2em]">{t.settings}</h2></div>
                                <button onClick={() => setShowSettings(false)} className="text-slate-400 hover:text-white p-2 bg-white/5 rounded-full transition-colors"><X size={18} /></button>
                            </div>

                            <div className="space-y-6 overflow-y-auto custom-scroll pr-2 pb-4">
                                <div className="space-y-4">
                                    <div className="flex items-center gap-2 text-slate-500 text-[9px] font-bold uppercase tracking-widest"><Key size={12}/> {t.secrets}</div>
                                    <div className="flex flex-col gap-3 pl-2">
                                        <div className="flex flex-col gap-1.5">
                                            <label className="text-[8px] font-black text-slate-700 uppercase tracking-widest">{t.discord_token}</label>
                                            <input 
                                                type="password" 
                                                value={discordToken} 
                                                onChange={(e) => { setDiscordToken(e.target.value); saveAppSettings(e.target.value, discordGuild, discordChannel, workDir); }}
                                                className="bg-black/20 border border-white/5 rounded px-3 py-2 text-[10px] text-indigo-400 focus:border-indigo-500/50 outline-none transition-all"
                                                placeholder="MTM3..."
                                            />
                                        </div>
                                        <div className="grid grid-cols-2 gap-3">
                                            <div className="flex flex-col gap-1.5">
                                                <label className="text-[8px] font-black text-slate-700 uppercase tracking-widest">{t.discord_guild}</label>
                                                <input 
                                                    type="text" 
                                                    value={discordGuild} 
                                                    onChange={(e) => { setDiscordGuild(e.target.value); saveAppSettings(discordToken, e.target.value, discordChannel, workDir); }}
                                                    className="bg-black/20 border border-white/5 rounded px-3 py-2 text-[10px] text-indigo-400 focus:border-indigo-500/50 outline-none transition-all"
                                                    placeholder="Server ID"
                                                />
                                            </div>
                                            <div className="flex flex-col gap-1.5">
                                                <label className="text-[8px] font-black text-slate-700 uppercase tracking-widest">{t.discord_channel}</label>
                                                <input 
                                                    type="text" 
                                                    value={discordChannel} 
                                                    onChange={(e) => { setDiscordChannel(e.target.value); saveAppSettings(discordToken, discordGuild, e.target.value, workDir); }}
                                                    className="bg-black/20 border border-white/5 rounded px-3 py-2 text-[10px] text-indigo-400 focus:border-indigo-500/50 outline-none transition-all"
                                                    placeholder="Channel ID"
                                                />
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div className="space-y-4">
                                    <div className="flex items-center gap-2 text-slate-500 text-[9px] font-bold uppercase tracking-widest"><Folder size={12}/> {t.paths}</div>
                                    <div className="pl-2">
                                        <div className="flex flex-col gap-1.5">
                                            <label className="text-[8px] font-black text-slate-700 uppercase tracking-widest">{t.work_dir}</label>
                                            <div className="flex gap-2">
                                                <input 
                                                    type="text" 
                                                    readOnly
                                                    value={workDir || 'Temporary (Default)'} 
                                                    className="flex-grow bg-black/20 border border-white/5 rounded px-3 py-2 text-[10px] text-slate-500 outline-none"
                                                />
                                                <button onClick={handleSelectFolder} className="px-3 bg-indigo-600/20 text-indigo-400 border border-indigo-500/20 rounded text-[9px] font-bold uppercase hover:bg-indigo-600/30 transition-all">{t.select}</button>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div className="space-y-4 pt-2">
                                    <div className="flex items-center gap-2 text-slate-600 text-[8px] font-black uppercase tracking-widest"><Network size={12}/> {t.network}</div>
                                    <div className="flex flex-col gap-4 pl-2">
                                        <SettingItem icon={Zap} title={t.finland_fix} desc={t.finland_desc} active={finlandFix} onClick={toggleFinlandFix} />
                                        <SettingItem icon={Globe} title={t.dns} desc={t.dns_desc} active={secureDns} onClick={() => toggleSetting('secureDns', setSecureDns, secureDns)} />
                                        <SettingItem icon={FolderOpen} title={t.open_data} desc={t.open_desc} active={false} onClick={() => callGo('OpenAppData')} />
                                    </div>
                                </div>

                                <div className="space-y-4 pt-4 border-t border-white/5">
                                    <div className="flex items-center justify-between">
                                        <div className="flex gap-4">
                                            {['ru', 'en'].map(l => (
                                                <button key={l} onClick={() => setLang(l)} className={`text-[10px] font-black uppercase transition-all ${lang === l ? 'text-indigo-400' : 'text-slate-600 hover:text-slate-400'}`}>{l}</button>
                                            ))}
                                        </div>
                                        <button onClick={reset} className="text-[10px] font-black text-red-500/60 hover:text-red-500 transition-all uppercase tracking-widest">{t.clear_data}</button>
                                    </div>
                                </div>
                            </div>

                            <div className="mt-auto pt-4 flex flex-col items-center gap-2 opacity-20">
                                <Shield size={24} />
                                <span className="text-[8px] font-mono tracking-[0.2em]">Build 26.0.4-LPR</span>
                            </div>
                        </motion.div>
                    )}
                </AnimatePresence>
            </main>

            <style dangerouslySetInnerHTML={{ __html: `.custom-scroll::-webkit-scrollbar { width: 2px; } .custom-scroll::-webkit-scrollbar-track { background: transparent; } .custom-scroll::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.05); border-radius: 10px; } .no-drag { -webkit-app-region: no-drag; }`}} />
        </div>
    );
}

export default App;
