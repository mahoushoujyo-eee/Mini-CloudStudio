package util

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"learn/biz/config"
	"learn/biz/model"
)

type KubernetesUtil struct {
	ctx context.Context
}

func NewKubernetesUtil(ctx context.Context) *KubernetesUtil {
	return &KubernetesUtil{ctx: ctx}
}

func (s *KubernetesUtil) EnsureNamespace(namespace string) error {
	// 先检查命名空间是否存在
	_, err := config.KubernetesClient.CoreV1().Namespaces().Get(s.ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// 命名空间不存在，创建它
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
					Labels: map[string]string{
						"created-by": "hertz",
						"purpose":    "code-server",
					},
				},
			}

			_, createErr := config.KubernetesClient.CoreV1().Namespaces().Create(s.ctx, ns, metav1.CreateOptions{})
			if createErr != nil && !errors.IsAlreadyExists(createErr) {
				return fmt.Errorf("创建命名空间失败: %w", createErr)
			}

			log.Printf("命名空间 %s 创建成功\n", namespace)
			return nil
		}
		return fmt.Errorf("检查命名空间失败: %w", err)
	}

	log.Printf("命名空间 %s 已存在\n", namespace)
	return nil
}

func (s *KubernetesUtil) GetPodList(param *model.KubernetesParam) (*corev1.PodList, error) {
	// 获取 namespace 下的所有 Pod
	pods, err := config.KubernetesClient.CoreV1().Pods(param.Namespace).List(s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d pods in '%s' namespace\n", len(pods.Items), param.Namespace)

	return pods, nil
}

func (s *KubernetesUtil) CreateDeployment(kbParam *model.KubernetesParam, appParam *model.AppParam) error {
	replicas := int32(1) // 默认1个副本，您可以根据需要调整

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kbParam.Deployment, // 这里可能需要改为deploymentName之类的参数
			Namespace: kbParam.Namespace,
			Labels: map[string]string{
				"app":        "code-server",
				"deployment": kbParam.Deployment,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":        "code-server",
					"deployment": kbParam.Deployment,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "code-server",
						"deployment": kbParam.Deployment,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "code-server",
							Image:           "docker.1ms.run/linuxserver/code-server:4.103.0",
							ImagePullPolicy: corev1.PullNever,
							Env: []corev1.EnvVar{
								{Name: "PUID", Value: "1000"},
								{Name: "PGID", Value: "1000"},
								{Name: "TZ", Value: "Etc/UTC"},
								{Name: "PASSWORD", Value: appParam.PodPassword},
								{Name: "SUDO_PASSWORD", Value: appParam.PodPassword},
								{Name: "PWA_APPNAME", Value: "code-server"},
								{Name: "HTTP_PROXY", Value: "http://223.2.19.172:3128"},
								{Name: "http_proxy", Value: "http://223.2.19.172:3128"},
								{Name: "HTTPS_PROXY", Value: "http://223.2.19.172:3128"},
								{Name: "https_proxy", Value: "http://223.2.19.172:3128"},
							},
							Ports: []corev1.ContainerPort{{
								ContainerPort: 8443,
								Protocol:      corev1.ProtocolTCP,
								Name:          "https",
							}},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "data",
								MountPath: "/config",
							}},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(appParam.Cpu),
									corev1.ResourceMemory: resource.MustParse(appParam.Memory),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(appParam.Cpu),
									corev1.ResourceMemory: resource.MustParse(appParam.Memory),
								},
							},
						},
						{
							Name:            "heartbeater",
							Image:           "223.2.19.172:30002/minics/heartbeater:1.0.0",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{Name: "USER_ID", Value: strconv.Itoa(int(appParam.UserId))},
								{Name: "KUBERNETES_POD_START_TIME", Value: time.Now().UTC().Format("2006-01-02T15:04:05Z")},
								{Name: "KAFKA_TOPIC", Value: "timecounter"},
								{Name: "KAFKA_BROKERS", Value: "kafka.minics.svc.cluster.local:9092"},
								{Name: "DEPLOY_NAME", Value: kbParam.Deployment},
							},
						}},
					Volumes: []corev1.Volume{{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: kbParam.Pvc,
							},
						},
					}},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}

	_, err := config.KubernetesClient.AppsV1().Deployments(kbParam.Namespace).Create(s.ctx, deployment, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (s *KubernetesUtil) DeleteDeployment(kbParam *model.KubernetesParam) error {
	// 4. 删除 Deployment
	err := config.KubernetesClient.AppsV1().Deployments(kbParam.Namespace).Delete(s.ctx, kbParam.Deployment, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		log.Printf("删除 Deployment %s 失败: %v", kbParam.Deployment, err)
		return nil
	}
	return err
}

func (s *KubernetesUtil) ScaleDeployment(kbParam *model.KubernetesParam, replicas int32) error {
	// 获取现有的Deployment
	deployment, err := config.KubernetesClient.AppsV1().Deployments(kbParam.Namespace).Get(s.ctx, kbParam.Deployment, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取Deployment信息失败: %w", err)
	}

	// 修改副本数
	deployment.Spec.Replicas = &replicas

	// 更新Deployment
	_, err = config.KubernetesClient.AppsV1().Deployments(kbParam.Namespace).Update(s.ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新Deployment副本数失败: %w", err)
	}

	log.Printf("成功调整Deployment %s 的副本数为: %d", kbParam.Deployment, replicas)
	return nil
}

func (s *KubernetesUtil) CreatePod(kbParam *model.KubernetesParam, appParam *model.AppParam) error {
	// 3. 创建 Pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kbParam.Pod,
			Namespace: kbParam.Namespace,
			Labels: map[string]string{
				"app": "code-server",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				// 1. 原来的 code-server
				{
					Name:  "code-server",
					Image: "docker.1ms.run/linuxserver/code-server:4.103.0",
					Env: []corev1.EnvVar{
						{Name: "PUID", Value: "1000"},
						{Name: "PGID", Value: "1000"},
						{Name: "TZ", Value: "Etc/UTC"},
						{Name: "PASSWORD", Value: appParam.PodPassword},
						{Name: "SUDO_PASSWORD", Value: appParam.PodPassword},
						{Name: "PWA_APPNAME", Value: "code-server"},
					},
					Ports: []corev1.ContainerPort{{
						ContainerPort: 8443,
						Protocol:      corev1.ProtocolTCP,
						Name:          "https",
					}},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "data",
						MountPath: "/config",
					}},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(appParam.Cpu),
							corev1.ResourceMemory: resource.MustParse(appParam.Memory),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(appParam.Cpu),
							corev1.ResourceMemory: resource.MustParse(appParam.Memory),
						},
					},
				},
				// 2. 新增的 heartbeater
				{
					Name:            "heartbeater",
					Image:           "223.2.19.172:30002/minics/heartbeater:1.0.0",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Env: []corev1.EnvVar{
						{
							Name: "POD_NAME",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.name",
								},
							},
						},
						{
							Name: "NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.namespace",
								},
							},
						},
						{Name: "USER_ID", Value: "33"},
						{Name: "KUBERNETES_POD_START_TIME", Value: time.Now().UTC().Format("2006-01-02T15:04:05Z")},
						{Name: "KAFKA_TOPIC", Value: "timecounter"},
						{Name: "KAFKA_BROKERS", Value: "kafka.minics.svc.cluster.local:9092"},
					},
				},
			},
			Volumes: []corev1.Volume{{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: kbParam.Pvc,
					},
				},
			}},
			RestartPolicy: corev1.RestartPolicyAlways,
		},
	}

	_, err := config.KubernetesClient.CoreV1().Pods(kbParam.Namespace).Create(s.ctx, pod, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil // 已存在算成功
	}
	return err
}

