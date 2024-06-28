package main

import (
	"log"

	"github.com/spf13/cobra"
	banking "github.com/tomwheeler/demo-bank/app/bank"
)

var (
	name string
	port int
)

var rootCmd = &cobra.Command{
	Use:   "Bank Service for recipient",
	Short: "Starts the service for the recipient's bank",
	RunE: func(*cobra.Command, []string) error {
		bank := banking.NewBank(name)
		data := bank.GetDataPath()

		log.Println("Starting the recipient's banking service")
		log.Printf("   Name: %s\n", name)
		log.Printf("   Data: %s\n", data)
		log.Printf("   Port: %d\n", port)

		service := banking.NewBankingService(bank, port)
		return service.Start()
	},
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&name,
		"name", "n", "Ted", "Name of recipient")
	rootCmd.PersistentFlags().IntVarP(&port,
		"port", "p", 8889, "Port for recipient's banking service")

	cobra.CheckErr(rootCmd.Execute())
}
