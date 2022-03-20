package main

import (
	"bufio"
	_ "embed"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/go-wordwrap"
	"github.com/tmck-code/pokesay-go/src/pokedex"
)

var (
	//go:embed build/cows.gob
	data []byte
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func printSpeechBubbleLine(line string, width int) {
	if len(line) > width {
		fmt.Println("|", line)
	} else {
		fmt.Println("|", line, strings.Repeat(" ", width-len(line)), "|")
	}
}

func printWrappedText(line string, width int, tabSpaces string) {
	for _, wline := range strings.Split(wordwrap.WrapString(strings.Replace(line, "\t", tabSpaces, -1), uint(width)), "\n") {
		printSpeechBubbleLine(wline, width)
	}
}

func printSpeechBubble(scanner *bufio.Scanner, args Args) {
	border := strings.Repeat("-", args.Width+2)
	fmt.Println("/" + border + "\\")

	for scanner.Scan() {
		line := scanner.Text()

		if !args.NoTabSpaces {
			line = strings.Replace(line, "\t", args.TabSpaces, -1)
		}
		if args.NoWrap {
			printSpeechBubbleLine(line, args.Width)
		} else {
			printWrappedText(line, args.Width, args.TabSpaces)
		}
	}
	fmt.Println("\\" + border + "/")
	for i := 0; i < 4; i++ {
		fmt.Println(strings.Repeat(" ", i+8), "\\")
	}
}

func randomInt(n int) int {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n) + 1
}

func printPokemon(choice pokedex.PokemonEntry) {
	binary.Write(os.Stdout, binary.LittleEndian, pokedex.Decompress(choice.Data))
	fmt.Printf("choice: %s / categories: %s\n", choice.Name, choice.Categories)

}

func chooseRandomCategory(entries pokedex.PokemonEntryMap) []pokedex.PokemonEntry {
	choice := 3 // randomInt(entries.NCategories)
	idx := 0
	for category, _ := range entries.Categories {
		if idx == choice {
			return entries.Categories[category]
		}
		idx++
	}
	return entries.Categories["regular"]
}

func chooseRandomPokemon(pokemon []pokedex.PokemonEntry) pokedex.PokemonEntry {
	return pokemon[randomInt(len(pokemon))]
}

type Args struct {
	Width          int
	NoWrap         bool
	TabSpaces      string
	NoTabSpaces    bool
	ListCategories bool
	Category       string
}

func parseFlags() Args {
	width := flag.Int("width", 80, "the max speech bubble width")
	noWrap := flag.Bool("nowrap", false, "disable text wrapping (fastest)")
	tabWidth := flag.Int("tabwidth", 4, "replace any tab characters with N spaces")
	noTabSpaces := flag.Bool("notabspaces", false, "do not replace tab characters (fastest)")
	fastest := flag.Bool("fastest", false, "run with the fastest possible configuration (-nowrap -notabspaces)")
	listCategories := flag.Bool("categories-list", false, "list all available categories")
	category := flag.String("category", "", "list all available categories")

	flag.Parse()
	var args Args

	if *fastest {
		args = Args{
			Width:       *width,
			NoWrap:      true,
			TabSpaces:   "    ",
			NoTabSpaces: true,
		}
	} else {
		args = Args{
			Width:          *width,
			NoWrap:         *noWrap,
			TabSpaces:      strings.Repeat(" ", *tabWidth),
			NoTabSpaces:    *noTabSpaces,
			ListCategories: *listCategories,
			Category:       *category,
		}
	}
	return args
}

func main() {
	args := parseFlags()
	pokemon := pokedex.ReadFromBytes(data)

	if args.ListCategories {
		for k, v := range pokemon.Categories {
			fmt.Printf("%s (%d)\n", k, len(v))
		}
		os.Exit(0)
	}
	printSpeechBubble(bufio.NewScanner(os.Stdin), args)

	if len(args.Category) < 0 {
		printPokemon(chooseRandomPokemon(pokemon.Categories[args.Category]))
	} else {
		printPokemon(chooseRandomPokemon(chooseRandomCategory(pokemon)))
	}
}
