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
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kufast/objectFactory"
	"kufast/tools"
	"os"
	"time"
)

// CreateDeploymentSecret creates a new deploy-secret. All parameters are drawn from the cobra command.
func CreateDeploymentSecret(secretName string, cmd *cobra.Command) error {

	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return err
	}

	fileName, err := cmd.Flags().GetString("input")
	if err != nil {
		return err
	}

	namespaceName, err := GetTenantTargetNameFromCmd(cmd)
	if err != nil {
		return err
	}

	creds, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	deploymentSecretObject := objectFactory.NewDeploymentSecret(namespaceName, secretName, creds)

	_, err = clientset.CoreV1().Secrets(namespaceName).Create(context.TODO(), deploymentSecretObject, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// CreateSecret creates a new secret. All parameters are drawn from the cobra command.
func CreateSecret(secretName string, secretData string, cmd *cobra.Command) error {
	//Default config block
	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return err
	}

	//Get the namespace
	namespaceName, err := GetTenantTargetNameFromCmd(cmd)
	if err != nil {
		return err
	}

	//create secret object
	secretObject := objectFactory.NewSecret(namespaceName, secretName, secretData)

	//Push secret
	_, err = clientset.CoreV1().Secrets(namespaceName).Create(context.TODO(), secretObject, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// GetSecret gets an existing secret. All parameters are drawn from the cobra command.
func GetSecret(secretName string, cmd *cobra.Command) (*v1.Secret, error) {
	//Initial config block
	namespaceName, err := GetTenantTargetNameFromCmd(cmd)
	if err != nil {
		return nil, err
	}

	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return nil, err
	}

	//execute request
	secret, err := clientset.CoreV1().Secrets(namespaceName).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		tools.HandleError(err, cmd)
	}

	return secret, nil
}

// ListSecrets lists all secrets of a tenant. All parameters are drawn from the cobra command.
func ListSecrets(cmd *cobra.Command) ([]v1.Secret, error) {
	clientset, _, err := tools.GetUserClient(cmd)
	if err != nil {
		return nil, err
	}

	tenantName, err := GetTenantNameFromCmd(cmd)
	if err != nil {
		return nil, err
	}

	targets, err := ListTargetsFromCmd(cmd, false)
	if err != nil {
		return nil, err
	}

	var results []v1.Secret
	for _, target := range targets {
		list, err := clientset.CoreV1().Secrets(tenantName+"-"+target.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		results = append(results, list.Items...)
	}

	return results, nil

}

// DeleteSecret deletes a secret of a tenant. All parameters are drawn from the cobra command.
func DeleteSecret(secretName string, cmd *cobra.Command) <-chan string {
	r := make(chan string)

	go func() {
		defer close(r)

		clientset, _, err := tools.GetUserClient(cmd)
		if err != nil {
			r <- err.Error()
			return
		}

		namespaceName, err := GetTenantTargetNameFromCmd(cmd)

		err = clientset.CoreV1().Secrets(namespaceName).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
		if err != nil {
			r <- err.Error()
			return
		}

		for true {
			_, err := clientset.CoreV1().Secrets(namespaceName).Get(context.TODO(), secretName, metav1.GetOptions{})
			if err != nil {
				r <- ""
				break
			}
			time.Sleep(time.Millisecond * 250)
		}
		r <- ""

	}()
	return r
}
