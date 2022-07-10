package helpers

import (
	"fmt"
	"os"

	"github.com/k0kubun/pp/v3"
)

func PP(inputs ...interface{}) {
	fmt.Printf("\n############## Debug ##############")

	for _, input := range inputs {
		fmt.Printf("\n")

		pp.Print(input)

		fmt.Printf("\n")
	}

	fmt.Printf("###################################\n\n")
}

func DD(inputs ...interface{}) {
	PP(inputs...)

	os.Exit(1)
}
