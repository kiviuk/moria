package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kiviuk/pwdgen/internal/app"
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
		if flag == "--magic" || flag == "--pretty" || flag == "--live" {
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

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: pwdgen [--magic|--pretty|--live] [--max-len N] <spell>\n")
		fmt.Fprintf(os.Stderr, "  --magic    Generate a 300-character master password\n")
		fmt.Fprintf(os.Stderr, "  --pretty   Display the password matrix from your master password\n")
		fmt.Fprintf(os.Stderr, "  --live     Interactive mode: type your spell and see the password build in real-time\n")
		fmt.Fprintf(os.Stderr, "  --max-len  Truncate output to N characters (live and batch modes only)\n")
		fmt.Fprintf(os.Stderr, "  <spell>    Generate a service password from your spell\n")
		os.Exit(1)
	}

	cfg, flags, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := validateConfig(cfg, flags); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	mode := modes[cfg.Mode]

	if mode.NeedsStdin {
		cfg.Master = readMasterPassword()
		expectedLen := app.PasswordMatrixRows * app.PasswordMatrixColumns * app.CharactersPerMatrixCell
		if len(cfg.Master) != expectedLen {
			fmt.Fprintf(os.Stderr, "Master password must be %d characters (got %d)\n", expectedLen, len(cfg.Master))
			os.Exit(1)
		}
	}

	switch cfg.Mode {
	case "magic":
		master, err := app.GenerateRandomString(app.PasswordMatrixRows*app.PasswordMatrixColumns*app.CharactersPerMatrixCell, app.MasterPasswordChars)
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
		password, err := LiveMode(matrix, cfg.MaxLen)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Live mode error: %v\n", err)
			os.Exit(1)
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

func readMasterPassword() string {
	stat, _ := os.Stdin.Stat()
	isPipe := (stat.Mode() & os.ModeCharDevice) == 0

	var master string
	if !isPipe {
		fmt.Print("Enter master password: ")
	}
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		master = strings.TrimSpace(scanner.Text())
	}
	return master
}
