package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Global flags
var Verbose bool
var Coords bool
var boxBorderPadding = 1
var paddingBetweenX = 5
var paddingBetweenY = 5
var graphDirection = "LR"
var useAscii = false
var headStyle = HeadStylePointer

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mermaid-ascii",
	Short: "Generate ASCII diagrams from mermaid code.",
	Run: func(cmd *cobra.Command, args []string) {
		if Verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		var mermaid []byte
		var err error

		filePath := cmd.Flag("file").Value.String()
		if filePath == "" || filePath == "-" {
			// Read from stdin
			mermaid, err = io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal("Failed to read from stdin: ", err)
				return
			}
		} else {
			// Read from file
			mermaid, err = os.ReadFile(filePath)
			if err != nil {
				log.Fatal("Failed to read mermaid file: ", err)
				return
			}
		}

		paddingX := paddingBetweenX
		paddingY := paddingBetweenY
		if _, err := lib.Render(string(mermaid), lib.Options{
			UseASCII: useAscii,
			PaddingX: &paddingX,
			PaddingY: &paddingY,
			Heads:    headStyleString(headStyle),
		}); err != nil {
			log.Fatal("Failed to render mermaid input: ", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&useAscii, "ascii", "a", false, "Don't use extended character set")
	rootCmd.PersistentFlags().BoolVarP(&Coords, "coords", "c", false, "Show coordinates")
	rootCmd.PersistentFlags().IntVarP(&paddingBetweenX, "paddingX", "x", paddingBetweenX, "Horizontal space between nodes")
	rootCmd.PersistentFlags().IntVarP(&paddingBetweenY, "paddingY", "y", paddingBetweenY, "Vertical space between nodes")
	rootCmd.PersistentFlags().IntVarP(&boxBorderPadding, "borderPadding", "p", boxBorderPadding, "Padding between text and border")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("file", "f", "", "Mermaid file to parse (use '-' for stdin)")

	lib.RegisterRenderer(renderMermaidWithOptions)
}

func renderMermaidWithOptions(mermaid string, opts lib.Options) (string, error) {
	originalUseASCII := useAscii
	originalPaddingX := paddingBetweenX
	originalPaddingY := paddingBetweenY
	originalHeadStyle := headStyle
	defer func() {
		useAscii = originalUseASCII
		paddingBetweenX = originalPaddingX
		paddingBetweenY = originalPaddingY
		headStyle = originalHeadStyle
	}()

	useAscii = opts.UseASCII
	if opts.PaddingX != nil {
		paddingBetweenX = *opts.PaddingX
	}
	if opts.PaddingY != nil {
		paddingBetweenY = *opts.PaddingY
	}
	nextHeadStyle, err := parseHeadStyle(opts.Heads)
	if err != nil {
		return "", err
	}
	headStyle = nextHeadStyle

	properties, err := mermaidFileToMap(mermaid, "cli")
	if err != nil {
		return "", err
	}

	result := drawMap(properties)
	return result, nil
}

func parseHeadStyle(value string) (HeadStyle, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "pointer":
		return HeadStylePointer, nil
	case "open":
		return HeadStyleOpen, nil
	case "none":
		return HeadStyleNone, nil
	case "wide":
		return HeadStyleWide, nil
	default:
		return HeadStylePointer, fmt.Errorf("unknown head style %q", value)
	}
}

func headStyleString(style HeadStyle) string {
	switch style {
	case HeadStyleOpen:
		return "open"
	case HeadStyleNone:
		return "none"
	case HeadStyleWide:
		return "wide"
	default:
		return "pointer"
	}
}
