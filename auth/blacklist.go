package auth

func CheckBlacklist(fullrepo string) bool {
	if fullrepo == "test/test1" {
		logw("%s in blacklist", fullrepo)
		return true
	}
	logw("%s not in blacklist", fullrepo)
	return false
}