func (s *KubernetesUtil) CreatePvc(kbParam *model.KubernetesParam, appParam *model.AppParam) error {
	storageClassName := "dynamic-hostpath"

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kbParam.Pvc,
			Namespace: kbParam.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("20Gi"),
				},
			},
			StorageClassName: &storageClassName,
		},
	}
	if _, err := config.KubernetesClient.CoreV1().PersistentVolumeClaims(kbParam.Namespace).Create(s.ctx, pvc, metav1.CreateOptions{}); err != nil && !errors.IsAlreadyExists(err) {
		log.Printf("创建 PVC 失败: %w", err)
		return fmt.Errorf("创建 PVC 失败: %w", err)
	}

	return nil
}

func (s *KubernetesUtil) CreateSvc(kbParam *model.KubernetesParam, application *model.Application) error {
	// 3. 构造 Service 对象
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kbParam.Svc,
			Namespace: kbParam.Namespace,
			Labels: map[string]string{
				"app":        "code-server",
				"deployment": kbParam.Deployment,
			},
		},
		Spec: corev1.ServiceSpec{
			// 绑定到哪些 Pod（通过 label 选择）
			Selector: map[string]string{
				"app":        "code-server",
				"deployment": kbParam.Deployment,
			},
			// 暴露的端口列表
			Ports: []corev1.ServicePort{
				{
					Name:       "web",
					Port:       443,                    // Service 自己的端口
					TargetPort: intstr.FromInt32(8443), // 目标 Pod 的端口
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeNodePort, // 可改成 NodePort / LoadBalancer
		},
	}

	// 4. 调用 API 创建
	result, err := config.KubernetesClient.CoreV1().
		Services(kbParam.Namespace).
		Create(s.ctx, svc, metav1.CreateOptions{})
	if err != nil {
		log.Printf("创建 Service 失败: %w", err)
		return err
	}

	nodePort := result.Spec.Ports[0].NodePort
	application.Url = fmt.Sprintf("http://223.2.19.172:%d", nodePort)
	log.Printf("分配的NodePort端口: %d", nodePort)

	return nil
}

