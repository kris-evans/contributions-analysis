package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/otiai10/gosseract/v2"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type config struct {
	isDebug    bool
	imagePath  string
	imageDir   string
	pdfPath    string
	outputPath string
}

func main() {

	config := &config{}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "extract-contribution-images",
				Usage:  "add a task to the list",
				Action: extractContributionFiles(config),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "debug",
						Aliases:     []string{"v"},
						Usage:       "print debug content",
						Destination: &config.isDebug,
					},
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
				Name:   "parse-image",
				Usage:  "complete a task on the list",
				Action: parseImage(config),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "debug",
						Aliases:     []string{"v"},
						Usage:       "print debug content",
						Destination: &config.isDebug,
					},
					&cli.StringFlag{
						Name:        "image-file-path",
						Usage:       "Specifies the path to the contribution image to parse",
						Destination: &config.imagePath,
						Value:       "",
					},
				},
			},
			{
				Name:   "parse-images-dir",
				Usage:  "complete a task on the list",
				Action: parseImageDirectory(config),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "debug",
						Aliases:     []string{"v"},
						Usage:       "print debug content",
						Destination: &config.isDebug,
					},
					&cli.StringFlag{
						Name:        "image-dir",
						Usage:       "Specifies the dir path to the contribution images to parse",
						Destination: &config.imageDir,
						Value:       "",
					},
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						Usage:       "Specifies the output csv path",
						Destination: &config.outputPath,
						Value:       "./output.csv",
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

type Contribution struct {
	Date       string `json:"Date"`
	FirstName  string `json:"FirstName"`
	LastName   string `json:"LastName"`
	Amount     string `json:"Amount"`
	Address    string `json:"Address"`
	Occupation string `json:"Occupation"`
}

func parseImageDirectory(config *config) func(c *cli.Context) error {
	return func(c *cli.Context) error {

		logger, err := zap.NewProduction()
		if err != nil {
			return err
		}

		if config.isDebug {
			logger, err = zap.NewDevelopment()
			if err != nil {
				return err
			}
		}

		csvFile, err := os.Create(config.outputPath)
		if err != nil {
			return err
		}

		writer, err := writeCSVHeader(csvFile)
		if err != nil {
			return nil
		}
		writer.Flush()

		logger.Debug("iterating over files in directory", zap.String("directory", config.imageDir))

		files, err := os.ReadDir(config.imageDir)
		if err != nil {
			return err
		}

		for index, file := range files {
			if index <= 4 {
				continue
			}

			logger.Debug("parsing file", zap.String("file", file.Name()))

			c, err := parseImageForContributions(logger, config.imageDir+"/"+file.Name())
			if err != nil {
				return err
			}

			err = writeCSVLine(writer, c)
			if err != nil {
				return nil
			}
			writer.Flush()
		}

		return nil
	}
}

// TODO: Need to modify this to take in a directory of images instead of a single image and then
// exact the content from a CSV file.
//
// The problem is that the OCR content is inconsistent so you do not always get consistent text to
// parse on.
func parseImage(config *config) func(c *cli.Context) error {
	return func(c *cli.Context) error {

		logger, err := zap.NewProduction()
		if err != nil {
			return err
		}

		if config.isDebug {
			logger, err = zap.NewDevelopment()
			if err != nil {
				return err
			}
		}

		defer logger.Sync()

		contributions, err := parseImageForContributions(logger, config.imagePath)
		if err != nil {
			return err
		}

		out, err := json.MarshalIndent(contributions, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println("")
		fmt.Println("Output:")
		fmt.Println(string(out))

		return nil
	}
}

func writeCSVHeader(csvFile *os.File) (*csv.Writer, error) {
	csvWriter := csv.NewWriter(csvFile)

	// Write Header
	err := csvWriter.Write([]string{"Date", "First Name", "Last Name", "Amount", "Address"})
	if err != nil {
		return nil, err
	}

	return csvWriter, nil
}

func writeCSVLine(writer *csv.Writer, contributions []Contribution) error {
	// Write Lines
	for _, contribution := range contributions {
		err := writer.Write([]string{contribution.Date, contribution.FirstName, contribution.LastName, contribution.Amount, contribution.Address})
		if err != nil {
			return err
		}
	}

	return nil
}

func parseImageForContributions(logger *zap.Logger, imagePath string) ([]Contribution, error) {
	contributions := []Contribution{}

	// Instantiate an instance of go tesseract library
	client := gosseract.NewClient()
	defer client.Close()

	// Specify that the lange for hte client is english.
	err := client.SetLanguage("eng")
	if err != nil {
		return contributions, err
	}

	// Specify the image you want to. Change image here.
	client.SetImage(imagePath)

	// OCR the text
	text, err := client.Text()
	if err != nil {
		return contributions, err
	}

	// Parse text into an IOReader and than into a scanner to go line by line.
	reader := strings.NewReader(text)
	scanner := bufio.NewScanner(reader)

	documentLines := []string{}

	// Iterate over each line and parse out the text you you need. This is missing a lot, but
	// the more complicates stuff like extracting OCR content from an image is done. This needs
	// to be cleaned up.
	line := 0
	for scanner.Scan() {
		logger.Debug("original", zap.Int("line", line), zap.String("content", scanner.Text()))
		documentLines = append(documentLines, scanner.Text())
		line++
	}

	for i, line := range documentLines {

		_, err := time.Parse("01/02/2006", strings.Split(line, " ")[0])
		if err == nil {
			lineCounter := i

			logger.Debug("parsing", zap.Int("line", i), zap.String("content", line))

			contribution := Contribution{}
			contributionsLine := strings.Split(line, " ")
			contribution.Date = contributionsLine[0]
			contribution.LastName = strings.Replace(contributionsLine[1], ",", "", -1)
			contribution.FirstName = strings.Replace(contributionsLine[2], ",", "", -1)
			contribution.Amount = contributionsLine[3]

			lineCounter++
			logger.Debug("Skipping", zap.Int("line", lineCounter), zap.String("content", documentLines[lineCounter]))

			lineCounter++
			logger.Debug("Processing", zap.Int("line", lineCounter), zap.String("content", documentLines[lineCounter]))

			addressText := strings.Replace(documentLines[lineCounter], "|", "", -1)

			lineCounter++
			logger.Debug("Processing", zap.Int("line", lineCounter), zap.String("content", documentLines[lineCounter]))

			addressText = addressText + " " + strings.Replace(documentLines[lineCounter], "|", "", -1)
			contribution.Address = strings.TrimSpace(addressText)

			contributions = append(contributions, contribution)
		} else {
			logger.Debug("skipping", zap.Int("line", i), zap.String("content", line))
		}
	}

	return contributions, nil
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
