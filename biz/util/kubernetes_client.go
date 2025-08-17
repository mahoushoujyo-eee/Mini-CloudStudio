package util

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	pods, err := config.KubernetesClient.CoreV1().Pods(param.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d pods in 'default' namespace:\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(" -", pod.Name)
	}
}

func (s *KubernetesUtil) CreateCodeServerPod(kbParam *model.KubernetesParam, appParam *model.AppParam) error {
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
			Containers: []corev1.Container{{
				Name:  "code-server",
				Image: "linuxserver/code-server:4.103.0",
				Env: []corev1.EnvVar{
					{Name: "PUID", Value: "1000"},
					{Name: "PGID", Value: "1000"},
					{Name: "TZ", Value: "Etc/UTC"},
					{Name: "PASSWORD", Value: "password"}, // 让前端传
					{Name: "SUDO_PASSWORD", Value: "root"},
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
	}

	_, err := config.KubernetesClient.CoreV1().Pods(kbParam.Namespace).Create(s.ctx, pod, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil // 已存在算成功
	}
	return err
}

func (s *KubernetesUtil) CreatePvc(kbParam *model.KubernetesParam, appParam *model.AppParam) error {
	// 2. 创建 PVC（如已存在则忽略）
	pvcName := fmt.Sprintf("code-server-%s", uuid.NewString()[:8])
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: kbParam.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
			},
			// 不写 StorageClassName 就走默认 openebs-hostpath
		},
	}
	if _, err := config.KubernetesClient.CoreV1().PersistentVolumeClaims(kbParam.Namespace).Create(s.ctx, pvc, metav1.CreateOptions{}); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("创建 PVC 失败: %w", err)
	}
	kbParam.Pvc = pvcName

	return nil
}

func PodForward() {

}

func (s *KubernetesUtil) CreateSvc(kbParam *model.KubernetesParam, appParam *model.AppParam) error {
	// 3. 构造 Service 对象
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "code-server-svc",
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
			Type: corev1.ServiceTypeClusterIP, // 可改成 NodePort / LoadBalancer
		},
	}

	// 4. 调用 API 创建
	result, err := config.KubernetesClient.CoreV1().
		Services(kbParam.Namespace).
		Create(s.ctx, svc, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Service %q created, clusterIP=%s\n", result.GetName(), result.Spec.ClusterIP)

	return nil
}

func CreateHttpRoute() {

}
