// ---- shared 3D projector ---------------------------------------------------
// World: pole = +Y. dir(θ,φ) = (sinφ cosθ, cosφ, sinφ sinθ). View tilted about X so the
// equator reads as an ellipse and we can see "above" the equator.
const TILT = -22 * Math.PI / 180;
// small yaw about the vertical (pole) axis: keeps the pole vertical but spreads
// the equatorial axes apart in screen-x, so the 3rd line no longer hides behind it.
const YAW = 28 * Math.PI / 180;
function dirOf(th, ph){ return { x: Math.sin(ph)*Math.cos(th), y: Math.cos(ph), z: Math.sin(ph)*Math.sin(th) }; }
function proj(th, ph, cx, cy, R){
  const d0 = dirOf(th, ph);
  const d = { x: d0.x*Math.cos(YAW) + d0.z*Math.sin(YAW), y: d0.y, z: -d0.x*Math.sin(YAW) + d0.z*Math.cos(YAW) };
  const y2 = d.y*Math.cos(TILT) - d.z*Math.sin(TILT);
  const z2 = d.y*Math.sin(TILT) + d.z*Math.cos(TILT);
  return { x: cx + R*d.x, y: cy - R*y2, front: z2 >= -0.02 };
}
const NS = "http://www.w3.org/2000/svg";
function el(tag, attrs){ const e=document.createElementNS(NS,tag); for(const k in attrs) e.setAttribute(k, attrs[k]); return e; }
function poly(samples, attrs){
  const p = el("polyline", attrs); p.setAttribute("points", samples.map(s=>s.x.toFixed(1)+","+s.y.toFixed(1)).join(" ")); return p;
}
// draw a curve on the sphere, splitting front (solid) / back (dashed)
function sphereCurve(svg, fn, n, cx, cy, R, color, w){
  let run=[], front=null;
  const flush=()=>{ if(run.length>1){ svg.appendChild(poly(run,{fill:"none",stroke:color,"stroke-width":w,
    "stroke-opacity":front?1:0.35,"stroke-dasharray":front?"":"4 5"})); } run=[]; };
  for(let i=0;i<=n;i++){ const p=fn(i/n); if(front===null)front=p.front; if(p.front!==front){flush();front=p.front;} run.push(p); }
  flush();
}
function dot(svg,x,y,c,r){ svg.appendChild(el("circle",{cx:x,cy:y,r:r||4,fill:c})); }
function line(svg,x1,y1,x2,y2,c,w,dash){ svg.appendChild(el("line",{x1,y1,x2,y2,stroke:c,"stroke-width":w||2,"stroke-dasharray":dash||""})); }
function label(svg,x,y,t,c,size){ const e=el("text",{x,y,fill:c||"#e7ecf5","font-size":size||13}); e.textContent=t; svg.appendChild(e); }
function arrow(svg,x1,y1,x2,y2,c,w){
  line(svg,x1,y1,x2,y2,c,w||2);
  const a=Math.atan2(y2-y1,x2-x1), h=9;
  line(svg,x2,y2,x2-h*Math.cos(a-0.45),y2-h*Math.sin(a-0.45),c,w||2);
  line(svg,x2,y2,x2-h*Math.cos(a+0.45),y2-h*Math.sin(a+0.45),c,w||2);
}

// ---- generic sphere scaffold ----------------------------------------------
function sphere(svg, cx, cy, R){
  svg.appendChild(el("circle",{cx,cy,r:R,fill:"#10182e",stroke:"#2b3350","stroke-width":1.5}));
  // parallels
  for(const ph of [30,60,90,120,150].map(d=>d*Math.PI/180))
    sphereCurve(svg,t=>proj(t*2*Math.PI,ph,cx,cy,R),80,cx,cy,R,"#2f3a5c",1);
  // meridians
  for(const th of [0,45,90,135].map(d=>d*Math.PI/180))
    sphereCurve(svg,t=>proj(th,t*2*Math.PI,cx,cy,R),80,cx,cy,R,"#2f3a5c",1);
}
function poleAxis(svg,cx,cy,R){
  const top=proj(0,0,cx,cy,R), bot=proj(0,Math.PI,cx,cy,R);
  line(svg,cx,bot.y+18,cx,top.y-18,"#c79cff",2);
  label(svg,cx+8,top.y-10,"pole","#c79cff");
  dot(svg,top.x,top.y,"#c79cff",3.5); dot(svg,bot.x,bot.y,"#c79cff",3.5);
}
