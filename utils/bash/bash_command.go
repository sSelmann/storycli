package bash

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		fmt.Printf("\n✖ Error executing command: %s %v\n", name, args)
		fmt.Println("Error:", err)
		fmt.Println("Stdout:", stdout.String())
		fmt.Println("Stderr:", stderr.String())
		return err
	}

	return nil
}
