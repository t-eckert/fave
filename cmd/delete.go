package cmd

import (
	"fmt"
	"io"
	"net/http"
)

func RunDelete(args []string) error {
	if len(args) < 1 {
		return nil
	}

	id := args[0]

	req, err := http.NewRequest("DELETE", host+"/bookmarks/"+id, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)

	fmt.Println(string(respBody))

	return nil
}