func (s *KubernetesUtil) CreateSvcWithUpdate(kbParam *model.KubernetesParam, application *model.Application) error {
	// 3. 构造 Service 对象
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kbParam.Svc,
			Namespace: kbParam.Namespace,
			Labels: map[string]string{
				"app": "code-server",
			},
		},
		Spec: corev1.ServiceSpec{
			// 绑定到哪些 Pod（通过 label 选择）
			Selector: map[string]string{
				"app": "code-server",
			},
			// 暴露的端口列表
			Ports: []corev1.ServicePort{
				{
					Name:       "web",
					Port:       443,                    // Service 自己的端口
					TargetPort: intstr.FromInt32(8443), // 目标 Pod 的端口
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeNodePort, // 可改成 NodePort / LoadBalancer
		},
	}
	// 4. 调用 API 创建
	result, err := config.KubernetesClient.CoreV1().
		Services(kbParam.Namespace).
		Update(s.ctx, svc, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("创建 Service 失败: %w", err)
		return err
	}

	nodePort := result.Spec.Ports[0].NodePort
	application.Url = fmt.Sprintf("http://223.2.19.172:%d", nodePort)
	log.Printf("分配的NodePort端口: %d", nodePort)

	return nil
}

func (s *KubernetesUtil) DeleteSvc(kbParam *model.KubernetesParam) error {

	if err := config.KubernetesClient.CoreV1().Services(kbParam.Namespace).Delete(s.ctx, kbParam.Svc, metav1.DeleteOptions{}); err != nil {
		log.Printf("删除 Service 失败: %v", err)
		return fmt.Errorf("删除 Service 失败: %w", err)
	}
	return nil
}

func (s *KubernetesUtil) ModifyPodResource(kbParam *model.KubernetesParam) error {
	// 获取现有Pod
	pod, err := config.KubernetesClient.CoreV1().Pods(kbParam.Namespace).Get(s.ctx, kbParam.Pod, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取Pod信息失败: %w", err)
	}

	// 修改第一个容器的资源限制
	if len(pod.Spec.Containers) > 0 {
		pod.Spec.Containers[0].Resources.Limits = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(kbParam.Cpu),
			corev1.ResourceMemory: resource.MustParse(kbParam.Memory),
		}
		pod.Spec.Containers[0].Resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(kbParam.Cpu),
			corev1.ResourceMemory: resource.MustParse(kbParam.Memory),
		}
	}

	// 更新Pod
	_, err = config.KubernetesClient.CoreV1().Pods(kbParam.Namespace).Update(s.ctx, pod, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新Pod资源失败: %w", err)
	}

	log.Printf("成功修改Pod %s 的资源限制: CPU=%s, Memory=%s", kbParam.Pod, kbParam.Cpu, kbParam.Memory)
	return nil
}

func (s *KubernetesUtil) DeletePodSvc(kbParam *model.KubernetesParam) error {
	err := config.KubernetesClient.CoreV1().Pods(kbParam.Namespace).Delete(s.ctx, kbParam.Pod, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("删除 Pod 失败: %v", err)
		return fmt.Errorf("删除 Pod 失败: %w", err)
	}

	if err := config.KubernetesClient.CoreV1().Services(kbParam.Namespace).Delete(s.ctx, kbParam.Svc, metav1.DeleteOptions{}); err != nil {
		log.Printf("删除 Service 失败: %v", err)
		return fmt.Errorf("删除 Service 失败: %w", err)
	}

	return nil
}

