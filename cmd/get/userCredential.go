package get

import (
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"kufast/tools"
	"os"
	"time"
)

var getUserCredsCmd = &cobra.Command{
	Use:   "user-creds <user>..",
	Short: "Generate user credentials for specific users.",
	Long:  `Generate User credentials for specific user. Can only be used by admins.`,
	Run: func(cmd *cobra.Command, args []string) {
		namespaceName, _ := tools.GetNamespaceFromUserConfig(cmd)

		clientset, client, err := tools.GetUserClient(cmd)
		if err != nil {
			tools.HandleError(err, cmd)
		}

		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
		s.Prefix = "Creating Objects - Please wait!  "
		s.Start()
		var writeOps []<-chan int32
		var results []int32
		for _, user := range args {
			writeOps = append(writeOps, tools.WriteNewUserYamlToFile(user, namespaceName, client, clientset, cmd, s))
		}
		//Ensure all operations are done
		for _, op := range writeOps {
			results = append(results, <-op)
		}
		s.Stop()
	},
}

func init() {
	getCmd.AddCommand(getUserCredsCmd)
	getUserCredsCmd.Flags().StringP("output", "o", ".", "Folder to store the created client credentials. Mandatory, when defining -u")
	getUserCredsCmd.MarkFlagRequired("output")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
