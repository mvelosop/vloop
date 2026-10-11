// The T9 gate's judge: prints, one per line, what budget.Check reports for the
// plugin directory named by its argument.
package main

import (
	"fmt"
	"os"

	"github.com/mvelosop/vloop/internal/budget"
)

func main() {
	for _, m := range budget.Check(os.Args[1]) {
		fmt.Println(m)
	}
}
