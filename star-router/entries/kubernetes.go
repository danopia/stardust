package entries

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/stardustapp/core/base"
	"github.com/stardustapp/core/extras"
	"github.com/stardustapp/core/inmem"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	appsv1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// Directory containing the clone function
func getKubernetesDriver() *inmem.Folder {
	return inmem.NewFolderOf("kubernetes",
		inmem.NewFunction("invoke", startKubernetes),
		inmem.NewShape(inmem.NewFolderOf("input-shape",
			inmem.NewString("type", "Folder"),
			inmem.NewFolderOf("props",
				inmem.NewString("config-path", "String"),
			),
		)),
	).Freeze()
}

// Function that creates a new Kubernetes client when invoked
func startKubernetes(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	configPath, _ := extras.GetChildString(inputFolder, "config-path")

	// uses the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		log.Println(err)
		return nil
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &kubeApi{
		//config: config,
		svc: clientset,
	}
}

// Presents APIs to inspect and provision against a Kubernetes cluster
type kubeApi struct {
	//config *rest.Config
	svc *kubernetes.Clientset
}

var _ base.Folder = (*kubeApi)(nil)

func (e *kubeApi) Name() string {
	return "kubernetes"
}

func (e *kubeApi) Children() []string {
	return []string{
		"list-pods",
		"run-pod",
		"deploy-svc",
		"submit-job",
	}
}

func (e *kubeApi) Fetch(name string) (entry base.Entry, ok bool) {
	switch name {

	case "list-pods":
		return inmem.NewFolderOf("list-pods",
			&kubeListPodsFunc{e.svc},
			kubeListPodsShape,
			stringOutputShape,
		).Freeze(), true

	case "run-pod":
		return inmem.NewFolderOf("run-pod",
			&kubeRunPodFunc{e.svc},
			kubeRunPodShape,
			stringOutputShape,
		).Freeze(), true

	case "deploy-svc":
		return inmem.NewFolderOf("deploy-svc",
			&kubeDeploySvcFunc{e.svc},
			kubeDeploySvcShape,
			stringOutputShape,
		).Freeze(), true

	case "submit-job":
		return inmem.NewFolderOf("submit-job",
			&kubeSubmitJobFunc{e.svc},
			kubeSubmitJobShape,
			stringOutputShape,
		).Freeze(), true

	default:
		return
	}
}

func (e *kubeApi) Put(name string, entry base.Entry) (ok bool) {
	return false
}

var kubeListPodsShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props"),
	))

type kubeListPodsFunc struct {
	svc *kubernetes.Clientset
}

var _ base.Function = (*kubeListPodsFunc)(nil)

func (e *kubeListPodsFunc) Name() string {
	return "invoke"
}

func (e *kubeListPodsFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	pods, err := e.svc.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.Println(err)
		return nil
	}
	//log.Println("There are", len(pods.Items), "pods in the cluster")

	//list := inmem.NewFolder("pod-list")
	var list string
	for _, pod := range pods.Items {
		//list.Put(pod.ObjectMeta.Name, inmem.NewFolderOf(pod.ObjectMeta.Name,
		//  ""
		list += pod.ObjectMeta.Name + "\n"
	}

	return inmem.NewString("pod-list", list)
}

var kubeRunPodShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewString("name", "String"),
			inmem.NewString("image", "String"),
			inmem.NewString("command", "String"),
			inmem.NewString("privileged", "String"),
		),
	))

type kubeRunPodFunc struct {
	svc *kubernetes.Clientset
}

var _ base.Function = (*kubeRunPodFunc)(nil)

func (e *kubeRunPodFunc) Name() string {
	return "invoke"
}

func (e *kubeRunPodFunc) getLogs(podName string, out *bytes.Buffer) error {
	req := e.svc.CoreV1().RESTClient().Get().
		Namespace("default").
		Name(podName).
		Resource("pods").
		SubResource("log") //.
		//Param("follow", strconv.FormatBool(logOptions.Follow)).
		//Param("container", "pod").
		//Param("previous", strconv.FormatBool(logOptions.Previous)).
		//Param("timestamps", strconv.FormatBool(logOptions.Timestamps)).
		//Param("sinceSeconds", strconv.FormatInt(*logOptions.SinceSeconds, 10)).
		//Param("sinceTime", logOptions.SinceTime.Format(time.RFC3339)).
		//Param("limitBytes", strconv.FormatInt(*logOptions.LimitBytes, 10)).
		//Param("tailLines", strconv.FormatInt(*logOptions.TailLines, 10))

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}

	defer readCloser.Close()
	_, err = io.Copy(out, readCloser)
	return err
}

