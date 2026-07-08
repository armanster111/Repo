//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type barsPayload struct {
	Bars   []float64 `json:"bars"`
	Kick   float64   `json:"kick"`
	Snare  float64   `json:"snare"`
	HiHat  float64   `json:"hihat"`
	Mode   string    `json:"mode"`
	Mood   string    `json:"mood"`
	BPM    float64   `json:"bpm"`
	Phase  float64   `json:"phase"`
	Title  string    `json:"title"`
	Artist string    `json:"artist"`
}

var streamMu sync.RWMutex
var lastBarsPayload barsPayload

func (s *appState) publishBarsStream(bars []float64) {
	if !s.remoteEnabled {
		return
	}
	payload := barsPayload{
		Bars:   append([]float64(nil), bars...),
		Mode:   modeName(s.mode),
		Mood:   s.mood,
		BPM:    s.bpm,
		Title:  s.meta.Title,
		Artist: s.meta.Artist,
	}
	if s.onset != nil {
		payload.Kick = s.onset.Kick
		payload.Snare = s.onset.Snare
		payload.HiHat = s.onset.HiHat
		payload.Phase = s.onset.Phase
	}
	streamMu.Lock()
	lastBarsPayload = payload
	streamMu.Unlock()
}

func registerVizRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/bars", func(w http.ResponseWriter, r *http.Request) {
		streamMu.RLock()
		payload := lastBarsPayload
		streamMu.RUnlock()
		if len(payload.Bars) == 0 && app.currentBars != nil {
			payload.Bars = append([]float64(nil), app.currentBars...)
			payload.Mode = modeName(app.mode)
			payload.Mood = app.mood
			payload.BPM = app.bpm
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_ = json.NewEncoder(w).Encode(payload)
	})
	mux.HandleFunc("/viz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, vizPageHTML)
	})
}

const vizPageHTML = `<!DOCTYPE html>
<html lang="en"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Music Visualizer Ultra — Live</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#06040f;color:#fff;font-family:Segoe UI,system-ui,sans-serif;overflow:hidden}
#meta{position:fixed;top:12px;left:16px;z-index:2;text-shadow:0 2px 8px #000}
#meta h1{font-size:18px;font-weight:600}
#meta p{opacity:.75;font-size:13px;margin-top:4px}
canvas{display:block;width:100vw;height:100vh}
#badge{position:fixed;bottom:12px;right:16px;font-size:12px;opacity:.5}
</style></head><body>
<div id="meta"><h1 id="title">Music Visualizer Ultra</h1><p id="sub">Live stream</p></div>
<canvas id="c"></canvas>
<div id="badge">localhost:8765/viz</div>
<script>
const c=document.getElementById('c'),x=c.getContext('2d');
let W,H,bars=new Array(64).fill(0),kick=0,phase=0;
function resize(){W=c.width=innerWidth;H=c.height=innerHeight}
addEventListener('resize',resize);resize();
async function tick(){
  try{
    const r=await fetch('/api/bars');const d=await r.json();
    if(d.bars&&d.bars.length)bars=d.bars;
    kick=d.kick||0;phase=d.phase||0;
    document.getElementById('title').textContent=(d.title||'Music Visualizer Ultra');
    document.getElementById('sub').textContent=[d.artist,d.mode,d.mood,d.bpm?Math.round(d.bpm)+' BPM':''].filter(Boolean).join(' · ');
  }catch(e){}
  x.fillStyle='rgba(6,4,15,0.35)';x.fillRect(0,0,W,H);
  const cx=W/2,cy=H/2,maxR=Math.min(W,H)*0.42;
  const pulse=1+kick*0.6;
  for(let i=0;i<bars.length;i++){
    const e=bars[i]*pulse,a=i/bars.length*Math.PI*2+phase*Math.PI*2;
    const r=maxR*(0.12+e*0.88),px=cx+Math.cos(a)*r,py=cy+Math.sin(a)*r;
    const hue=(i/bars.length*280+phase*120)%360;
    x.fillStyle='hsla('+hue+',90%,'+(45+e*35)+'%,'+(0.35+e*0.65)+')';
    x.beginPath();x.arc(px,py,2+e*10,0,Math.PI*2);x.fill();
  }
  if(kick>0.15){
    x.strokeStyle='hsla(280,80%,70%,'+kick*0.5+')';
    x.lineWidth=2+kick*6;
    x.beginPath();x.arc(cx,cy,maxR*kick*0.9,0,Math.PI*2);x.stroke();
  }
  requestAnimationFrame(tick);
}
tick();
</script></body></html>`
