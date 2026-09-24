package bootstrap

import (
	"errors"
	"flag"
)

type Options struct {
	Environment string
	EnvFile     string
	Debug       bool
}

func ParseOptions(args []string) (Options, error) {
	flags := flag.NewFlagSet("foundry-stack", flag.ContinueOnError)

	dev := flags.Bool(EnvDevelopment, false, "use .env.dev environment file")
	prod := flags.Bool(EnvProduction, false, "use .env.prod environment file")
	stag := flags.Bool(EnvStaging, false, "use .env.stag environment file")
	debug := flags.Bool("debug", false, "enable debug mode")

	if err := flags.Parse(args); err != nil {
		return Options{}, err
	}
	if flags.NArg() != 0 {
		return Options{}, errors.New("positional arguments are not supported")
	}

	if (*dev && *stag) || (*dev && *prod) || (*stag && *prod) {
		return Options{}, errors.New("specify only one of -dev, -stag, or -prod")
	}

	var environment, envFile string

	switch {
	case *dev:
		environment = EnvDevelopment
		envFile = ".env.dev"
	case *stag:
		environment = EnvStaging
		envFile = ".env.stag"
	case *prod:
		environment = EnvProduction
		envFile = ".env.prod"
	}

	if envFile == "" {
		return Options{}, errors.New("specify one of -dev, -stag, or -prod")
	}

	return Options{
		Environment: environment,
		EnvFile:     envFile,
		Debug:       *debug,
	}, nil
}
