package config

import (
	"log"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

func fatalGetString(key string) string {
	checkKey(key)
	value := os.Getenv(key)
	if value == "" {
		value = viper.GetString(key)
	}
	return value
}

func getString(key string) string {
	value := os.Getenv(key)
	if value == "" {
		value = viper.GetString(key)
	}
	return value
}

func getFeature(key string) bool {
	v, err := strconv.ParseBool(fatalGetString(key))
	if err != nil {
		return false
	}
	return v
}

func checkKey(key string) {
	if !viper.IsSet(key) && os.Getenv(key) == "" {
		log.Fatalf("%s key is not set", key)
	}
}

type EnvVar struct {
	Present bool
	Key     string
	Value   string
}

func NewEnvVar(key, value string) EnvVar {
	return EnvVar{
		Present: true,
		Key:     key,
		Value:   value,
	}
}

func MissingEnvVar(key string) EnvVar {
	return EnvVar{
		Present: false,
		Key:     key,
	}
}

type EnvVarMutation struct {
	previousValues []EnvVar
}

func (e EnvVarMutation) Rollback() {

}

func ApplyEnvVars(vars ...EnvVar) EnvVarMutation {
	var oldEnvVars []EnvVar

	for _, envVar := range vars {
		oldValue, exists := os.LookupEnv(envVar.Key)

		oldEnvVars = append(oldEnvVars, EnvVar{
			Present: exists,
			Key:     envVar.Key,
			Value:   oldValue,
		})

		if envVar.Present {
			Must(os.Setenv(envVar.Key, envVar.Value))
		} else {
			Must(os.Unsetenv(envVar.Key))
		}
	}

	return EnvVarMutation{previousValues: oldEnvVars}
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
