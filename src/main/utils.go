package main

type JsonMap map[string]interface{}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func Any(vs []string, dst string) bool {
	for _, v := range vs {
		if v == dst {
			return true
		}
	}
	return false
}
