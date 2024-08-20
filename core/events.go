package core

func Shutdown(store *Store) {
	evalBGREWRITEAOF([]string{}, store)
}
