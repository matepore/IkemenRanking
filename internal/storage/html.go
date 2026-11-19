package storage

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/matepore/IkemenRanking/internal/ranking"
	"github.com/matepore/IkemenRanking/internal/tournament"
)

// TournamentReport contains the information presented in the standalone HTML report.
type TournamentReport struct {
	GeneratedAt time.Time
	Installation string
	Executable string
	Participants []string
	AILevel int
	RoundsToWin int
	Stage string
	Completed int
	Total int
	Stopped bool
	Standings []ranking.Standing
	Matches []tournament.MatchRecord
}

var reportTemplate = template.Must(template.New("report").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>IkemenRanking — Tournament Report</title>
<style>
:root{color-scheme:dark;--bg:#0b1020;--panel:#141d31;--panel2:#1a2740;--line:#2b3a56;--text:#edf3ff;--muted:#9aabc5;--accent:#61d5c5;--gold:#f2c66d}
*{box-sizing:border-box}body{margin:0;background:radial-gradient(ellipse at 15% 0%,#1b3150 0,transparent 42%),var(--bg);color:var(--text);font:15px/1.55 system-ui,-apple-system,"Segoe UI",sans-serif}
main{max-width:1120px;margin:0 auto;padding:42px 24px 64px}.eyebrow{color:var(--accent);font-size:12px;font-weight:800;letter-spacing:.18em;text-transform:uppercase}
h1{font-size:clamp(32px,5vw,52px);line-height:1.05;margin:10px 0}h2{font-size:20px;margin:0 0 18px}.sub{color:var(--muted);margin:0}
.hero{display:flex;justify-content:space-between;gap:24px;align-items:flex-end;margin-bottom:28px}.status{border:1px solid var(--line);background:var(--panel);padding:10px 14px;border-radius:12px;white-space:nowrap;color:var(--accent);font-weight:700}
.grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px;margin:22px 0}.metric,.panel{background:linear-gradient(145deg,var(--panel),#10192b);border:1px solid var(--line);border-radius:16px}.metric{padding:18px}.metric span{display:block;color:var(--muted);font-size:12px;text-transform:uppercase;letter-spacing:.08em}.metric strong{display:block;font-size:28px;margin-top:5px}
.panel{padding:22px;margin-top:18px}.table-wrap{overflow-x:auto}table{width:100%;border-collapse:collapse;white-space:nowrap}th,td{text-align:left;padding:12px 14px;border-bottom:1px solid var(--line)}th{color:var(--muted);font-size:11px;text-transform:uppercase;letter-spacing:.09em}tbody tr:last-child td{border-bottom:0}.rank{color:var(--gold);font-weight:800}.points{color:var(--accent);font-weight:800}
.config{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px}.config div{min-width:0}.config span{display:block;color:var(--muted);font-size:12px}.config strong{overflow-wrap:anywhere}
footer{color:var(--muted);font-size:12px;text-align:center;margin-top:28px}
@media(max-width:760px){main{padding:28px 14px}.hero{display:block}.status{display:inline-block;margin-top:16px}.grid{grid-template-columns:repeat(2,minmax(0,1fr))}.config{grid-template-columns:1fr}}
@media print{body{background:#fff;color:#111}.metric,.panel{background:#fff;border-color:#ccc}.sub,th,.config span,footer,.metric span{color:#555}.status{color:#075b51;border-color:#aaa}}
</style>
</head>
<body><main>
<header class="hero"><div><div class="eyebrow">Ikemen GO · Tournament analytics</div><h1>Tournament Report</h1><p class="sub">Generated {{.GeneratedAt.Format "2006-01-02 15:04:05"}} · {{len .Participants}} participants</p></div><div class="status">{{if .Stopped}}PARTIAL RESULTS{{else}}TOURNAMENT COMPLETE{{end}}</div></header>
<section class="grid">
<div class="metric"><span>Participants</span><strong>{{len .Participants}}</strong></div>
<div class="metric"><span>Fights completed</span><strong>{{.Completed}} / {{.Total}}</strong></div>
<div class="metric"><span>Leader</span><strong>{{if .Standings}}{{(index .Standings 0).Character}}{{else}}—{{end}}</strong></div>
<div class="metric"><span>Top score</span><strong>{{if .Standings}}{{(index .Standings 0).Points}}{{else}}0{{end}}</strong></div>
</section>
<section class="panel"><h2>Character standings</h2><div class="table-wrap"><table><thead><tr><th>Rank</th><th>Character</th><th>Points</th><th>Wins</th><th>Losses</th><th>Draws</th><th>Played</th></tr></thead><tbody>
{{range .Standings}}<tr><td class="rank">{{.Position}}</td><td>{{.Character}}</td><td class="points">{{.Points}}</td><td>{{.Wins}}</td><td>{{.Losses}}</td><td>{{.Draws}}</td><td>{{.Played}}</td></tr>{{else}}<tr><td colspan="7">No completed fights.</td></tr>{{end}}
</tbody></table></div></section>
<section class="panel"><h2>Fight results</h2><div class="table-wrap"><table><thead><tr><th>#</th><th>Player 1</th><th>Score</th><th>Player 2</th><th>Draws</th><th>Winner</th><th>Stage</th></tr></thead><tbody>
{{range .Matches}}<tr><td>{{.Index}}</td><td>{{.Match.Player1}}</td><td>{{.Result.Player1Wins}} – {{.Result.Player2Wins}}</td><td>{{.Match.Player2}}</td><td>{{.Result.Draws}}</td><td>{{if .Result.IsDraw}}Draw{{else}}{{.Result.Winner}}{{end}}</td><td>{{.Result.Stage}}</td></tr>{{else}}<tr><td colspan="7">No fights completed.</td></tr>{{end}}
</tbody></table></div></section>
<section class="panel"><h2>Tournament configuration</h2><div class="config">
<div><span>Ikemen GO installation</span><strong>{{.Installation}}</strong></div><div><span>Executable</span><strong>{{.Executable}}</strong></div>
<div><span>AI level</span><strong>{{.AILevel}}</strong></div><div><span>Rounds to win</span><strong>{{.RoundsToWin}}</strong></div>
<div><span>Stage</span><strong>{{.Stage}}</strong></div><div><span>Participants</span><strong>{{range $i,$p := .Participants}}{{if $i}}, {{end}}{{$p}}{{end}}</strong></div>
</div></section>
<footer>Generated by IkemenRanking · Results are provided for offline viewing</footer>
</main></body></html>`))

// ExportTournamentHTML writes a self-contained report that can be opened in any modern browser.
func ExportTournamentHTML(filePath string, report TournamentReport) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create HTML tournament report: %w", err)
	}
	if err := reportTemplate.Execute(file, report); err != nil {
		_ = file.Close()
		return fmt.Errorf("could not render HTML tournament report: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("could not close HTML tournament report: %w", err)
	}
	return nil
}
