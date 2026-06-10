package disgoplus

import (
	"fmt"
	"strconv"

	"github.com/disgoorg/snowflake/v2"
)

// VarInt parses a route variable as an int.
func VarInt(vars map[string]string, key string) (int, error) {
	v, ok := vars[key]
	if !ok {
		return 0, fmt.Errorf("var %q not found", key)
	}

	return strconv.Atoi(v)
}

// VarSnowflake parses a route variable as a snowflake.ID.
func VarSnowflake(vars map[string]string, key string) (snowflake.ID, error) {
	v, ok := vars[key]
	if !ok {
		return 0, fmt.Errorf("var %q not found", key)
	}

	return snowflake.Parse(v)
}