// jobs will create unlimited pods if the failure keeps happening
// do we even want jobs then?
func (e *kubeRunPodFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	podName, _ := extras.GetChildString(inputFolder, "name")
	podImage, _ := extras.GetChildString(inputFolder, "image")
	podCommand, _ := extras.GetChildString(inputFolder, "command")
	podCmdParts := strings.Split(podCommand, " ")

	podPrivileged, _ := extras.GetChildString(inputFolder, "privileged")
	isPrivileged := podPrivileged == "yes"

	pods := e.svc.CoreV1().Pods("default")
	_, err := pods.Create(&apiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "core/v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: "Never",
			Volumes: []apiv1.Volume{
				{
					Name: "docker-socket",
					VolumeSource: apiv1.VolumeSource{
						HostPath: &apiv1.HostPathVolumeSource{
							Path: "/var/run/docker.sock",
						},
					},
				},
				{
					Name: "go-src-cache",
					VolumeSource: apiv1.VolumeSource{
						HostPath: &apiv1.HostPathVolumeSource{
							Path: "/data/go-src-cache",
						},
					},
				},
			},
			Containers: []apiv1.Container{
				{
					Name:            "pod",
					Image:           podImage,
					ImagePullPolicy: "Never",
					Command:         []string{podCmdParts[0]},
					Args:            podCmdParts[1:],
					SecurityContext: &apiv1.SecurityContext{
						Privileged: &isPrivileged,
					},
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "docker-socket",
							MountPath: "/var/run/docker.sock",
						},
						{
							Name:      "go-src-cache",
							MountPath: "/go/src",
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Pod submission failed:", err)
		return inmem.NewString("error", err.Error())
	}

	watcher, err := pods.Watch(metav1.ListOptions{
		FieldSelector: "metadata.name=" + podName,
	})
	if err != nil {
		log.Println("Pod watching failed:", err)
		return inmem.NewString("error", err.Error())
	}

	var logs bytes.Buffer
	var terminated bool

	for evt := range watcher.ResultChan() {
		pod := evt.Object.(*apiv1.Pod)

		containerStatuses := pod.Status.ContainerStatuses
		if len(containerStatuses) > 0 {
			containerState := containerStatuses[0].State
			if containerState.Terminated != nil && !terminated {
				terminated = true

				logs.WriteString(fmt.Sprintf("Pod terminated with code %v\n\n", containerState.Terminated.ExitCode))
				e.getLogs(podName, &logs)
				pods.Delete(podName, nil) // TODO: err
			}
		}

		switch evt.Type {
		case "ADDED":
			log.Println("Pod was added", pod.Status)

		case "MODIFIED":
			log.Println("Pod is now", pod.Status.Phase)

		case "DELETED":
			log.Println("Pod was deleted!")
			watcher.Stop()
		}
	}

	return inmem.NewString("output", logs.String())
}

var kubeDeploySvcShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewString("name", "String"),
			inmem.NewString("image", "String"),
		),
	))

type kubeDeploySvcFunc struct {
	svc *kubernetes.Clientset
}

var _ base.Function = (*kubeDeploySvcFunc)(nil)

func (e *kubeDeploySvcFunc) Name() string {
	return "invoke"
}

/*
func (e *kubeDeploySvcFunc) getLogs(podName string, out *bytes.Buffer) error {
	req := e.svc.CoreV1().RESTClient().Get().
		Namespace("default").
		Name(podName).
		Resource("pods").
		SubResource("log") //.
		//Param("follow", strconv.FormatBool(logOptions.Follow)).
		//Param("container", "pod").
		//Param("previous", strconv.FormatBool(logOptions.Previous)).
		//Param("timestamps", strconv.FormatBool(logOptions.Timestamps)).
		//Param("sinceSeconds", strconv.FormatInt(*logOptions.SinceSeconds, 10)).
		//Param("sinceTime", logOptions.SinceTime.Format(time.RFC3339)).
		//Param("limitBytes", strconv.FormatInt(*logOptions.LimitBytes, 10)).
		//Param("tailLines", strconv.FormatInt(*logOptions.TailLines, 10))

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}

	defer readCloser.Close()
	_, err = io.Copy(out, readCloser)
	return err
}
*/

