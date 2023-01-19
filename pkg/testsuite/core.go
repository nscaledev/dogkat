package testsuite

import (
	"errors"
	"fmt"
	"github.com/drew-viles/k8s-e2e-tester/pkg/helpers"
	"github.com/drew-viles/k8s-e2e-tester/pkg/workloads/coreworkloads"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"time"
)

func CheckReadyForTesting(resource []coreworkloads.Resource) error {
	//Thread the tests to run in parallel
	checksCompleted := make(chan coreworkloads.ResourceReady, len(resource))

	readyCheck := func(obj coreworkloads.Resource) {
		r := coreworkloads.ResourceReady{}
		r.Resource = obj
		statusResults := checkIfResourceIsReady(obj, 0, 5)
		if !statusResults {
			log.Printf("%s:%s is not ready\n", obj.GetResourceKind(), obj.GetResourceName())
			r.Ready = false
			checksCompleted <- r
			return
		}
		log.Printf("%s:%s is ready\n", obj.GetResourceKind(), obj.GetResourceName())
		r.Ready = true
		checksCompleted <- r
	}

	defer close(checksCompleted)
	for _, r := range resource {
		if r == nil {
			continue
		}
		go readyCheck(r)
	}

	for _, _ = range resource {
		<-checksCompleted
	}

	return nil
}

// checkIfResourceIsReady validates the readiness of the resource.
func checkIfResourceIsReady(r coreworkloads.Resource, counter int, delaySeconds time.Duration) bool {
	delay := time.Second * delaySeconds
	if counter >= 100 {
		return false
	}
	log.Printf("Waiting for resource to be ready: %s/%s\n", r.GetResourceKind(), r.GetResourceName())
	if !r.IsReady() {
		time.Sleep(delay)
		return checkIfResourceIsReady(r, counter+1, delaySeconds)
	}
	return true
}

//// ScaleUpGPUNodes scales up the GPU nodes that generic workloads will sit on.
//func ScaleUpGPUNodes(resource *coreworkloads.Pod) error {
//	replicaSize := int32(2)
//	//Get number of nodes
//	initialNodeCount := countNodes(resource.Client)
//	//Scale up the workload
//	initialReplicaSize := *resource.Resource.Spec.Replicas
//	log.Println("Testing cluster scaling")
//	log.Printf("Node count before Scale %v\n", initialNodeCount)
//
//	resource.Resource.Spec.Replicas = helpers.IntPtr(replicaSize)
//
//	if err := resource.Update(); err != nil {
//		return errors.New(fmt.Sprintf("Failed to increase replicas for %s:%s: %v\n", resource.Resource.Kind, resource.Resource.Name, err))
//	}
//
//	log.Printf("Waiting for Deployment to scale\n")
//	time.Sleep(time.Second * 60)
//
//	isReady := checkIfResourceIsReady(resource, 0, 5)
//	if !isReady {
//		return errors.New(fmt.Sprintf("There was a problem scaling up the resource - it was not considered ready - you may need to ensure your nodes can support %v of these workloads\n", replicaSize))
//	}
//
//	//Get number of nodes
//	newNodeAmount := countNodes(resource.Client)
//	if newNodeAmount <= initialNodeCount {
//		log.Printf("The node count did not increase - either the nodes were not required, cluster-autoscaler didn't kick in or you're running a single node cluster\n")
//		//Scale down the workload
//		log.Printf("Replicas after Scale %v\n", *resource.Resource.Spec.Replicas)
//		resource.Resource.Spec.Replicas = helpers.IntPtr(initialReplicaSize)
//		if err := resource.Update(); err != nil {
//			log.Printf("Failed to restore replias for %s:%s: %v\n", resource.Resource.Kind, resource.Resource.Name, err)
//		}
//		return nil
//	}
//	log.Printf("Replicas after Scale %v\n", *resource.Resource.Spec.Replicas)
//	log.Printf("Nodes after Scale %v\n", newNodeAmount)
//
//	//Scale down the workload
//	resource.Resource.Spec.Replicas = helpers.IntPtr(initialReplicaSize)
//	if err := resource.Update(); err != nil {
//		log.Printf("Failed to restore replias for %s:%s: %v\n", resource.Resource.Kind, resource.Resource.Name, err)
//	}
//
//	return nil
//}

// ScaleUpStandardNodes scales up the standard nodes that generic workloads will sit on.
func ScaleUpStandardNodes(resource *coreworkloads.Deployment) error {
	replicaSize := int32(5)
	//Get number of nodes
	initialNodeCount := countNodes(resource.Client)
	//Scale up the workload
	initialReplicaSize := *resource.Resource.Spec.Replicas
	log.Println("Testing cluster scaling")
	log.Printf("Node count before Scale %v\n", initialNodeCount)

	resource.Resource.Spec.Replicas = helpers.IntPtr(replicaSize)

	if err := resource.Update(); err != nil {
		return errors.New(fmt.Sprintf("Failed to increase replicas for %s:%s: %v\n", resource.Resource.Kind, resource.Resource.Name, err))
	}

	log.Printf("Waiting for Deployment to scale\n")
	time.Sleep(time.Second * 60)

	isReady := checkIfResourceIsReady(resource, 0, 5)
	if !isReady {
		return errors.New(fmt.Sprintf("There was a problem scaling up the resource - it was not considered ready - you may need to ensure your nodes can support %v of these workloads\n", replicaSize))
	}

	//Get number of nodes
	newNodeAmount := countNodes(resource.Client)
	if newNodeAmount <= initialNodeCount {
		log.Printf("The node count did not increase - either the nodes were not required, cluster-autoscaler didn't kick in or you're running a single node cluster\n")
		//Scale down the workload
		resource.Resource.Spec.Replicas = helpers.IntPtr(initialReplicaSize)
		if err := resource.Update(); err != nil {
			log.Printf("Failed to restore replias for %s:%s: %v\n", resource.Resource.Kind, resource.Resource.Name, err)
		}
		return nil
	}
	log.Printf("Replicas after Scale %v\n", *resource.Resource.Spec.Replicas)
	log.Printf("Nodes after Scale %v\n", newNodeAmount)

	//Scale down the workload
	resource.Resource.Spec.Replicas = helpers.IntPtr(initialReplicaSize)
	if err := resource.Update(); err != nil {
		log.Printf("Failed to restore replias for %s:%s: %v\n", resource.Resource.Kind, resource.Resource.Name, err)
	}

	return nil
}

// countNodes returns the current number of nodes in the cluster.
func countNodes(client *kubernetes.Clientset) int {
	allNodes, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	return len(allNodes.Items)
}
