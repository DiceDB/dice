package config

func HandleSecureConnection() {
	if InitSecureMode {
		DiceConfig.Server.Port = DiceConfig.Security.RespsPort
	}
}
