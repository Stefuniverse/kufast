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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kufast/objectFactory"
	"kufast/tools"
	"time"
)

// CreateTenant creates a new tenant. All parameters are drawn from the environment on the command line.
func CreateTenant(tenantName string, cmd *cobra.Command) error {

	//Configblock
	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().ServiceAccounts("default").Create(context.TODO(), objectFactory.NewTenantUser(tenantName, "default"), metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_, err = clientset.RbacV1().Roles("default").Create(context.TODO(), objectFactory.NewTenantDefaultRole(tenantName), metav1.CreateOptions{})
	if err != nil {
		return err
	}

	_, err = clientset.RbacV1().RoleBindings("default").Create(context.TODO(), objectFactory.NewTenantDefaultRoleBinding(tenantName), metav1.CreateOptions{})
	if err != nil {
		return err
	}

	timeout := 600
	for true {
		timeout--

		if timeout == 0 {
			return errors.New(`Operation Timeout. Your tenant has been initialized but it is not ready yet. 
Please ensure it is fully initialized and get its credentials from 'kufast get tenant-creds'`)
		}
		tenant, err := clientset.CoreV1().ServiceAccounts("default").Get(context.TODO(), tenantName+"-user", metav1.GetOptions{})
		time.Sleep(time.Millisecond * 1000)
		if err == nil && tenant.Secrets != nil && len(tenant.Secrets) > 0 {
			break
		} else if err != nil {
			return err
		}
	}
	return nil
}

// DeleteTenant Deletes a tenant. All parameters are drawn from the environment on the command line.
func DeleteTenant(tenantName string, cmd *cobra.Command) error {
	//Configblock
	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return err
	}

	err = clientset.CoreV1().ServiceAccounts("default").Delete(context.TODO(), tenantName+"-user", metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = clientset.RbacV1().Roles("default").Delete(context.TODO(), tenantName+"-defaultrole", metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = clientset.RbacV1().RoleBindings("default").Delete(context.TODO(), tenantName+"-defaultrolebinding", metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// GetTenantNameFromCmd gets the name of a tenant from cmd. All parameters are drawn from the environment on the command line.
func GetTenantNameFromCmd(cmd *cobra.Command) (string, error) {
	tenant, _ := cmd.Flags().GetString("tenant")
	if tenant == "" {
		namespaceName, err := tools.GetNamespaceFromUserConfig(cmd)
		if err != nil {
			return "", err
		}
		return tools.GetTenantFromNamespace(namespaceName), nil
	}
	return tenant, nil
}

// GetTenantFromCmd gets a tenant object. All parameters are drawn from the environment on the command line.
func GetTenantFromCmd(cmd *cobra.Command) (*v1.ServiceAccount, error) {

	tenantName, err := GetTenantNameFromCmd(cmd)
	if err != nil {
		return nil, err
	}

	//Configblock
	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return nil, err
	}

	user, err := clientset.CoreV1().ServiceAccounts("default").Get(context.TODO(), tenantName+"-user", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetTenantFromString gets a tenant object from its name. All parameters are drawn from the environment on the command line.
func GetTenantFromString(cmd *cobra.Command, tenantName string) (*v1.ServiceAccount, error) {

	//Configblock
	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return nil, err
	}

	user, err := clientset.CoreV1().ServiceAccounts("default").Get(context.TODO(), tenantName+"-user", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateTenantDefaultDeployTarget sets the kufast/default label of a tenant to a new value.
func UpdateTenantDefaultDeployTarget(newDefaultTarget string, cmd *cobra.Command) error {
	//Configblock
	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return err
	}

	tenant, err := GetTenantFromCmd(cmd)
	if err != nil {
		return err
	}

	tenant.ObjectMeta.Labels[tools.KUFAST_TENANT_DEFAULT_LABEL] = newDefaultTarget
	_, err = clientset.CoreV1().ServiceAccounts("default").Update(context.TODO(), tenant, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil

}

// DeleteTargetFromTenant deletes a target from a tenant.
func DeleteTargetFromTenant(targetName string, tenantName string, cmd *cobra.Command) error {
	if IsValidTenantTarget(cmd, targetName, tenantName, false) {
		clientset, _, err := tools.GetUserClient(cmd)
		if err != nil {
			return errors.New(err.Error())
		}

		target, err := GetTargetFromTargetName(cmd, targetName, tenantName, false)
		if err != nil {
			return errors.New(err.Error())
		}

		tenant, err := GetTenantFromString(cmd, tenantName)
		if err != nil {
			return errors.New(err.Error())
		}

		if target.AccessType == "node" {
			delete(tenant.ObjectMeta.Labels, tools.KUFAST_TENANT_NODEACCESS_LABEL+targetName)
		} else {
			delete(tenant.ObjectMeta.Labels, tools.KUFAST_TENANT_GROUPACCESS_LABEL+targetName)
		}
		_, err = clientset.CoreV1().ServiceAccounts("default").Update(context.TODO(), tenant, metav1.UpdateOptions{})
		if err != nil {
			return errors.New(err.Error())
		}

	} else {
		return errors.New("Not a valid target for this tenant: " + targetName)
	}

	return nil
}

// AddTargetToTenant adds a new target to a tenant.
func AddTargetToTenant(cmd *cobra.Command, targetName string, tenantName string) error {
	if IsValidTenantTarget(cmd, targetName, tenantName, true) {
		clientset, _, err := tools.GetUserClient(cmd)
		if err != nil {
			return errors.New(err.Error())
		}

		target, err := GetTargetFromTargetName(cmd, targetName, tenantName, true)
		if err != nil {
			return err
		}
		tenant, err := GetTenantFromString(cmd, tenantName)
		if err != nil {
			return err
		}
		if target.AccessType == "node" {
			tenant.ObjectMeta.Labels[tools.KUFAST_TENANT_NODEACCESS_LABEL+targetName] = "true"
		} else {
			tenant.ObjectMeta.Labels[tools.KUFAST_TENANT_GROUPACCESS_LABEL+targetName] = "true"
		}

		// Populate default label if possible
		if tenant.ObjectMeta.Labels[tools.KUFAST_TENANT_DEFAULT_LABEL] == "" {
			tenant.ObjectMeta.Labels[tools.KUFAST_TENANT_DEFAULT_LABEL] = targetName
		}
		_, err = clientset.CoreV1().ServiceAccounts("default").Update(context.TODO(), tenant, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("Invalid target!")
}

// GetTenantDefaultTargetNameFromCmd returns the default target name of a tenant from cmd parameters.
func GetTenantDefaultTargetNameFromCmd(cmd *cobra.Command) (string, error) {

	user, err := GetTenantFromCmd(cmd)
	if err != nil {
		return "", err
	}

	return user.ObjectMeta.Labels[tools.KUFAST_TENANT_DEFAULT_LABEL], nil
}
