package ikemen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const legacySpeedScript = `if main.flags['-speed'] ~= nil and tonumber(main.flags['-speed']) > 0 then
	setGameSpeed(tonumber(main.flags['-speed']) * config.Framerate / 100)
end`

const modernSpeedScript = `if getCommandLineValue("-speed") ~= nil then
	local speed_input = tonumber(getCommandLineValue("-speed"))
	if speed_input ~= nil and speed_input >= -9 and speed_input <= 9 then
		local target_game_speed
		if speed_input > 0 then target_game_speed = speed_input end
		setGameSpeed(target_game_speed)
	end
end`

func TestParseGameSpeedMode(t *testing.T) {
	tests := []struct{name, source string; want GameSpeedMode; wantErr bool}{
		{name:"legacy percentage",source:legacySpeedScript,want:GameSpeedModePercentage},
		{name:"modern steps",source:modernSpeedScript,want:GameSpeedModeSteps},
		{name:"unknown implementation",source:"-- changed speed implementation",wantErr:true},
	}
	for _,tt:=range tests { t.Run(tt.name,func(t *testing.T){
		got,err:=ParseGameSpeedMode(tt.source)
		if (err!=nil)!=tt.wantErr {t.Fatalf("error=%v wantErr=%v",err,tt.wantErr)}
		if err==nil&&got!=tt.want {t.Fatalf("mode=%v want %v",got,tt.want)}
	})}
}
func TestDetectGameSpeedMode(t *testing.T){
	dir:=t.TempDir();path:=filepath.Join(dir,"external","script","main.lua")
	if err:=os.MkdirAll(filepath.Dir(path),0755);err!=nil{t.Fatal(err)}
	if err:=os.WriteFile(path,[]byte(legacySpeedScript),0600);err!=nil{t.Fatal(err)}
	got,err:=DetectGameSpeedMode(dir);if err!=nil{t.Fatal(err)}
	if got!=GameSpeedModePercentage{t.Fatalf("mode=%v want legacy percentage",got)}
}
func TestValidateGameSpeed(t *testing.T){
	tests:=[]struct{mode GameSpeedMode;value int;wantErr bool}{
		{GameSpeedModeSteps,0,false},{GameSpeedModeSteps,9,false},{GameSpeedModeSteps,10,true},{GameSpeedModeSteps,-1,true},
		{GameSpeedModePercentage,100,false},{GameSpeedModePercentage,200,false},{GameSpeedModePercentage,99,true},{GameSpeedModePercentage,201,true},
	}
	for _,tt:=range tests{err:=ValidateGameSpeed(tt.mode,tt.value);if (err!=nil)!=tt.wantErr{t.Errorf("ValidateGameSpeed(%v,%d) error=%v",tt.mode,tt.value,err)}}
}
func TestGameSpeedOptions(t *testing.T){
	legacy,lv,ld,err:=GameSpeedOptions(GameSpeedModePercentage);if err!=nil{t.Fatal(err)}
	if len(legacy)!=11||ld!="Normal (100%)"||lv[ld]!=100||lv["Faster (200%)"]!=200{t.Fatalf("bad legacy options: %v %#v",legacy,lv)}
	modern,mv,md,err:=GameSpeedOptions(GameSpeedModeSteps);if err!=nil{t.Fatal(err)}
	if len(modern)!=10||md!="Normal"||mv["Faster (+9)"]!=9{t.Fatalf("bad modern options: %v %#v",modern,mv)}
	for _,option:=range append(legacy,modern...){if strings.Contains(strings.ToLower(option),"slower"){t.Fatalf("slower option present: %q",option)}}
}
