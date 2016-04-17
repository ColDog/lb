package tools

import "log"

func Log(name string, evt map[string] interface{})  {
	x := 20 - len(name)
	sp := ""
	for i := 0; i < x; i++ {
		sp += " "
	}

	log.Printf("[%s]%s %v", name, sp, evt)
}
