//go:build linux

package config

var DicedbDataDir = func() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error fetching home directory, defaulting to /tmp/dicedb")
		return "/tmp/dicedb"
	}

	return filepath.Join(homeDir, ".local", "share", "dicedb"), nil
}()
