//go:build linux

package config

const DicedbDataDir = func() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error fetching home directory")
		return "/tmp/dicedb"
	}

	return filepath.Join(homeDir, ".local", "share", "dicedb"), nil
}()
