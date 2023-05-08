package update

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"kufast/clusterOperations"
	"kufast/tools"
)

// listCmd represents the list command
var updateTenantDefaultCmd = &cobra.Command{
	Use:   "tenant-default <newDefault>",
	Short: "Update a resource",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 2 {
			tools.HandleError(errors.New(tools.ERROR_WRONG_NUMBER_ARGUMENTS), cmd)
		}

		s := tools.CreateStandardSpinner(tools.MESSAGE_UPDATE_OBJECTS)

		err := clusterOperations.UpdateTenantDefaultDeployTarget(args[0], cmd)
		if err != nil {
			tools.HandleError(err, cmd)
		}

		s.Stop()
		fmt.Println(tools.MESSAGE_DONE)
	},
}

func init() {
	updateCmd.AddCommand(updateTenantDefaultCmd)

	updateTenantDefaultCmd.Flags().StringP("tenant", "", "", "The name of the tenant to deploy the pod to")
	_ = updateTenantDefaultCmd.MarkFlagRequired("tenant")

}