// jobs will create unlimited pods if the failure keeps happening
// do we even want jobs then?
func (e *kubeDeploySvcFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	svcName, _ := extras.GetChildString(inputFolder, "name")
	podImage, _ := extras.GetChildString(inputFolder, "image")

	var replicas int32 = 1
	var terminationSeconds int64 = 1
	//MaxUnavailable := intstr.FromInt(0)
	//MaxSurge := intstr.FromInt(1)
	desiredDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "stardriver-" + svcName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"stardriver": svcName,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
				// Type: appsv1.RollingUpdateDeploymentStrategyType,
				// RollingUpdate: &appsv1.RollingUpdateDeployment{
				// 	MaxUnavailable: &MaxUnavailable,
				// 	MaxSurge:       &MaxSurge,
				// },
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Labels: map[string]string{
						"stardriver": svcName,
					},
					Annotations: map[string]string{
						"stardust-nonce": extras.GenerateSecret(),
					},
				},
				Spec: apiv1.PodSpec{
					RestartPolicy: "Always",
					TerminationGracePeriodSeconds: &terminationSeconds,
					Volumes: []apiv1.Volume{
						{
							Name: "ca-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/ssl/certs/ca-certificates.crt",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:            "driver",
							Image:           podImage,
							ImagePullPolicy: "Never",
							//Command:         []string{podCmdParts[0]},
							//Args:            podCmdParts[1:],
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 9234,
									Protocol:      "TCP",
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "ca-certs",
									MountPath: "/etc/ssl/certs/ca-certificates.crt",
								},
							},
						},
					},
				},
			},
		},
	}

	desiredService := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "api/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "stardriver-" + svcName,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(9234),
				},
			},
			Selector: map[string]string{
				"stardriver": svcName,
			},
			Type: apiv1.ServiceTypeNodePort,
		},
	}

	deployments := e.svc.AppsV1beta1().Deployments("default")
	_, err := deployments.Update(desiredDeployment)
	if err != nil && strings.HasSuffix(err.Error(), "not found") {
		_, err = deployments.Create(desiredDeployment)
	}

	if err != nil {
		log.Println("Deployment submission failed:", err)
		return inmem.NewString("error", err.Error())
	}

	services := e.svc.CoreV1().Services("default")
	_, err = services.Create(desiredService)
	if err != nil && strings.HasSuffix(err.Error(), "already exists") {
		log.Println("k8s:", err.Error())
		err = nil
	}
	if err != nil {
		log.Println("Service submission failed:", err)
		return inmem.NewString("error", err.Error())
	}

	// Wait for deployment to stabilize
	var ready bool
	for !ready { // condition.Status != apiv1.ConditionTrue || condition.Reason != "NewReplicaSetAvailable" {
		time.Sleep(1000 * time.Millisecond)
		depl, err := deployments.Get(desiredDeployment.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			log.Println("Deployment polling failed:", err)
			return inmem.NewString("error", err.Error())
		}

		ready = true
		for _, cond := range depl.Status.Conditions {
			if cond.Type == "Progressing" {
				if cond.Status != apiv1.ConditionTrue || cond.Reason != "NewReplicaSetAvailable" {
					ready = false
					log.Printf("deployment %s progress: %v", desiredDeployment.ObjectMeta.Name, cond)
				}
			}
			if cond.Type == "Available" {
				if cond.Status != apiv1.ConditionTrue {
					ready = false
					log.Printf("deployment %s availability: %v", desiredDeployment.ObjectMeta.Name, cond)
				}
			}
		}
	}
	log.Printf("deployment %s is updated and available", desiredDeployment.ObjectMeta.Name)

	svc, err := services.Get(desiredService.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		log.Println("Service fetching failed:", err)
		return inmem.NewString("error", err.Error())
	}

	log.Printf("Driver is exposed on port %d", svc.Spec.Ports[0].NodePort)
	return inmem.NewString("endpoint-url", fmt.Sprintf("http://192.168.86.3:%d/~~export", svc.Spec.Ports[0].NodePort))
}

var kubeSubmitJobShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewString("name", "String"),
			inmem.NewString("image", "String"),
			inmem.NewString("command", "String"),
		),
	))

type kubeSubmitJobFunc struct {
	svc *kubernetes.Clientset
}

var _ base.Function = (*kubeSubmitJobFunc)(nil)

func (e *kubeSubmitJobFunc) Name() string {
	return "invoke"
}

func (e *kubeSubmitJobFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	jobName, _ := extras.GetChildString(inputFolder, "name")
	jobImage, _ := extras.GetChildString(inputFolder, "image")
	jobCommand, _ := extras.GetChildString(inputFolder, "command")
	jobCmdParts := strings.Split(jobCommand, " ")

	jobs := e.svc.BatchV1().Jobs("default")

	var parallelism int32 = 1
	job, err := jobs.Create(&batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &parallelism,
			Completions: &parallelism,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: jobName,
				},
				Spec: apiv1.PodSpec{
					RestartPolicy: "OnFailure",
					Containers: []apiv1.Container{
						{
							Name:    "job",
							Image:   jobImage,
							Command: []string{jobCmdParts[0]},
							Args:    jobCmdParts[1:],
						},
					},
				},
			},
		},
	})

	log.Printf("Job submission result: %+v %v", job, err)
	return nil
}
