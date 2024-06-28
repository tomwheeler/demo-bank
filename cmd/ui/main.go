package main

import (
	"log"

	"github.com/spf13/cobra"
	banking "github.com/tomwheeler/demo-bank/app/bank"
	"github.com/tomwheeler/demo-bank/app/ui"
)

var (
	// sender bank service info
	sHost string
	sPort int
	// recipient bank service info
	rHost string
	rPort int
)

var rootCmd = &cobra.Command{
	Use:   "start-ui",
	Short: "Start UI for Demo",
	RunE: func(*cobra.Command, []string) error {
		log.Println("Starting Bank UI")
		log.Printf("   Sender Bank:    http://%s:%d/\n", sHost, sPort)
		log.Printf("   Recipient Bank: http://%s:%d/\n", rHost, rPort)

		senderClient := banking.NewBankClient("localhost", 8888)
		recipientClient := banking.NewBankClient("localhost", 8889)

		ui.BuildUI(senderClient, recipientClient)

		return nil
	},
}

func main() {
	rootCmd.PersistentFlags().StringVar(&sHost,
		"sender-host", "localhost", "Service host for sender's bank")
	rootCmd.PersistentFlags().IntVar(&sPort,
		"sender-port", 8888, "Service port for sender's bank")
	rootCmd.PersistentFlags().StringVar(&rHost,
		"recipient-host", "localhost", "Service host for recipient bank")
	rootCmd.PersistentFlags().IntVar(&rPort,
		"recipient-port", 8889, "Service port for recipient's bank")

	cobra.CheckErr(rootCmd.Execute())
}
