package httpx

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/projectdiscovery/httpx/runner"
)

func TestRun(t *testing.T) {
	client := NewHttpxClient()
	client.Run([]string{"a.fofa.info"}, func(r runner.Result) {
		if r.Err != nil {
			fmt.Println(r.Err)
			return
		}

		i, err := json.Marshal(r)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(i))
	})
}
