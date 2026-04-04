package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kiviuk/pwdgen/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: pwdgen [--magic|--pretty] <spell>\n")
		fmt.Fprintf(os.Stderr, "  --magic    Generate a 300-character master password\n")
		fmt.Fprintf(os.Stderr, "  --pretty   Display the password matrix from your master password\n")
		fmt.Fprintf(os.Stderr, "  <spell>    Generate a service password from your spell\n")
		os.Exit(1)
	}

	if os.Args[1] == "--magic" {
		master, err := app.GenerateRandomString(app.PasswordMatrixRows*app.PasswordMatrixColumns*app.CharactersPerMatrixCell, app.MasterPasswordChars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate master password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(master)
		return
	}

	if os.Args[1] == "--pretty" {
		master := readMasterPassword()
		expectedLen := app.PasswordMatrixRows * app.PasswordMatrixColumns * app.CharactersPerMatrixCell
		if len(master) != expectedLen {
			fmt.Fprintf(os.Stderr, "Master password must be %d characters (got %d)\n", expectedLen, len(master))
			os.Exit(1)
		}
		matrix, err := app.NewMatrix(master)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create matrix: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(matrix.Pretty())
		return
	}

	spell := os.Args[1]

	master := readMasterPassword()

	expectedLen := app.PasswordMatrixRows * app.PasswordMatrixColumns * app.CharactersPerMatrixCell
	if len(master) != expectedLen {
		fmt.Fprintf(os.Stderr, "Master password must be %d characters (got %d)\n", expectedLen, len(master))
		os.Exit(1)
	}

	matrix, err := app.NewMatrix(master)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create matrix: %v\n", err)
		os.Exit(1)
	}

	dirty := app.DirtySpell{Spell: spell}
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

	fmt.Println(password)
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
