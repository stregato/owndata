package assist

import (
	"fmt"

	"github.com/stregato/mio/cli/styles"
)

func Errorf(format string, args ...interface{}) {
	styles.ErrorStyle.Render(fmt.Sprintf(format, args...))
}
