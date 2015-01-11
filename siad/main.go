package main

import (
	"fmt"
	"os"

	"code.google.com/p/gcfg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	config Config
)

type Config struct {
	Siacore struct {
		RPCaddr       string
		HostDirectory string
		NoBootstrap   bool
	}

	Siad struct {
		APIaddr           string
		ConfigFilename    string
		DownloadDirectory string
		StyleDirectory    string
		WalletFile        string
	}
}

// findSiaDir first checks the current directory, then checks ~/.config/sia,
// looking for the html index file that's needed to run the web app.
// findSiaDir will return home="" if it can't find the home dir, but it won't
// report an error for that. It'll only report an error if it can't find
// index.html.
func findSiaDir() (home, siaDir string, err error) {
	// Check the current directory for the index file.
	var found bool
	if _, err = os.Stat("style/index.html"); err == nil {
		found = true
	}

	// Check ~/.config/sia for the index file.
	home, err = homedir.Dir()
	if err == nil && !found {
		dirname := home + "/.config/sia/style/index.html"
		if _, err = os.Stat(dirname); err == nil {
			siaDir = home + "/.config/sia/"
			return
		}
	}

	// This is the only error that can be returned.
	if !found {
		err = fmt.Errorf("Style folder not found, please put the 'style/' folder in the current directory")
	} else {
		err = nil
	}
	return
}

// startEnvironment calls createEnvironment(), which will handle everything
// else.
func startEnvironment(cmd *cobra.Command, args []string) {
	_, err := createDaemon(config)
	if err != nil {
		fmt.Println(err)
	}
}

// Prints version information about Sia Daemon.
func version(cmd *cobra.Command, args []string) {
	fmt.Println("Sia Daemon v0.1.0")
}

func main() {
	root := &cobra.Command{
		Use:   os.Args[0],
		Short: "Sia Daemon v0.1.0",
		Long:  "Sia Daemon v0.1.0",
		Run:   startEnvironment,
	}

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information about the Sia Daemon",
		Run:   version,
	})

	// Add flag defaults, which have the lowest priority.
	home, siaDir, err := findSiaDir()
	if err != nil {
		fmt.Println("Warning:", err)
	}
	defaultConfigFile := siaDir + "config"
	defaultHostDir := siaDir + "host/"
	defaultStyleDir := siaDir + "style/"
	defaultDownloadDir := home + "/Downloads/"
	defaultWalletFile := siaDir + "sia.wallet"
	root.PersistentFlags().StringVarP(&config.Siad.APIaddr, "api-addr", "a", "localhost:9980", "which port is used to communicate with the user")
	root.PersistentFlags().StringVarP(&config.Siacore.RPCaddr, "rpc-addr", "r", ":9988", "which port is used when talking to other nodes on the network")
	root.PersistentFlags().BoolVarP(&config.Siacore.NoBootstrap, "no-bootstrap", "n", false, "disable bootstrapping on this run.")
	root.PersistentFlags().StringVarP(&config.Siad.ConfigFilename, "config-file", "c", defaultConfigFile, "tell siad where to load the config file")
	root.PersistentFlags().StringVarP(&config.Siacore.HostDirectory, "host-dir", "H", defaultHostDir, "where the host puts all uploaded files")
	root.PersistentFlags().StringVarP(&config.Siad.StyleDirectory, "style-dir", "s", defaultStyleDir, "where to find the files that compose the frontend")
	root.PersistentFlags().StringVarP(&config.Siad.DownloadDirectory, "download-dir", "d", defaultDownloadDir, "where to download files")
	root.PersistentFlags().StringVarP(&config.Siad.WalletFile, "wallet-file", "w", defaultWalletFile, "where to keep the wallet")

	// Load the config file, which has the middle priorty. Only values defined
	// in the config file will be set.
	if _, err = os.Stat(config.Siad.ConfigFilename); err == nil {
		err := gcfg.ReadFileInto(&config, config.Siad.ConfigFilename)
		if err != nil {
			fmt.Println("Error reading config file:", err)
		}
	}

	// Execute wil over-write any flags set by the config file, but only if the
	// user specified them manually.
	root.Execute()
}
