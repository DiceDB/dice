package eval

func setGetKeys(args []string, ks *KeySpecs) {

	for i := 2; i < len(args); i++ {
		if (len(args[i]) == 3) &&
			(args[i][0] == 'g' || args[i][0] == 'G') &&
			(args[i][1] == 'e' || args[i][1] == 'E') &&
			(args[i][2] == 't' || args[i][2] == 'T') {
			ks.Flags = RW | ACCESS | UPDATE
			return
		}
	}

	ks.Flags = OW | UPDATE
}
