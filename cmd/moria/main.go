package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/kiviuk/moria/internal/app"
)

type Mode int

const (
	ModeBatch Mode = iota
	ModeMagic
	ModePretty
	ModeLive
)

func (m Mode) String() string {
	return [...]string{"batch", "magic", "pretty", "live"}[m]
}

func (m Mode) needsStdin() bool {
	return m == ModePretty || m == ModeLive || m == ModeBatch
}

func (m Mode) needsSpell() bool {
	return m == ModeBatch
}

func (m Mode) allowedMods() []string {
	switch m {
	case ModeLive:
		return []string{"--max-len", "--ignore-paste"}
	case ModeBatch:
		return []string{"--max-len"}
	default:
		return nil
	}
}

type Config struct {
	Mode   Mode
	Spell  string
	MaxLen int
	Master string
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
	cfg := Config{Mode: ModeBatch}
	flags := make(map[string]bool)
	var positional []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--magic":
			flags["--magic"] = true
			cfg.Mode = ModeMagic
		case "--pretty":
			flags["--pretty"] = true
			cfg.Mode = ModePretty
		case "--live":
			flags["--live"] = true
			cfg.Mode = ModeLive
		case "--max-len":
			flags["--max-len"] = true
			if i+1 >= len(args) {
				return cfg, flags, fmt.Errorf(ErrMaxLenRequiresValue)
			}
			i++
			val, err := strconv.Atoi(args[i])
			if err != nil {
				return cfg, flags, fmt.Errorf(ErrMaxLenNotNumber)
			}
			cfg.MaxLen = val
		case "--ignore-paste":
			flags["--ignore-paste"] = true
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
	for flag, present := range flags {
		if !present {
			continue
		}
		if flag == "--magic" || flag == "--pretty" || flag == "--live" || flag == "--help" {
			continue
		}
		if !contains(cfg.Mode.allowedMods(), flag) {
			return fmt.Errorf(ErrModNotAllowed, flag, cfg.Mode)
		}
	}

	if cfg.Mode.needsSpell() && cfg.Spell == "" {
		return fmt.Errorf(ErrSpellRequired, cfg.Mode)
	}

	return nil
}

func printUsage() {
	fmt.Println("moria — deterministic password generator")
	fmt.Println()
	fmt.Println("Usage: moria [--magic|--pretty|--live] [--max-len N] [--ignore-paste] <spell>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --magic          Generate a master password")
	fmt.Println("  --pretty         Display the password matrix from your master password")
	fmt.Println("  --live           Interactive mode: type your spell and see the password build in real-time")
	fmt.Println("  --max-len        Truncate generated output to N characters (live and batch modes only)")
	fmt.Println("  --ignore-paste   Ignore pasted input in live mode (single characters only, live mode only)")
	fmt.Println("  -h, --help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  moria --magic                              # Generate a new master password")
	fmt.Println("  moria \"amazon\"                             # Generate password for Amazon")
	fmt.Println("  cat master.txt | moria \"amazon\"             # Piped from password manager")
	fmt.Println("  cat master.txt | moria --pretty             # Display the matrix")
	fmt.Println("  cat master.txt | moria --live               # Interactive mode (paste allowed)")
	fmt.Println("  cat master.txt | moria --live --ignore-paste # Interactive mode (paste blocked)")
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

	if cfg.Mode.needsStdin() {
		stat, err := os.Stdin.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not stat stdin: %v\n", err)
		}
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		if isPiped {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, ErrFailedReadMaster+"\n", err)
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
	case ModeMagic:
		master, err := app.GenerateMasterPassword(app.MatrixBytes, app.MasterPasswordChars)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrFailedGenerateMaster+"\n", err)
			os.Exit(1)
		}
		fmt.Print(master)

	case ModePretty:
		matrix, err := app.NewMatrix(cfg.Master)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrFailedCreateMatrix+"\n", err)
			os.Exit(1)
		}
		fmt.Print(matrix.Pretty())

	case ModeLive:
		matrix, err := app.NewMatrix(cfg.Master)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrFailedCreateMatrix+"\n", err)
			os.Exit(1)
		}
		pasteMode := PasteAllowed
		if flags["--ignore-paste"] {
			pasteMode = PasteIgnored
		}
		finalModel, err := LiveMode(matrix, cfg.MaxLen, pasteMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrLiveMode+"\n", err)
			os.Exit(1)
		}
		password := finalModel.password
		if cfg.MaxLen > 0 && len(password) > cfg.MaxLen {
			password = password[:cfg.MaxLen]
		}
		if password != "" {
			fmt.Print(password)
		}

	case ModeBatch:
		matrix, err := app.NewMatrix(cfg.Master)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrFailedCreateMatrix+"\n", err)
			os.Exit(1)
		}
		dirty := app.DirtySpell{Spell: cfg.Spell}
		magic, err := dirty.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrInvalidSpell+"\n", err)
			os.Exit(1)
		}
		password, err := magic.ExtractPassword(matrix)
		if err != nil {
			fmt.Fprintf(os.Stderr, ErrExtractPassword+"\n", err)
			os.Exit(1)
		}
		if cfg.MaxLen > 0 && len(password) > cfg.MaxLen {
			password = password[:cfg.MaxLen]
		}
		fmt.Print(password)
	}
}
