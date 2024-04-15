package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"github.com/otiai10/gosseract/v2"
	"github.com/urfave/cli/v2"
)

type config struct {
	imagePath string
	pdfPath   string
}

func main() {

	config := &config{}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "extract-contribution-images",
				Aliases: []string{"a"},
				Usage:   "add a task to the list",
				Action:  extractContributionFiles(config),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "pdf-file-path",
						Usage:       "Specifies the path to the contributions pdf to split images out",
						Destination: &config.pdfPath,
						Value:       "",
					},
					&cli.StringFlag{
						Name:        "image-output-path",
						Usage:       "Specifies the dir to write the images to",
						Destination: &config.imagePath,
						Value:       "",
					},
				},
			},
			{
				Name:    "parse-image",
				Aliases: []string{"c"},
				Usage:   "complete a task on the list",
				Action:  parseImage(config),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "image-file-path",
						Usage:       "Specifies the path to the contribution image to parse",
						Destination: &config.imagePath,
						Value:       "",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// TODO: Need to modify this to take in a directory of images instead of a single image and then
// exact the content from a CSV file.
//
// The problem is that the OCR content is inconsistent so you do not always get consistent text to
// parse on.
func parseImage(config *config) func(c *cli.Context) error {
	return func(c *cli.Context) error {

		// Instantiate an instance of go tesseract library
		client := gosseract.NewClient()
		defer client.Close()

		// Specify that the lange for hte client is english.
		err := client.SetLanguage("eng")
		if err != nil {
			return err
		}

		// Specify the image you want to. Change image here.
		client.SetImage(config.imagePath)

		// OCR the text
		text, err := client.Text()
		if err != nil {
			return err
		}

		// Parse text into an IOReader and than into a scanner to go line by line.
		reader := strings.NewReader(text)
		scanner := bufio.NewScanner(reader)

		// Iterate over each line and parse out the text you you need. This is missing a lot, but
		// the more complicates stuff like extracting OCR content from an image is done. This needs
		// to be cleaned up.
		line := 0
		for scanner.Scan() {

			// Parse the 6th line as that is the person being donated to.
			if line == 6 {
				fmt.Printf("Line %d: "+scanner.Text()+"\n", line)
			}

			// Parse out lines that start with a date as those are contribution lines.
			_, err := time.Parse("01/02/2006", strings.Split(scanner.Text(), " ")[0])
			if err == nil {
				fmt.Printf("Line %d: "+scanner.Text()+"\n", line)
			}

			// Uncomment to see all object sand their line number.
			fmt.Printf("Line %d: "+scanner.Text()+"\n", line)

			line++
		}

		return nil
	}
}

// TODO: This does not yet handle if the directory already exists.
func extractContributionFiles(config *config) func(c *cli.Context) error {
	return func(c *cli.Context) error {

		newDirPath := filepath.Join(config.imagePath)

		err := os.MkdirAll(newDirPath, os.ModePerm)
		if err != nil {
			return err
		}

		err = api.ExtractImagesFile(config.pdfPath, config.imagePath, nil, nil)
		if err != nil {
			return err
		}
		return nil
	}
}
