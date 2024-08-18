package env

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

func GetRequiredEnvString(name string, logger *slog.Logger) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		err := fmt.Errorf("env variable %s not set", name)

		logger.Error(err.Error())
		panic(err)
	}

	return value
}

func GetEnvInt(name string, defaultValue int, logger *slog.Logger) int {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		err := fmt.Errorf("could not parse %s", name)

		logger.Error(err.Error())
		panic(err)
	}

	return parsedValue
}

func GetRequiredEnvDuration(name string, logger *slog.Logger) time.Duration {
	value := GetRequiredEnvString(name, logger)

	parsedValue, err := time.ParseDuration(value)
	if err != nil {
		err := fmt.Errorf("could not parse %s", name)

		logger.Error(err.Error())
		panic(err)
	}

	return parsedValue
}
