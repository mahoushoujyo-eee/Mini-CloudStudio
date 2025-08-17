package util

import (
	"context"
	"fmt"

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

func (s *KubernetesUtil) GetPods(param model.KubernetesParam) {
	// 获取 default namespace 下的所有 Pod
	pods, err := config.KubernetesClient.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d pods in 'default' namespace:\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(" -", pod.Name)
	}
}

func (s *KubernetesUtil) CreatePods(param model.KubernetesParam) {
	// 3. 定义 Pod 对象
	// 3. 定义 Pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "code-server",
			Namespace: "default", // 可改为其他命名空间
			Labels: map[string]string{
				"app": "code-server",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "code-server",
					Image: "linuxserver/code-server:4.103.0",
					Env: []corev1.EnvVar{
						{Name: "PUID", Value: "1000"},
						{Name: "PGID", Value: "1000"},
						{Name: "TZ", Value: "Etc/UTC"},
						{Name: "PASSWORD", Value: "yuzaoqian123"},
						{Name: "SUDO_PASSWORD", Value: "root"},
						// {Name: "PROXY_DOMAIN", Value: "code-server.my.domain"},
						// {Name: "DEFAULT_WORKSPACE", Value: "/config/workspace"},
						{Name: "PWA_APPNAME", Value: "code-server"},
					},
					Ports: []corev1.ContainerPort{
						{ContainerPort: 8443, Protocol: "TCP", Name: "http"},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "config-volume",
							MountPath: "/config",
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("1024Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1000m"),
							corev1.ResourceMemory: resource.MustParse("2048Mi"),
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "config-volume",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "code-server-pvc",
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyAlways, // 等价于 --restart unless-stopped
		},
	}

	// 4. 调用 API 创建 Pod
	_, err := config.KubernetesClient.CoreV1().Pods("default").Create(
		context.TODO(),
		pod,
		metav1.CreateOptions{},
	)
	if err != nil {
		// 判断是否已存在
		if errors.IsAlreadyExists(err) {
			fmt.Println("Pod 已存在")
		} else {
			panic(err)
		}
	} else {
		fmt.Println("Pod 创建成功")
	}
}

func (s *KubernetesUtil) CreatePvc(param model.KubernetesParam) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "code-server-pvc",
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
			},
		},
	}

	// 4. 调用 API 创建
	created, err := config.KubernetesClient.CoreV1().
		PersistentVolumeClaims("default").
		Create(context.TODO(), pvc, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("PVC %s created at %s\n", created.Name, created.CreationTimestamp)
}

func PodForward() {

}

func (s *KubernetesUtil) CreateSvc(param model.KubernetesParam) {
	// 3. 构造 Service 对象
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "code-server-svc",
			Namespace: "default",
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
					Port:       80,                     // Service 自己的端口
					TargetPort: intstr.FromInt32(8443), // 目标 Pod 的端口
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP, // 可改成 NodePort / LoadBalancer
		},
	}

	// 4. 调用 API 创建
	result, err := config.KubernetesClient.CoreV1().
		Services("default").
		Create(context.TODO(), svc, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Service %q created, clusterIP=%s\n",
		result.GetName(), result.Spec.ClusterIP)
}

func CreateHttpRoute() {

}
