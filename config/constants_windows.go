//go:build windows

package config

const DicedbDataDir = func() string {
	// Use %LOCALAPPDATA% for dir on Windows i.e C:\Users\<YourUsername>\AppData\Local
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		fmt.Println("Error fetching home directory, defaulting to current dir")
	}

	return filepath.Join(localAppData, "dicedb")
}()
