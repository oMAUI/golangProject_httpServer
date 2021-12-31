package cuncurrency

type CunChecker func(string) bool

func CheckWebsites(wc CunChecker, urls []string) map[string]bool{
	results := make(map[string]bool)

	for _, url := range urls{
		go func(u string){
			results[u] = wc(u)
		}(url)
		
	}

	return results
}