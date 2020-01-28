package utils

import (
	"fmt"
	"github.com/mpetavy/common"
	"math"
	"path/filepath"
	"strings"
)

func CreateHierarchicalPath(flat bool, id int) (string, error) {
	if flat {
		return fmt.Sprintf("%d", id), nil
	} else {
		sb := strings.Builder{}

		for i := 3; i >= 0; i-- {
			if i < 3 {
				_, err := sb.WriteString(string(filepath.Separator))
				if common.Error(err) {
					return "", err
				}
			}

			t := math.Pow(float64(1000), float64(i))
			v := (id / int(t)) * int(t)

			vs := fmt.Sprintf("%012d", v)

			_, err := sb.WriteString(vs)
			if common.Error(err) {
				return "", err
			}
		}

		return common.CleanPath(sb.String()), nil
	}
}
