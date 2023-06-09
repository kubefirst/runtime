/*
Copyright (C) 2021-2023, Kubefirst

This program is licensed under MIT.
See the LICENSE file for more details.
*/
package k8s

import (
	"github.com/rs/zerolog/log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// OpenPortForwardPodWrapper wrapper for PortForwardPod function. This functions make it easier to open and close port
// forward request. By providing the function parameters, the function will manage to create the port forward. The
// parameter for the stopChannel controls when the port forward must be closed.
//
// Example:
//
//	vaultStopChannel := make(chan struct{}, 1)
//	go func() {
//		OpenPortForwardWrapper(
//			pkg.VaultPodName,
//			pkg.VaultNamespace,
//			pkg.VaultPodPort,
//			pkg.VaultPodLocalPort,
//			vaultStopChannel)
//		wg.Done()
//	}()
func OpenPortForwardPodWrapper(
	clientset *kubernetes.Clientset,
	restConfig *rest.Config,
	podName string,
	namespace string,
	podPort int,
	podLocalPort int,
	stopChannel chan struct{},
) {
	// readyCh communicate when the port forward is ready to get traffic
	readyCh := make(chan struct{})

	portForwardRequest := PortForwardAPodRequest{
		RestConfig: restConfig,
		Pod: v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: namespace,
			},
		},
		PodPort:   podPort,
		LocalPort: podLocalPort,
		StopCh:    stopChannel,
		ReadyCh:   readyCh,
	}

	// Check to see if the port is already used
	err := CheckForExistingPortForwards(podLocalPort)
	if err != nil {
		log.Fatal().Msgf("unable to start port forward for pod %s in namespace %s: %s", podName, namespace, err)
	}

	go func() {
		err := PortForwardPodWithRetry(clientset, portForwardRequest)
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
		}
	}()

	select {
	case <-stopChannel:
		log.Info().Msg("leaving...")
		close(stopChannel)
		close(readyCh)
		break
	case <-readyCh:
		log.Info().Msg("port forwarding is ready to get traffic")
	}

	log.Info().Msgf("pod %q at namespace %q has port-forward accepting local connections at port %d\n", podName, namespace, podLocalPort)

}
