package dashboard

// css is the one inline style sheet. Colours are custom properties defined
// once on :root and redefined for dark mode. The four series colours are
// assigned by position: series 1 is always --c1. No url(), no import.
const css = `:root{--bg:#f4f5f7;--card:#ffffff;--border:#d0d7de;--text:#1f2328;--muted:#59636e;--grid:#e3e7eb;--axis:#8c959f;--c1:#1f6feb;--c2:#c2570c;--c3:#1a7f4b;--c4:#8250df}
@media (prefers-color-scheme:dark){:root{--bg:#0d1117;--card:#161b22;--border:#30363d;--text:#e6edf3;--muted:#9198a1;--grid:#262c34;--axis:#6e7681;--c1:#58a6ff;--c2:#f0883e;--c3:#3fb950;--c4:#bc8cff}}
*{box-sizing:border-box}
body{margin:0;padding:16px;background:var(--bg);color:var(--text);font-family:system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;line-height:1.4}
header,main{max-width:1200px;margin:0 auto}
header{padding:0 0 16px}
h1{margin:0;font-size:1.5rem;line-height:1.25;overflow-wrap:anywhere}
h2{margin:0 0 12px;font-size:1rem;overflow-wrap:anywhere}
.subtitle{margin:4px 0 0;color:var(--muted)}
main{display:grid;gap:16px}
.row{display:grid;gap:16px}
.cols-1{grid-template-columns:minmax(0,1fr)}
.cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}
.cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}
.cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}
article{min-width:0;padding:16px;background:var(--card);border:1px solid var(--border);border-radius:8px}
.stat-label{margin:0;font-size:.8125rem;color:var(--muted);overflow-wrap:anywhere}
.stat-value{margin:4px 0 0;font-size:2rem;font-weight:600;line-height:1.2;overflow-wrap:anywhere}
.note{margin:4px 0 0;font-size:.8125rem;color:var(--muted);overflow-wrap:anywhere}
.legend{display:flex;flex-wrap:wrap;gap:4px 16px;margin:0 0 8px;padding:0;list-style:none;font-size:.8125rem}
.legend li{display:flex;align-items:center;gap:6px}
.sw{flex:none}
.chart{display:block;width:100%;max-width:720px;height:auto;margin:0 auto}
.chart text{fill:var(--muted);font-size:13px;font-family:inherit}
.plot{fill:none}
.grid{stroke:var(--grid);stroke-width:1}
.zero{stroke:var(--axis);stroke-width:1}
.b1{fill:var(--c1)}.b2{fill:var(--c2)}.b3{fill:var(--c3)}.b4{fill:var(--c4)}
.m1{fill:var(--c1)}.m2{fill:var(--c2)}.m3{fill:var(--c3)}.m4{fill:var(--c4)}
path.l1,path.l2,path.l3,path.l4{fill:none;stroke-width:2;stroke-linejoin:round}
.l1{stroke:var(--c1)}.l2{stroke:var(--c2);stroke-dasharray:7 4}.l3{stroke:var(--c3);stroke-dasharray:2 3}.l4{stroke:var(--c4);stroke-dasharray:9 3 2 3}
.scroll{overflow-x:auto}
table{border-collapse:collapse;width:100%;font-size:.875rem}
th,td{padding:6px 12px;border-bottom:1px solid var(--border);text-align:left;white-space:nowrap}
th{color:var(--muted);font-weight:600}
.num{text-align:right;font-variant-numeric:tabular-nums}
@media (max-width:720px){.row.cols-2,.row.cols-3,.row.cols-4{grid-template-columns:minmax(0,1fr)}}
`

