package customlog

import (
	"encoding/json"
	"log"
)

func PrintJSON(v any, p ...any) {
	jsonres, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Panic(err)
	}
	s := []any{string(jsonres)}
	p = append(s, p...)
	log.Println(p...)
}
