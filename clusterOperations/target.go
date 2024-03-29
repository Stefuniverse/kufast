/*
MIT License

Copyright (c) 2023 Stefan Pawlowski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package clusterOperations

import (
	"context"
	"errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
	"kufast/tools"
	"strings"
)

// IsValidTarget returns true, if the target is valid for this tenant. If all is true, the function returns if this
// is a valid target within the cluster.
func IsValidTarget(cmd *cobra.Command, target string, all bool) bool {
	if strings.Contains(target, "_") {
		return false
	}
	targets, err := ListTargetsFromCmd(cmd, all)
	if err != nil {
		return false
	}
	for _, t := range targets {
		if t.Name == target {
			return true
		}
	}
	return false
}

// IsValidTenantTarget returns true, if the target is valid for this tenant. If all is true, the function returns if this
// is a valid target within the cluster. reads the tenant name from a string.
func IsValidTenantTarget(cmd *cobra.Command, target string, tenantName string, all bool) bool {

	targets, err := ListTargetsFromString(cmd, tenantName, all)
	if err != nil {
		return false
	}
	for _, t := range targets {
		if t.Name == target {
			return true
		}
	}
	return false
}

// GetTargetFromTargetName returns the target to a specific tragetName.
func GetTargetFromTargetName(cmd *cobra.Command, targetName string, tenantName string, all bool) (tools.Target, error) {
	targets, err := ListTargetsFromString(cmd, tenantName, all)
	if err != nil {
		return tools.Target{}, err
	}
	for _, t := range targets {
		if t.Name == targetName {
			return t, nil
		}
	}
	return tools.Target{}, errors.New("the target does not exist or the tenant has no access to the target")
}

// ListTargetsFromString returns a list of targets for a tenant. If all is true, it returns a list of all targets of the
// cluster.
func ListTargetsFromString(cmd *cobra.Command, tenantName string, all bool) ([]tools.Target, error) {

	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return nil, err
	}
	var results []tools.Target

	//Do we want the target of the user or all?
	if all {
		//This information is only available by parsing the nodes
		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		var groups []string
		for _, node := range nodes.Items {
			//Append node target
			results = append(results, tools.Target{
				Name:       node.ObjectMeta.Labels[tools.KUFAST_NODE_HOSTNAME_LABEL],
				AccessType: "node",
			})
			for key, elem := range node.ObjectMeta.Labels {
				if strings.Contains(key, tools.KUFAST_NODE_GROUP_LABEL) && elem != "false" && !slices.Contains(groups, strings.TrimPrefix(key, tools.KUFAST_NODE_GROUP_LABEL)) {
					groups = append(groups, strings.TrimPrefix(key, tools.KUFAST_NODE_GROUP_LABEL))
				}
			}
		}
		for _, target := range groups {
			if target != "" {
				results = append(results, tools.Target{
					Name:       target,
					AccessType: "group",
				})
			}
		}

	} else {

		user, err := clientset.CoreV1().ServiceAccounts("default").Get(context.TODO(), tenantName+"-user", metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		for key, elem := range user.ObjectMeta.Labels {
			if strings.Contains(key, tools.KUFAST_TENANT_GROUPACCESS_LABEL) && elem == "true" {
				results = append(results, tools.Target{
					Name:       strings.TrimPrefix(key, tools.KUFAST_TENANT_GROUPACCESS_LABEL),
					AccessType: "group",
				})
			} else if strings.Contains(key, tools.KUFAST_TENANT_NODEACCESS_LABEL) && elem == "true" {
				results = append(results, tools.Target{
					Name:       strings.TrimPrefix(key, tools.KUFAST_TENANT_NODEACCESS_LABEL),
					AccessType: "node",
				})
			}
		}
	}
	return results, nil

}

// ListTargetsFromCmd returns a list of targets for a tenant. If all is true, it returns a list of all targets of the
// cluster.
func ListTargetsFromCmd(cmd *cobra.Command, all bool) ([]tools.Target, error) {

	//Get the information from the tenant
	namespaceName, _ := tools.GetNamespaceFromUserConfig(cmd)
	tenant, _ := cmd.Flags().GetString("tenant")
	if tenant == "" {
		tenant = tools.GetTenantFromNamespace(namespaceName)
	}

	return ListTargetsFromString(cmd, tenant, all)

}

// SetTargetGroupToNodes Adds all nodes from the array to a target-group. Overwrites previous config.
func SetTargetGroupToNodes(targetName string, targetNodes []string, cmd *cobra.Command) error {
	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return errors.New(err.Error())
	}

	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.New(err.Error())
	}

	if !IsValidTarget(cmd, targetName, true) {
		for _, node := range nodeList.Items {
			if slices.Contains(targetNodes, node.Name) {
				node.ObjectMeta.Labels["kufast.group/"+targetName] = "true"
			} else {
				node.ObjectMeta.Labels["kufast.group/"+targetName] = "false"
			}
			_, err = clientset.CoreV1().Nodes().Update(context.TODO(), &node, metav1.UpdateOptions{})
			if err != nil {
				return errors.New(err.Error())
			}
		}
	}

	return nil
}

// DeleteTargetGroupFromNodes removes a target-group from all nodes.
func DeleteTargetGroupFromNodes(targetName string, cmd *cobra.Command) error {
	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return errors.New(err.Error())
	}

	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.New(err.Error())
	}
	if IsValidTarget(cmd, targetName, true) {
		for _, node := range nodeList.Items {
			delete(node.ObjectMeta.Labels, "kufast.group/"+targetName)
			_, err = clientset.CoreV1().Nodes().Update(context.TODO(), &node, metav1.UpdateOptions{})
			if err != nil {
				return errors.New(err.Error())
			}
		}
	}
	return nil
}
