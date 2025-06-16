package metadata

// DJIModelMap 是 DJI 機型代號對應友善名稱
var DJIModelMap = map[string]string{
	"AC002":  "DJI_OsmoAction3",
	"AC003":  "DJI_OsmoAction4",
	"AC004":  "DJI_OsmoAction5_Pro",
	"OT-210": "DJI_POCKET_2",
}

// GetDJIModelFriendlyName 回傳對應的友善機型名稱，找不到則回傳原字串
func GetDJIModelFriendlyName(code string) (string, bool) {
	if name, ok := DJIModelMap[code]; ok {
		return name, true
	}
	return code, false
}
