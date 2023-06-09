package get

import (
	"errors"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"kufast/clusterOperations"
	"kufast/tools"
	"os"
)

// getDeploySecretCmd represents the get deploy-secret command
var getDeploySecretCmd = &cobra.Command{
	Use:   "deploy-secret <secret>",
	Short: "Gain information about a deploy-secret.",
	Long:  `Gain information about a deploy-secret. Output includes name, tenant-target and the secret data.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			tools.HandleError(errors.New(tools.ERROR_WRONG_NUMBER_ARGUMENTS), cmd)
		}

		s := tools.CreateStandardSpinner(tools.MESSAGE_GET_OBJECTS)

		secret, err := clusterOperations.GetSecret(args[0], cmd)
		if err != nil {
			tools.HandleError(err, cmd)
		}

		if secret.Data[".dockerconfigjson"] == nil {
			err := errors.New("Error: This is not a deploy-secret")
			tools.HandleError(err, cmd)
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"ATTRIBUTE", "VALUE"})
		t.AppendRow(table.Row{"Name", secret.Name})
		t.AppendRow(table.Row{"Namespace", secret.Namespace})
		t.AppendRow(table.Row{"Data", string(secret.Data[".dockerconfigjson"])})

		s.Stop()
		t.AppendSeparator()
		t.Render()

	},
}

// init is a helper function from cobra to initialize the command. It sets all flags, standard values and documentation for this command.
func init() {
	getCmd.AddCommand(getDeploySecretCmd)

	getDeploySecretCmd.Flags().StringP("target", "", "", tools.DOCU_FLAG_TARGET)
	getDeploySecretCmd.Flags().StringP("tenant", "", "", tools.DOCU_FLAG_TENANT)
}
