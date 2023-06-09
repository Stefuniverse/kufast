package list

import (
	"github.com/spf13/cobra/doc"
	"kufast/cmd"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// listCmd represents the list command. It cannot be executed itself but only its subcommands.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List kufast objects",
	Long: `The list subcommand is a collection of all list operations available in kufast.
Use these features to list tenants, pods and more.`,
}

// init is a helper function from cobra to initialize the command. It sets all flags, standard values and documentation for this command.
func init() {
	cmd.RootCmd.AddCommand(listCmd)

}

func CreateListDocs(fileP func(string) string, linkH func(string) string) {

	err := os.MkdirAll("./kufast.wiki/list/", 0770)
	if err != nil {
		panic(err)
	}

	err = doc.GenMarkdownTreeCustom(listCmd, "./kufast.wiki/list/", fileP, linkH)
	if err != nil {
		log.Fatal(err)
	}
}
