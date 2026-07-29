import React, { useState } from 'react';
import GlobeView from './components/GlobeView';
import { Activity, ShieldAlert, Globe2, Radio, Menu, Search, Filter } from 'lucide-react';
import { mockEvents } from './lib/data';
import { formatDistanceToNow } from 'date-fns';
import { AreaChart, Area, ResponsiveContainer, YAxis, Tooltip } from 'recharts';

const mockChartData = Array.from({ length: 24 }).map((_, i) => ({
  time: i,
  value: Math.floor(Math.random() * 100) + 20
}));

export default function App() {
  const [selectedCountry, setSelectedCountry] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  return (
    <div className="w-screen h-screen bg-zinc-950 text-slate-200 overflow-hidden font-sans flex flex-col relative selection:bg-cyan-500/30">
      
      {/* Background Globe */}
      <div className="absolute inset-0 z-0">
        <GlobeView onSelectCountry={setSelectedCountry} />
      </div>

      {/* Top Header overlay */}
      <header className="absolute top-0 left-0 right-0 h-16 z-10 bg-zinc-950/40 backdrop-blur-md border-b border-white/5 flex items-center justify-between px-6">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded bg-cyan-500/20 border border-cyan-500/50 flex items-center justify-center">
            <Globe2 className="w-5 h-5 text-cyan-400" />
          </div>
          <div>
            <h1 className="text-lg font-bold tracking-wider text-slate-100 uppercase">GlobePulse AI</h1>
            <p className="text-[10px] text-cyan-400 font-mono tracking-widest uppercase">Global Threat Intelligence</p>
          </div>
        </div>

        <div className="flex-1 max-w-xl mx-8">
          <div className="relative group">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Search className="h-4 w-4 text-slate-400 group-focus-within:text-cyan-400 transition-colors" />
            </div>
            <input
              type="text"
              className="block w-full pl-10 pr-3 py-2 border border-white/10 rounded-md leading-5 bg-zinc-900/50 placeholder-slate-400 focus:outline-none focus:bg-zinc-900 focus:ring-1 focus:ring-cyan-500 focus:border-cyan-500 sm:text-sm transition-all"
              placeholder="Search countries, entities, or threat actors..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
            <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
              <Filter className="h-4 w-4 text-slate-400 hover:text-white cursor-pointer transition-colors" />
            </div>
          </div>
        </div>

        <div className="flex items-center gap-4 text-sm font-mono text-slate-400">
          <div className="flex items-center gap-2">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500"></span>
            </span>
            DEFCON 3
          </div>
          <div className="h-4 w-px bg-white/10"></div>
          <div className="flex items-center gap-2">
            <Radio className="w-4 h-4 text-cyan-500" />
            LIVE FEED
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 relative z-10 flex justify-between p-6 mt-16 pointer-events-none">
        
        {/* Left Panel: Global Feed */}
        <div className="w-80 flex flex-col gap-4 pointer-events-auto">
          <div className="bg-zinc-900/60 backdrop-blur-xl border border-white/10 rounded-xl overflow-hidden shadow-2xl flex flex-col h-[calc(100vh-140px)]">
            <div className="p-4 border-b border-white/5 flex items-center justify-between bg-white/5">
              <h2 className="font-semibold text-sm tracking-wider uppercase flex items-center gap-2 text-slate-200">
                <Activity className="w-4 h-4 text-cyan-400" />
                Intelligence Feed
              </h2>
              <span className="text-[10px] font-mono bg-cyan-500/10 text-cyan-400 px-2 py-0.5 rounded border border-cyan-500/20">
                {mockEvents.length} EVENTS
              </span>
            </div>
            
            <div className="flex-1 overflow-y-auto p-2 space-y-2 custom-scrollbar">
              {mockEvents.map((event) => (
                <div key={event.id} className="p-3 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 hover:border-white/10 transition-colors cursor-pointer group">
                  <div className="flex justify-between items-start mb-2">
                    <span className={`text-[10px] font-mono px-1.5 py-0.5 rounded border ${
                      event.sentiment === 'negative' ? 'bg-red-500/10 text-red-400 border-red-500/20' : 
                      event.sentiment === 'positive' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 
                      'bg-blue-500/10 text-blue-400 border-blue-500/20'
                    }`}>
                      {event.topic.toUpperCase()}
                    </span>
                    <span className="text-[10px] text-slate-500">
                      {formatDistanceToNow(new Date(event.timestamp), { addSuffix: true })}
                    </span>
                  </div>
                  <h3 className="text-sm font-medium text-slate-200 group-hover:text-cyan-400 transition-colors leading-snug mb-1">
                    {event.title}
                  </h3>
                  <div className="flex items-center gap-2 text-xs text-slate-400 mt-2">
                    <span className="truncate">{event.country}</span>
                    <span className="w-1 h-1 rounded-full bg-slate-600"></span>
                    <span className="font-mono text-cyan-500/70">ACT: {(event.score * 100).toFixed(0)}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Right Panel: Analytics & Details */}
        <div className="w-80 flex flex-col gap-4 pointer-events-auto">
          {selectedCountry ? (
            <div className="bg-zinc-900/60 backdrop-blur-xl border border-white/10 rounded-xl overflow-hidden shadow-2xl animate-in fade-in slide-in-from-right-4 duration-300">
              <div className="p-4 border-b border-white/5 bg-white/5">
                <h2 className="font-semibold text-lg text-white mb-1">{selectedCountry}</h2>
                <div className="flex items-center gap-2 text-xs font-mono text-slate-400">
                  <span>LAT: --</span>
                  <span>LNG: --</span>
                </div>
              </div>
              <div className="p-4 space-y-4">
                <div className="space-y-2">
                  <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider">AI Threat Assessment</h4>
                  <p className="text-sm text-slate-300 leading-relaxed">
                    Elevated activity detected in the region regarding infrastructure and technology sectors. NLP models indicate a 65% probability of continued geopolitical tension.
                  </p>
                </div>
                
                <div className="grid grid-cols-2 gap-2">
                  <div className="bg-white/5 border border-white/5 p-3 rounded-lg">
                    <div className="text-[10px] text-slate-500 uppercase">Threat Level</div>
                    <div className="text-lg font-mono text-amber-400 mt-1">ELEVATED</div>
                  </div>
                  <div className="bg-white/5 border border-white/5 p-3 rounded-lg">
                    <div className="text-[10px] text-slate-500 uppercase">Confidence</div>
                    <div className="text-lg font-mono text-cyan-400 mt-1">92.4%</div>
                  </div>
                </div>

                <div className="bg-white/5 border border-white/5 p-3 rounded-lg h-32 flex flex-col">
                  <div className="text-[10px] text-slate-500 uppercase mb-2">24H Activity Volume</div>
                  <div className="flex-1 w-full min-h-0">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={mockChartData}>
                        <defs>
                          <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#06b6d4" stopOpacity={0.3}/>
                            <stop offset="95%" stopColor="#06b6d4" stopOpacity={0}/>
                          </linearGradient>
                        </defs>
                        <Area type="monotone" dataKey="value" stroke="#06b6d4" strokeWidth={2} fillOpacity={1} fill="url(#colorValue)" />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                </div>
                
                <button className="w-full py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded text-xs font-medium transition-colors" onClick={() => setSelectedCountry(null)}>
                  Clear Selection
                </button>
              </div>
            </div>
          ) : (
            <div className="bg-zinc-900/60 backdrop-blur-xl border border-white/10 rounded-xl overflow-hidden shadow-2xl h-64 flex flex-col items-center justify-center text-center p-6 text-slate-500">
              <ShieldAlert className="w-12 h-12 mb-4 opacity-20" />
              <p className="text-sm">Select a region on the globe to view localized AI threat analysis and metrics.</p>
            </div>
          )}
          
          {/* Quick Stats Panel */}
          <div className="bg-zinc-900/60 backdrop-blur-xl border border-white/10 rounded-xl overflow-hidden shadow-2xl p-4">
            <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-4">Global Metrics</h3>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between text-xs mb-1">
                  <span className="text-slate-400">Total Analyzed</span>
                  <span className="font-mono text-cyan-400">12,492</span>
                </div>
                <div className="h-1 bg-white/10 rounded overflow-hidden">
                  <div className="h-full bg-cyan-500 w-[70%]"></div>
                </div>
              </div>
              <div>
                <div className="flex justify-between text-xs mb-1">
                  <span className="text-slate-400">Critical Threats</span>
                  <span className="font-mono text-red-400">42</span>
                </div>
                <div className="h-1 bg-white/10 rounded overflow-hidden">
                  <div className="h-full bg-red-500 w-[15%]"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Bottom Ticker */}
      <footer className="absolute bottom-0 left-0 right-0 h-8 z-10 bg-zinc-950/80 border-t border-white/5 flex items-center overflow-hidden">
        <div className="bg-cyan-500 text-zinc-950 font-bold text-[10px] px-3 h-full flex items-center tracking-widest z-20 shadow-[4px_0_12px_rgba(6,182,212,0.3)]">
          LATEST ALERTS
        </div>
        <div className="flex-1 overflow-hidden relative h-full flex items-center">
          <div className="animate-ticker whitespace-nowrap flex items-center gap-8 pl-4">
            {mockEvents.map((event, i) => (
              <div key={`ticker-${event.id}-${i}`} className="flex items-center gap-2 text-xs">
                <span className={
                  event.sentiment === 'negative' ? 'text-red-400' :
                  event.sentiment === 'positive' ? 'text-emerald-400' :
                  'text-blue-400'
                }>[{event.countryCode}]</span>
                <span className="text-slate-300">{event.title.toUpperCase()}</span>
                <span className="text-slate-500 font-mono">{(event.score * 100).toFixed(0)}</span>
              </div>
            ))}
            {/* Duplicate for infinite effect */}
            {mockEvents.map((event, i) => (
              <div key={`ticker-dup-${event.id}-${i}`} className="flex items-center gap-2 text-xs">
                <span className={
                  event.sentiment === 'negative' ? 'text-red-400' :
                  event.sentiment === 'positive' ? 'text-emerald-400' :
                  'text-blue-400'
                }>[{event.countryCode}]</span>
                <span className="text-slate-300">{event.title.toUpperCase()}</span>
                <span className="text-slate-500 font-mono">{(event.score * 100).toFixed(0)}</span>
              </div>
            ))}
          </div>
        </div>
      </footer>
    </div>
  );
}
