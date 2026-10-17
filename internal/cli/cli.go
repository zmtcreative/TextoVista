package cli

import (
	"fmt"
	"os"
	"strings"

	"texto-vista/internal/utils"

	flag "github.com/spf13/pflag"
)

// Define a struct to hold the parsed command-line arguments
type CliArgs struct {
    InitialFile         *string
    ShowHelp            *bool
    CmdlineOptions      *string
    AppProgName         *string // Store the application name
    AppProgNameWithExt  *string // Store the application name with extension
}

var recognizedFlags = map[string]bool{
	"-f":      true,
	"--file":  true,
	"-h":      true,
	"--help":  true,
	"--":      true,
}

// GetArgs becomes a standalone function that parses flags and returns them.
func GetArgs() (*CliArgs, error) {
    // Use a new, local FlagSet to avoid conflicts with the global one,
    // which is a best practice for robust CLI parsing.
    fs := flag.NewFlagSet("texto-vista", flag.ContinueOnError)

    args := &CliArgs{}

    // Initialize with nil pointers - this indicates no CLI override was provided
    args.ShowHelp = nil
    args.InitialFile = nil

    // Define local variables for flag parsing
    var initialFile string
    var flagHelp bool

    // Define flags on the local FlagSet
    fs.StringVarP(&initialFile, "file", "f", "", "Path to the initial Markdown file")
    fs.BoolVarP(&flagHelp, "help", "h", false, "Show help message")

    // Configure flags that can be used without a value
    fs.Lookup("help").NoOptDefVal = "true"

    // Parse only the supported flags. Unknown flags are intentionally ignored so
    // the GUI app stays tolerant of shell wrappers and launcher-added arguments.
    filteredArgs := []string{}
    if len(os.Args) > 1 {
        filteredArgs = filterSupportedArgs(os.Args[1:])
    }

    // Parse the application's filtered arguments.
    // ContinueOnError prevents the flagset from exiting the app on its own.
    if len(os.Args) > 1 {
        _ = fs.Parse(filteredArgs)
    } else {
        _ = fs.Parse([]string{})
    }

    // Handle non-flag argument as the initial file (positional args take precedence)
    if len(fs.Args()) > 0 {
        initialFile = fs.Args()[0]
    }

    // Set values only if they were explicitly provided or inferred
    if initialFile != "" {
        args.InitialFile = &initialFile
    }

    if flagHelp {
        args.ShowHelp = &flagHelp
    }

    // Build the usage string
    var usageText strings.Builder

    // Get executable base names using shared utility
    appProgNameWithExt, appProgName := utils.GetExecutableBaseName()

    usageText.WriteString("<pre>\n")
    usageText.WriteString(fmt.Sprintf("Usage: %s [options] [filepath]\n\n", appProgName))
	usageText.WriteString("  filepath          Path to the Markdown file to open\n")
	usageText.WriteString("                      (if not specified with --file)\n\n")
    usageText.WriteString("Options:\n")
	usageText.WriteString("  -f &lt;path&gt;      Path to the Markdown file to open\n")
    usageText.WriteString("  --file &lt;path&gt;   \n\n")
    usageText.WriteString("</pre>\n")

    cmdlineOptions := usageText.String()
    args.AppProgName = &appProgName
    args.AppProgNameWithExt = &appProgNameWithExt
    args.CmdlineOptions = &cmdlineOptions


    // Assign the usage function to the global flag set for the main function to call.
    flag.Usage = func() {
        if args.CmdlineOptions != nil {
            fmt.Fprint(os.Stderr, *args.CmdlineOptions)
        }
    }

    return args, nil
}

func filterSupportedArgs(args []string) []string {
    filtered := make([]string, 0, len(args))

    for i := 0; i < len(args); i++ {
        arg := args[i]

        if arg == "--" {
            filtered = append(filtered, args[i:]...)
            break
        }

        if !strings.HasPrefix(arg, "-") || arg == "-" {
            filtered = append(filtered, arg)
            continue
        }

        if recognizedFlags[arg] {
            filtered = append(filtered, arg)
            if (arg == "-f" || arg == "--file") && i+1 < len(args) {
                filtered = append(filtered, args[i+1])
                i++
            }
            continue
        }

        if strings.HasPrefix(arg, "--file=") {
            filtered = append(filtered, arg)
            continue
        }

        if strings.HasPrefix(arg, "--help=") {
            filtered = append(filtered, "--help")
            continue
        }

        // Unknown flags are intentionally ignored.
    }

    return filtered
}