func (s *KubernetesUtil) DeleteDeploymentSvcPvc(kbParam *model.KubernetesParam) error {
	err := config.KubernetesClient.AppsV1().Deployments(kbParam.Namespace).Delete(s.ctx, kbParam.Deployment, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		log.Printf("删除 Deployment %s 失败: %v", kbParam.Deployment, err)
		return fmt.Errorf("删除 Deployment %s 失败: %w", kbParam.Deployment, err)
	}

	if err := config.KubernetesClient.CoreV1().Services(kbParam.Namespace).Delete(s.ctx, kbParam.Svc, metav1.DeleteOptions{}); err != nil {
		log.Printf("删除 Service 失败: %v", err)
		return fmt.Errorf("删除 Service 失败: %w", err)
	}

	if err := config.KubernetesClient.CoreV1().PersistentVolumeClaims(kbParam.Namespace).Delete(s.ctx, kbParam.Pvc, metav1.DeleteOptions{}); err != nil {
		log.Printf("删除 PVC 失败: %v", err)
		return fmt.Errorf("删除 PVC 失败: %w", err)
	}

	log.Printf("删除app: %s 相关的所有资源", kbParam.Pod)

	return nil
}

func (s *KubernetesUtil) DeletePodSvcPvc(kbParam *model.KubernetesParam) error {
	if err := config.KubernetesClient.CoreV1().Pods(kbParam.Namespace).Delete(s.ctx, kbParam.Pod, metav1.DeleteOptions{}); err != nil {
		log.Printf("删除 Pod 失败: %v", err)
		return fmt.Errorf("删除 Pod 失败: %w", err)
	}

	if err := config.KubernetesClient.CoreV1().Services(kbParam.Namespace).Delete(s.ctx, kbParam.Svc, metav1.DeleteOptions{}); err != nil {
		log.Printf("删除 Service 失败: %v", err)
		return fmt.Errorf("删除 Service 失败: %w", err)
	}

	if err := config.KubernetesClient.CoreV1().PersistentVolumeClaims(kbParam.Namespace).Delete(s.ctx, kbParam.Pvc, metav1.DeleteOptions{}); err != nil {
		log.Printf("删除 PVC 失败: %v", err)
		return fmt.Errorf("删除 PVC 失败: %w", err)
	}

	log.Printf("删除pod: %s 相关的所有资源", kbParam.Pod)

	return nil
}

func (s *KubernetesUtil) GetPodInfo(kbParam *model.KubernetesParam) (*corev1.Pod, error) {
	// 使用Deployment的标签选择器获取Pod
	labelSelector := fmt.Sprintf("app=code-server,deployment=%s", kbParam.Deployment)

	pods, err := config.KubernetesClient.CoreV1().Pods(kbParam.Namespace).List(s.ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取Deployment的Pod列表失败: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("未找到Deployment %s 对应的Pod", kbParam.Deployment)
	}

	// 获取第一个Pod
	pod := pods.Items[0]
	return &pod, nil
}

func (s *KubernetesUtil) GetPodLog(kbParam *model.KubernetesParam) (string, error) {
	// 使用Deployment的标签选择器获取Pod
	labelSelector := fmt.Sprintf("app=code-server,deployment=%s", kbParam.Deployment)

	pods, err := config.KubernetesClient.CoreV1().Pods(kbParam.Namespace).List(s.ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return "", fmt.Errorf("获取Deployment的Pod列表失败: %w", err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("未找到Deployment %s 对应的Pod", kbParam.Deployment)
	}

	// 获取第一个Pod的日志
	pod := pods.Items[0]
	req := config.KubernetesClient.CoreV1().Pods(kbParam.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
	podLogs, err := req.Stream(s.ctx)
	if err != nil {
		return "", fmt.Errorf("获取Pod日志失败: %w", err)
	}

	defer func(podLogs io.ReadCloser) {
		err := podLogs.Close()
		if err != nil {
			log.Printf("关闭Pod日志流失败: %v", err)
		}
	}(podLogs)

	data, err := io.ReadAll(podLogs)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func CreateHttpRoute() {

}
