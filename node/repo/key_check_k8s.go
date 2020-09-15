// +build k8s_keystore_perms

package repo

import (
	"os"
)

func badMode(mode os.FileMode) bool {
	return mode&0037 != 0
}
