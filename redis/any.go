package redisgroup

import (
	"encoding/json"
	"fmt"
	"github.com/cylScripter/chest/log"
	"math"
)

func toString(val interface{}) (string, error) {
	switch x := val.(type) {
	case bool:
		if x {
			return "1", nil
		}
		return "0", nil
	case int:
		return fmt.Sprintf("%d", x), nil
	case int8:
		return fmt.Sprintf("%d", x), nil
	case int16:
		return fmt.Sprintf("%d", x), nil
	case int32:
		return fmt.Sprintf("%d", x), nil
	case int64:
		return fmt.Sprintf("%d", x), nil
	case uint:
		return fmt.Sprintf("%d", x), nil
	case uint8:
		return fmt.Sprintf("%d", x), nil
	case uint16:
		return fmt.Sprintf("%d", x), nil
	case uint32:
		return fmt.Sprintf("%d", x), nil
	case uint64:
		return fmt.Sprintf("%d", x), nil
	case float32:
		if math.Floor(float64(x)) == float64(x) {
			return fmt.Sprintf("%.0f", x), nil
		}

		return fmt.Sprintf("%f", x), nil
	case float64:
		if math.Floor(x) == x {
			return fmt.Sprintf("%.0f", x), nil
		}

		return fmt.Sprintf("%f", x), nil
	case string:
		return x, nil
	case []byte:
		return string(x), nil
	case nil:
		return "", nil
	case error:
		return x.Error(), nil
	default:
		buf, err := json.Marshal(x)
		if err != nil {
			log.Errorf("err:%v", err)
			return "", err
		}

		return string(buf), nil
	}
}
