package e2e_test

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"

	// . "github.com/onsi/gomega"

	// . "github.com/ovn-org/ovn-kubernetes/test/e2e"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/test/e2e/framework"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
)

func checkContinuousConnectivity(f *framework.Framework, nodeName, podName, host string, port, timeout int, readyChan chan int, errChan chan error) {
	contName := fmt.Sprintf("%s-container", podName)

	command := []string{
		"bash", "-c",
		"set -xe; for i in {1..10}; do nc -vz -w " + strconv.Itoa(timeout) + " " + host + " " + strconv.Itoa(port) + "; sleep 2; done",
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    contName,
					Image:   framework.AgnHostImage,
					Command: command,
				},
			},
			NodeName:      nodeName,
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
	podClient := f.ClientSet.CoreV1().Pods(f.Namespace.Name)
	_, err := podClient.Create(pod)
	if err != nil {
		errChan <- err
		return
	}

	readyChan <- 0

	err = e2epod.WaitForPodSuccessInNamespace(f.ClientSet, podName, f.Namespace.Name)

	if err != nil {
		logs, logErr := e2epod.GetPodLogs(f.ClientSet, f.Namespace.Name, pod.Name, contName)
		if logErr != nil {
			framework.Logf("Warning: Failed to get logs from pod %q: %v", pod.Name, logErr)
		} else {
			framework.Logf("pod %s/%s logs:\n%s", f.Namespace.Name, pod.Name, logs)
		}
	}

	errChan <- err
}

// checkConnectivityToHost launches a pod to test connectivity to the specified
// host. An error will be returned if the host is not reachable from the pod.
//
// An empty nodeName will use the schedule to choose where the pod is executed.
func checkConnectivityToHost(f *framework.Framework, nodeName, podName, host string, port, timeout int) error {
	contName := fmt.Sprintf("%s-container", podName)

	command := []string{
		"nc",
		"-vz",
		"-w", strconv.Itoa(timeout),
		host,
		strconv.Itoa(port),
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    contName,
					Image:   framework.AgnHostImage,
					Command: command,
				},
			},
			NodeName:      nodeName,
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
	podClient := f.ClientSet.CoreV1().Pods(f.Namespace.Name)
	_, err := podClient.Create(pod)
	if err != nil {
		return err
	}
	err = e2epod.WaitForPodSuccessInNamespace(f.ClientSet, podName, f.Namespace.Name)

	if err != nil {
		logs, logErr := e2epod.GetPodLogs(f.ClientSet, f.Namespace.Name, pod.Name, contName)
		if logErr != nil {
			framework.Logf("Warning: Failed to get logs from pod %q: %v", pod.Name, logErr)
		} else {
			framework.Logf("pod %s/%s logs:\n%s", f.Namespace.Name, pod.Name, logs)
		}
	}

	return err
}

var _ = Describe("E2e", func() {
	var svcname = "nettest"

	f := framework.NewDefaultFramework(svcname)

	ginkgo.BeforeEach(func() {
		// Assert basic external connectivity.
		// Since this is not really a test of kubernetes in any way, we
		// leave it as a pre-test assertion, rather than a Ginko test.
		ginkgo.By("Executing a successful http request from the external internet")
		resp, err := http.Get("http://google.com")
		if err != nil {
			framework.Failf("Unable to connect/talk to the internet: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			framework.Failf("Unexpected error code, expected 200, got, %v (%v)", resp.StatusCode, resp)
		}
	})

	ginkgo.It("should provide Internet connection for containers [Feature:Networking-IPv4]", func() {
		ginkgo.By("Running container which tries to connect to 8.8.8.8")
		framework.ExpectNoError(
			checkConnectivityToHost(f, "", "connectivity-test", "8.8.8.8", 53, 30))
	})

	ginkgo.It("should provide Internet connection continuously when ovn-k8s pod is killed", func() {
		ginkgo.By("Running container which tries to connect to 8.8.8.8 in a loop")

		readyChan, errChan := make(chan int), make(chan error)
		go checkContinuousConnectivity(f, "", "connectivity-test-continuous", "8.8.8.8", 53, 30, readyChan, errChan)

		<-readyChan
		framework.Logf("Container is ready, waiting a few seconds")

		time.Sleep(10 * time.Second)
		podClient := f.ClientSet.CoreV1().Pods("ovn-kubernetes")
		podClient2 := f.ClientSet.CoreV1().Pods(f.Namespace.Name)

		podList, _ := podClient.List(metav1.ListOptions{})
		podList2, _ := podClient2.List(metav1.ListOptions{})
		framework.Logf("ovn-kubernetes %q", podList.String())
		framework.Logf("framework namespace %q", podList2.String())

		err := podClient2.Delete("ovnkube-node", metav1.NewDeleteOptions(0))

		framework.ExpectNoError(err, "should delete ovnkube-node pod")

		framework.ExpectNoError(<-errChan)
	})
})
