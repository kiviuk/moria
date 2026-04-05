package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/kiviuk/moria/internal/app"
)

type Config struct {
	Mode   string
	Spell  string
	MaxLen int
	Master string
}

type Mode struct {
	Name        string
	NeedsStdin  bool
	NeedsSpell  bool
	AllowedMods []string
}

var modes = map[string]Mode{
	"magic": {
		Name:        "magic",
		NeedsStdin:  false,
		NeedsSpell:  false,
		AllowedMods: nil,
	},
	"pretty": {
		Name:        "pretty",
		NeedsStdin:  true,
		NeedsSpell:  false,
		AllowedMods: nil,
	},
	"live": {
		Name:        "live",
		NeedsStdin:  true,
		NeedsSpell:  false,
		AllowedMods: []string{"--max-len"},
	},
	"batch": {
		Name:        "batch",
		NeedsStdin:  true,
		NeedsSpell:  true,
		AllowedMods: []string{"--max-len"},
	},
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseArgs(args []string) (Config, map[string]bool, error) {
	cfg := Config{Mode: "batch"}
	flags := make(map[string]bool)
	var positional []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--magic":
			flags["--magic"] = true
			cfg.Mode = "magic"
		case "--pretty":
			flags["--pretty"] = true
			cfg.Mode = "pretty"
		case "--live":
			flags["--live"] = true
			cfg.Mode = "live"
		case "--max-len":
			flags["--max-len"] = true
			if i+1 >= len(args) {
				return cfg, flags, fmt.Errorf("--max-len requires a value")
			}
			i++
			val, err := strconv.Atoi(args[i])
			if err != nil {
				return cfg, flags, fmt.Errorf("--max-len value must be a number")
			}
			cfg.MaxLen = val
		case "--help", "-h":
			flags["--help"] = true
		default:
			positional = append(positional, args[i])
		}
	}

	if len(positional) > 0 {
		cfg.Spell = positional[0]
	}

	return cfg, flags, nil
}

func validateConfig(cfg Config, flags map[string]bool) error {
	mode, ok := modes[cfg.Mode]
	if !ok {
		return fmt.Errorf("unknown mode: %s", cfg.Mode)
	}

	for flag, present := range flags {
		if !present {
			continue
		}
		if flag == "--magic" || flag == "--pretty" || flag == "--live" || flag == "--help" {
			continue
		}
		if !contains(mode.AllowedMods, flag) {
			return fmt.Errorf("%s not allowed in %s mode", flag, mode.Name)
		}
	}

	if mode.NeedsSpell && cfg.Spell == "" {
		return fmt.Errorf("%s mode requires a spell", mode.Name)
	}

	return nil
}

func printUsage() {
	fmt.Println("moria — deterministic password generator")
	fmt.Println()
	fmt.Println("Usage: moria [--magic|--pretty|--live] [--max-len N] <spell>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --magic    Generate a master password")
	fmt.Println("  --pretty   Display the password matrix from your master password")
	fmt.Println("  --live     Interactive mode: type your spell and see the password build in real-time")
	fmt.Println("  --max-len  Truncate output to N characters (live and batch modes only)")
	fmt.Println("  -h, --help Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  moria --magic                      # Generate a new master password")
	fmt.Println("  moria \"amazon\"                     # Generate password for Amazon")
	fmt.Println("  cat master.txt | moria \"amazon\"     # Piped from password manager")
	fmt.Println("  cat master.txt | moria --pretty     # Display the matrix")
	fmt.Println("  cat master.txt | moria --live       # Interactive mode")
	fmt.Println("  cat master.txt | moria --max-len 16 \"amazon\"  # Limited length")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	cfg, flags, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if flags["--help"] {
		printUsage()
		os.Exit(0)
	}

	if err := validateConfig(cfg, flags); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	mode := modes[cfg.Mode]

	if mode.NeedsStdin {
		stat, err := os.Stdin.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not stat stdin: %v\n", err)
		}
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		if isPiped {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read master password from pipe: %v\n", err)
				os.Exit(1)
			}
			cfg.Master = strings.TrimSpace(string(data))
		} else {
			master, err := getPassword()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			cfg.Master = master
		}
		cfg.Master = app.ExpandToMatrix(cfg.Master)
	}

	switch cfg.Mode {
	case "magic":
		master, err := app.GenerateMasterPassword(app.MatrixBytes, app.MasterPasswordChars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate master password: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(master)

	case "pretty":
		matrix, err := app.NewMatrix(cfg.Master)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create matrix: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(matrix.Pretty())

	case "live":
		matrix, err := app.NewMatrix(cfg.Master)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create matrix: %v\n", err)
			os.Exit(1)
		}
		finalModel, err := LiveMode(matrix, cfg.MaxLen)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Live mode error: %v\n", err)
			os.Exit(1)
		}
		password := finalModel.password
		if cfg.MaxLen > 0 && len(password) > cfg.MaxLen {
			password = password[:cfg.MaxLen]
		}
		if password != "" {
			fmt.Print(password)
		}

	case "batch":
		matrix, err := app.NewMatrix(cfg.Master)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create matrix: %v\n", err)
			os.Exit(1)
		}
		dirty := app.DirtySpell{Spell: cfg.Spell}
		magic, err := dirty.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid spell: %v\n", err)
			os.Exit(1)
		}
		password, err := magic.ExtractPassword(matrix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract password: %v\n", err)
			os.Exit(1)
		}
		if cfg.MaxLen > 0 && len(password) > cfg.MaxLen {
			password = password[:cfg.MaxLen]
		}
		fmt.Print(password)
	}
}
