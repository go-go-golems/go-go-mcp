package embeddable

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Arguments provides convenient methods for accessing tool arguments
// with type conversion and validation, inspired by mark3labs/mcp-go
type Arguments map[string]interface{}

// NewArguments creates an Arguments instance from a map
func NewArguments(args map[string]interface{}) Arguments {
	if args == nil {
		return make(Arguments)
	}
	return Arguments(args)
}

// Raw returns the underlying map for direct access
func (a Arguments) Raw() map[string]interface{} {
	return map[string]interface{}(a)
}

// BindArguments unmarshals the arguments into the provided struct
func (a Arguments) BindArguments(target interface{}) error {
	if target == nil || reflect.ValueOf(target).Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	data, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}

	return json.Unmarshal(data, target)
}

// String argument access
func (a Arguments) GetString(key string, defaultValue string) string {
	if val, ok := a[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (a Arguments) RequireString(key string) (string, error) {
	if val, ok := a[key]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("argument %q is not a string", key)
	}
	return "", fmt.Errorf("required argument %q not found", key)
}

// Integer argument access with flexible type conversion
func (a Arguments) GetInt(key string, defaultValue int) int {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

func (a Arguments) RequireInt(key string) (int, error) {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case int:
			return v, nil
		case float64:
			return int(v), nil
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i, nil
			}
			return 0, fmt.Errorf("argument %q cannot be converted to int", key)
		default:
			return 0, fmt.Errorf("argument %q is not an int", key)
		}
	}
	return 0, fmt.Errorf("required argument %q not found", key)
}

// Float argument access with flexible type conversion
func (a Arguments) GetFloat(key string, defaultValue float64) float64 {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return defaultValue
}

func (a Arguments) RequireFloat(key string) (float64, error) {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f, nil
			}
			return 0, fmt.Errorf("argument %q cannot be converted to float64", key)
		default:
			return 0, fmt.Errorf("argument %q is not a float64", key)
		}
	}
	return 0, fmt.Errorf("required argument %q not found", key)
}

// Boolean argument access with flexible type conversion
func (a Arguments) GetBool(key string, defaultValue bool) bool {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b
			}
		case int:
			return v != 0
		case float64:
			return v != 0
		}
	}
	return defaultValue
}

func (a Arguments) RequireBool(key string) (bool, error) {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case bool:
			return v, nil
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b, nil
			}
			return false, fmt.Errorf("argument %q cannot be converted to bool", key)
		case int:
			return v != 0, nil
		case float64:
			return v != 0, nil
		default:
			return false, fmt.Errorf("argument %q is not a bool", key)
		}
	}
	return false, fmt.Errorf("required argument %q not found", key)
}

// String slice argument access
func (a Arguments) GetStringSlice(key string, defaultValue []string) []string {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case []string:
			return v
		case []interface{}:
			result := make([]string, 0, len(v))
			for _, item := range v {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return defaultValue
}

func (a Arguments) RequireStringSlice(key string) ([]string, error) {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case []string:
			return v, nil
		case []interface{}:
			result := make([]string, 0, len(v))
			for i, item := range v {
				if str, ok := item.(string); ok {
					result = append(result, str)
				} else {
					return nil, fmt.Errorf("item %d in argument %q is not a string", i, key)
				}
			}
			return result, nil
		default:
			return nil, fmt.Errorf("argument %q is not a string slice", key)
		}
	}
	return nil, fmt.Errorf("required argument %q not found", key)
}

// Int slice argument access
func (a Arguments) GetIntSlice(key string, defaultValue []int) []int {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case []int:
			return v
		case []interface{}:
			result := make([]int, 0, len(v))
			for _, item := range v {
				switch num := item.(type) {
				case int:
					result = append(result, num)
				case float64:
					result = append(result, int(num))
				case string:
					if i, err := strconv.Atoi(num); err == nil {
						result = append(result, i)
					}
				}
			}
			return result
		}
	}
	return defaultValue
}

func (a Arguments) RequireIntSlice(key string) ([]int, error) {
	if val, ok := a[key]; ok {
		switch v := val.(type) {
		case []int:
			return v, nil
		case []interface{}:
			result := make([]int, 0, len(v))
			for i, item := range v {
				switch num := item.(type) {
				case int:
					result = append(result, num)
				case float64:
					result = append(result, int(num))
				case string:
					if i, err := strconv.Atoi(num); err == nil {
						result = append(result, i)
					} else {
						return nil, fmt.Errorf("item %d in argument %q cannot be converted to int", i, key)
					}
				default:
					return nil, fmt.Errorf("item %d in argument %q is not an int", i, key)
				}
			}
			return result, nil
		default:
			return nil, fmt.Errorf("argument %q is not an int slice", key)
		}
	}
	return nil, fmt.Errorf("required argument %q not found", key)
}

// Has checks if an argument exists
func (a Arguments) Has(key string) bool {
	_, ok := a[key]
	return ok
}

// Keys returns all argument keys
func (a Arguments) Keys() []string {
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	return keys
}
