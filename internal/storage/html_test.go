package storage

import (
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
 "github.com/matepore/IkemenRanking/internal/ikemen"
 "github.com/matepore/IkemenRanking/internal/ranking"
 "github.com/matepore/IkemenRanking/internal/tournament"
)

func TestExportRankingHTMLIncludesGlobalStandingsAndMatchups(t *testing.T) {
 path:=filepath.Join(t.TempDir(),"report.html")
 report:=RankingReport{GeneratedAt:time.Date(2026,11,18,12,30,0,0,time.UTC),
  Standings:[]ranking.Standing{{Position:1,Character:"Ryu",Points:2,Wins:2,Played:2},{Position:2,Character:"Ken",Points:1,Wins:1,Losses:1,Played:2},{Position:3,Character:"Chun-Li",Losses:2,Played:2}},
  Matches:[]tournament.MatchRecord{
   {Index:1,Match:tournament.Match{Player1:"Ryu",Player2:"Ken"},Result:ikemen.MatchResult{Player1Wins:2,Winner:"Ryu",Stage:"Training Room"}},
   {Index:2,Match:tournament.Match{Player1:"Ryu",Player2:"Chun-Li"},Result:ikemen.MatchResult{Player2Wins:2,Winner:"Chun-Li",Stage:"Temple"}},
   {Index:3,Match:tournament.Match{Player1:"Ken",Player2:"Chun-Li"},Result:ikemen.MatchResult{Player1Wins:2,Winner:"Ken",Stage:"Castle"}},
  }}
 if err:=ExportRankingHTML(path,report);err!=nil{t.Fatal(err)}
 content,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)};s:=string(content)
 for _,want:=range []string{"Global Character Ranking","Character standings","Current matchup results","Updated 2026-11-18 12:30:00","Characters ranked","Unique matchups","Ryu","Ken","Chun-Li","Training Room","Temple","Castle","2 – 0"}{if !strings.Contains(s,want){t.Errorf("report missing %q",want)}}
 if strings.Contains(s,"Tournament configuration")||strings.Contains(s,"Tournament Report"){t.Fatal("report contains tournament-specific sections")}
}
func TestExportRankingHTMLEscapesCharacterNames(t *testing.T){
 path:=filepath.Join(t.TempDir(),"report.html")
 if err:=ExportRankingHTML(path,RankingReport{Standings:[]ranking.Standing{{Position:1,Character:"<script>alert(1)</script>"}}});err!=nil{t.Fatal(err)}
 content,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
 if strings.Contains(string(content),"<script>alert(1)</script>")||!strings.Contains(string(content),"&lt;script&gt;"){t.Fatal("HTML escaping failed")}
}
func TestExportRankingHTMLKeepsOlderMatchupsWhenGlobalDataIncludesThem(t *testing.T){
 path:=filepath.Join(t.TempDir(),"report.html")
 report:=RankingReport{Standings:[]ranking.Standing{{Position:1,Character:"OldCharacter",Points:1}},Matches:[]tournament.MatchRecord{{Index:1,Match:tournament.Match{Player1:"OldCharacter",Player2:"OldOpponent"},Result:ikemen.MatchResult{Winner:"OldCharacter"}}}}
 if err:=ExportRankingHTML(path,report);err!=nil{t.Fatal(err)}
 report.GeneratedAt=time.Now();report.Standings=append(report.Standings,ranking.Standing{Position:2,Character:"NewCharacter",Points:1})
 report.Matches=append(report.Matches,tournament.MatchRecord{Index:2,Match:tournament.Match{Player1:"NewCharacter",Player2:"NewOpponent"},Result:ikemen.MatchResult{Winner:"NewCharacter"}})
 if err:=ExportRankingHTML(path,report);err!=nil{t.Fatal(err)}
 content,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
 for _,v:=range []string{"OldCharacter","OldOpponent","NewCharacter","NewOpponent"}{if !strings.Contains(string(content),v){t.Fatalf("report lost %q",v)}}
}
