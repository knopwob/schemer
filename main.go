package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/jqln-0/colorshow"
)

func loadImage(filepath string) image.Image {
	infile, err := os.Open(filepath)
	if err != nil {
		log.Panic(err)
	}
	defer infile.Close()

	src, _, err := image.Decode(infile)
	if err != nil {
		log.Panic(err)
	}
	return src
}

func abs(n int) int {
	if n >= 0 {
		return n
	}
	return -n
}

func colorDifference(col1 color.Color, col2 color.Color, threshold int) bool {
	c1 := col1.(color.NRGBA)
	c2 := col2.(color.NRGBA)

	rDiff := abs(int(c1.R) - int(c2.R))
	gDiff := abs(int(c1.G) - int(c2.G))
	bDiff := abs(int(c1.B) - int(c2.B))

	total := rDiff + gDiff + bDiff
	return total >= threshold
}

func getDistinctColors(colors []color.Color, threshold int, minBrightness, maxBrightness int) []color.Color {
	distinctColors := make([]color.Color, 0)
	for _, c := range colors {
		same := false
		if !colorDifference(c, color.NRGBAModel.Convert(color.Black), minBrightness*3) {
			continue
		}
		if !colorDifference(c, color.NRGBAModel.Convert(color.White), (255-maxBrightness)*3) {
			continue
		}
		for _, k := range distinctColors {
			if !colorDifference(c, k, threshold) {
				same = true
				break
			}
		}
		if !same {
			distinctColors = append(distinctColors, c)
		}
	}
	return distinctColors
}
func usage() {
	fmt.Fprintln(os.Stderr, "Usage: colorscheme [flags] -term=\"terminal format\" imagepath")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	terminalSupport := "Terminal format to output colors as. Currently supported: \n"
	for _, t := range terminals {
		terminalSupport += strings.Join([]string{"    ", t.friendlyName, ":", t.flagName, "\n"}, " ")
	}
	var (
		threshold     = flag.Int("t", 50, "Threshold for minimum color difference")
		display       = flag.Bool("d", false, "Display colors in SDL window")
		terminal      = flag.String("term", "default", terminalSupport)
		minBrightness = flag.Int("minBright", 50, "Minimum brightness for colors")
		maxBrightness = flag.Int("maxBright", 200, "Maximum brightness for colors")
		debug         = flag.Bool("debug", false, "Show debugging messages")
	)
	flag.Usage = usage
	flag.Parse()
	if *minBrightness > 255 || *maxBrightness > 255 {
		fmt.Print("Minimum and maximum brightness must be an integer between 0 and 255.\n")
		os.Exit(2)
	}
	if *threshold > 255 {
		fmt.Print("Threshold should be an integer between 0 and 255.\n")
		os.Exit(2)
	}
	if !*debug {
		log.SetOutput(ioutil.Discard)
	}

	// Load the image and create array of colors
	fuzzyness := 5
	img := loadImage(flag.Args()[0])
	w, h := img.Bounds().Max.X, img.Bounds().Max.Y
	colors := make([]color.Color, 0, w*h)
	for x := 0; x < w; x += fuzzyness {
		for y := 0; y < h; y += fuzzyness {
			col := color.NRGBAModel.Convert(img.At(x, y))
			colors = append(colors, col)
		}
	}
	// Get the distinct colors from the array by comparing differences with a threshold
	distinctColors := getDistinctColors(colors, *threshold, *minBrightness, *maxBrightness)

	// Ensure there are 16 colors
	count := 0
	for len(distinctColors) < 16 {
		count++
		distinctColors = append(distinctColors, getDistinctColors(colors, *threshold-count, *minBrightness, *maxBrightness)...)
		if count == *threshold {
			fmt.Print("Could not get colors from image with settings specified. Aborting.\n")
			os.Exit(1)
		}
	}

	if len(distinctColors) > 16 {
		distinctColors = distinctColors[:16]
	}

	if *display {
		colorshow.DisplaySwatches(distinctColors)
	}

	// Output the configuration specified
	terminalMatch := false
	for _, t := range terminals {
		if *terminal == t.flagName {
			fmt.Print(t.output(distinctColors))
			terminalMatch = true
			break
		}
	}
	if !terminalMatch {
		fmt.Printf("Did not recognise format %v. \n", *terminal)
	}
}