// solidCSS is dashboard.css of the solid target. Every selector starts with
// the class prefix, so the sheet cannot restyle the host page. Colours, space,
// type size and radius come from the host app's custom properties, each with a
// fallback; the sheet defines none of them, so the component takes on the
// app's theme. The series colours 3 and 4 have no host property: --fd-c3 and
// --fd-c4 may be set by the app, otherwise the fallbacks apply. No url(), no
// import.
const solidCSS = `.fd-dashboard{box-sizing:border-box;display:grid;gap:var(--space-lg,24px);color:var(--text,#1b1f24);line-height:1.4}
.fd-dashboard *,.fd-dashboard *::before,.fd-dashboard *::after{box-sizing:inherit}
.fd-header{margin:0}
.fd-title{margin:0;font-size:1.5rem;line-height:1.25;overflow-wrap:anywhere}
.fd-subtitle{margin:var(--space-xs,4px) 0 0;color:var(--muted,#59636e)}
.fd-main{display:grid;gap:var(--space-md,16px)}
.fd-row{display:grid;gap:var(--space-md,16px)}
.fd-cols-1{grid-template-columns:minmax(0,1fr)}
.fd-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}
.fd-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}
.fd-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}
.fd-widget{min-width:0;padding:var(--space-md,16px);background:var(--panel,#ffffff);border:1px solid var(--border,#d0d7de);border-radius:var(--radius-md,8px)}
.fd-heading{margin:0 0 var(--space-sm,8px);font-size:var(--font-sm,.875rem);overflow-wrap:anywhere}
.fd-stat-label{margin:0;font-size:var(--font-xs,.75rem);color:var(--muted,#59636e);overflow-wrap:anywhere}
.fd-stat-value{margin:var(--space-xs,4px) 0 0;font-size:2rem;font-weight:600;line-height:1.2;overflow-wrap:anywhere}
.fd-note{margin:var(--space-xs,4px) 0 0;font-size:var(--font-xs,.75rem);color:var(--muted,#59636e);overflow-wrap:anywhere}
.fd-legend{display:flex;flex-wrap:wrap;gap:var(--space-xs,4px) var(--space-md,16px);margin:0 0 var(--space-sm,8px);padding:0;list-style:none;font-size:var(--font-xs,.75rem);color:var(--muted,#59636e)}
.fd-legend li{display:flex;align-items:center;gap:var(--space-xs,4px)}
.fd-sw{flex:none}
.fd-chart{display:block;width:100%;max-width:720px;height:auto;margin:0 auto}
.fd-chart text{fill:var(--muted,#59636e);font-size:13px;font-family:inherit}
.fd-plot{fill:none}
.fd-grid{stroke:var(--border,#d0d7de);stroke-width:1}
.fd-zero{stroke:var(--faint,#8c959f);stroke-width:1}
.fd-b1,.fd-m1{fill:var(--accent,#1f6feb)}
.fd-b2,.fd-m2{fill:var(--warn,#c2570c)}
.fd-b3,.fd-m3{fill:var(--fd-c3,#1a7f4b)}
.fd-b4,.fd-m4{fill:var(--fd-c4,#8250df)}
path.fd-l1,path.fd-l2,path.fd-l3,path.fd-l4{fill:none;stroke-width:2;stroke-linejoin:round}
.fd-l1{stroke:var(--accent,#1f6feb)}
.fd-l2{stroke:var(--warn,#c2570c);stroke-dasharray:7 4}
.fd-l3{stroke:var(--fd-c3,#1a7f4b);stroke-dasharray:2 3}
.fd-l4{stroke:var(--fd-c4,#8250df);stroke-dasharray:9 3 2 3}
.fd-scroll{overflow-x:auto}
.fd-table{border-collapse:collapse;width:100%;font-size:var(--font-sm,.875rem)}
.fd-table th,.fd-table td{padding:var(--space-xs,4px) var(--space-md,16px);border-bottom:1px solid var(--border,#d0d7de);text-align:left;white-space:nowrap}
.fd-table th{color:var(--muted,#59636e);font-weight:600}
.fd-table .fd-num{text-align:right;font-variant-numeric:tabular-nums}
@media (max-width:720px){.fd-row.fd-cols-2,.fd-row.fd-cols-3,.fd-row.fd-cols-4{grid-template-columns:minmax(0,1fr)}}
`
