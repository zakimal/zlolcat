package main

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"regexp"
	"time"
)

var stripAnsiStart = regexp.MustCompile("^\033" + `\[(\d+)(;\d+)?(;\d+)?[m|K]`)

type LolWriter struct {
	base   io.Writer
	os     int
	li     int
	spread float64
	freq   float64
}

var tabSpaces = []byte("    ")

func (w *LolWriter) Write(data []byte) (int, error) {
	for i := 0; i < len(data); i++ {
		c := data[i]
		if c == '\n' {
			w.li = 0
			w.os++
			w.base.Write([]byte{'\n'})
		} else if c == '\t' {
			w.li += len(tabSpaces)
			w.base.Write(tabSpaces)
		} else {
			matchPos := stripAnsiStart.FindIndex(data[i:])
			if matchPos != nil {
				i += matchPos[1] - 1
				continue
			}
			r, g, b := rainbow(w.freq, float64(w.os)+(float64(w.li)/w.spread))
			fmt.Fprint(w.base, colored(string(c), r, g, b))
			w.li++
		}
	}
	return len(data), nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	seed := int(rand.Int31n(256))
	if len(os.Args) == 1 {
		runLol(seed, os.Stdout, os.Stdin)
		return
	}

	var exit int
	for _, filename := range os.Args[1:] {
		f, err := os.Open(filename)
		if err != nil {
			exit = 1
			fmt.Printf("%s: %s\n", os.Args[0], err.Error())
			continue
		}
		defer f.Close()
		runLol(seed, os.Stdout, f)
	}
	os.Exit(exit)
}

func runLol(seed int, output io.Writer, input io.Reader) {
	defer func() {
		if r := recover(); r != nil {
			io.Copy(output, input)
		}
	}()

	writer := LolWriter{
		os:     seed,
		base:   output,
		spread: 3.0,
		freq:   1.0,
	}

	cat(&writer, input)
}

func cat(writer io.Writer, reader io.Reader) {
	io.Copy(writer, reader)
}

func rainbow(freq, i float64) (r int, g int, b int) {
	r = int(math.Sin(freq*i+0)*127 + 128)
	g = int(math.Sin(freq*i+2*math.Pi/3)*127 + 128)
	b = int(math.Sin(freq*i+4*math.Pi/3)*127 + 128)
	return r, g, b
}

func colored(str string, r, g, b int) string {
	return fmt.Sprintf("\033[38%sm%s\033[0m", rgb(float64(r), float64(g), float64(b)), str)
}

func rgb(r, g, b float64) string {
	grayPossible := true
	sep := 42.5
	var isGray bool
	for grayPossible {
		if r < sep || g < sep || b < sep {
			isGray = r < sep && g < sep && b < sep
			grayPossible = false
		}
		sep += 42.5
	}

	if isGray {
		return fmt.Sprintf(";5;%d", 232+int((r+g+b)/33.0))
	} else {
		value := 16 + toBaseColor(r, 36) + toBaseColor(g, 6) + toBaseColor(b, 1)
		return fmt.Sprintf(";5;%d", value)
	}
}

func toBaseColor(color float64, mod int) int {
	return int(6*(color/256)) * mod
}
