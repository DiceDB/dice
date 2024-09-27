package config

func HandleSecureConnection() {
	if EnableSecureMode {
		DiceConfig.Server.Port = DiceConfig.Security.RespsPort
	}
}
