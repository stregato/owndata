package settings

type KeySetting string

const (
	DBPathSetting       KeySetting = "dbPath"
	CurrentUserSettings KeySetting = "currentUser"

	KeySettingNode = "settings"
)

var (
	DefaultNick = "alice"
)
