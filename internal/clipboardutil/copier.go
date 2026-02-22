package clipboardutil

import sysclipboard "github.com/atotto/clipboard"

type SystemCopier struct{}

func (SystemCopier) Copy(text string) error {
	return sysclipboard.WriteAll(text)
}
