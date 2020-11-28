package common

import(
    "os"
)

// GetEnvDefault is using for getting enviroment variable with default value
func GetEnvDefault(key string, defVal string) string {
    val, ex := os.LookupEnv(key)
    if !ex {
        return defVal
    }
    return val
}
