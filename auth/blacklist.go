package auth

func CheckBlacklist(fullrepo string, blist []string) bool {
	for _, blocked := range blist {
		if blocked == fullrepo {
			return true
		}
	}
	logw("%s not in blacklist", fullrepo)
	return false
}
