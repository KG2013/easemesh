package controlpanel

import (
	"fmt"
	"strings"
	"time"

	installbase "github.com/megaease/easemeshctl/cmd/client/command/meshinstall/base"
	"github.com/megaease/easemeshctl/cmd/common"
	"github.com/megaease/easemeshctl/cmd/common/client"
	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
)

// Deploy will deploy resource of control panel
func Deploy(context *installbase.StageContext) error {

	installFuncs := []installbase.InstallFunc{
		namespaceSpec(&context.Arguments),
		configMapSpec(&context.Arguments),
		serviceSpec(&context.Arguments),
		statefulsetSpec(&context.Arguments),
	}

	err := installbase.BatchDeployResources(context.Cmd, context.Client, &context.Arguments, installFuncs)
	if err != nil {
		return errors.Wrap(err, "deploy mesh control panel resource error")
	}

	err = checkEasegressControlPlaneStatus(context.Cmd, context.Client, &context.Arguments)
	if err != nil {
		return errors.Wrap(err, "check mesh control panel status error")
	}

	err = provisionEaseMeshControlPanel(context.Cmd, context.Client, &context.Arguments)
	if err != nil {
		return errors.Wrap(err, "provision mesh control panel error")
	}
	return nil
}

// PreCheck will check prerequisite for installing control plane
func PreCheck(context *installbase.StageContext) error {
	var err error

	// 1. check available PersistentVolume
	pvList, err := installbase.ListPersistentVolume(context.Client)
	if err != nil {
		return err
	}

	availablePVCount := 0
	quantity := resource.MustParse(context.Arguments.MeshControlPlanePersistVolumeCapacity)
	boundedPVCSuffixes := []string{}
	for i := 0; i < context.Arguments.EasegressControlPlaneReplicas; i++ {
		boundedPVCSuffixes = append(boundedPVCSuffixes, fmt.Sprintf("%s-%d", installbase.DefaultMeshControlPlaneName, i))

	}
	for _, pv := range pvList.Items {
		if pv.Status.Phase == v1.VolumeAvailable &&
			pv.Spec.StorageClassName == context.Arguments.MeshControlPlaneStorageClassName &&
			pv.Spec.Capacity.Storage().Cmp(quantity) >= 0 &&
			checkPVAccessModes(v1.ReadWriteOnce, &pv) {
			availablePVCount += 1
		} else if pv.Status.Phase == v1.VolumeBound {
			// If PV already bound to PVC of EaseMesh controlpanel
			// we regarded it as availablePVCount
			for _, pvNameSuffix := range boundedPVCSuffixes {
				if pv.Spec.ClaimRef.Kind == "PersistentVolumeClaim" &&
					pv.Spec.ClaimRef.Namespace == context.Arguments.MeshNameSpace &&
					strings.HasSuffix(pv.Spec.ClaimRef.Name, pvNameSuffix) {
					availablePVCount++
					break
				}
			}
		}
	}

	if availablePVCount < context.Arguments.EasegressControlPlaneReplicas {
		return errors.Errorf(installbase.MeshControlPlanePVNotExistedHelpStr,
			context.Arguments.EasegressControlPlaneReplicas,
			availablePVCount,
			context.Arguments.MeshControlPlaneStorageClassName,
			context.Arguments.MeshControlPlaneStorageClassName,
			installbase.DefaultMeshControlPlanePersistVolumeCapacity)
	}

	return nil

}

// Clear will clear all installed resource about control panel
func Clear(context *installbase.StageContext) error {
	statefulsetResource := [][]string{
		{"statefulsets", installbase.DefaultMeshControlPlaneName},
	}
	coreV1Resources := [][]string{
		{"services", context.Arguments.EgServiceName},
		{"services", installbase.DefaultMeshControlPlanePlubicServiceName},
		{"services", installbase.DefaultMeshControlPlaneHeadlessServiceName},
		{"configmaps", installbase.DefaultMeshControlPlaneConfig},
	}

	clearEaseMeshControlPanelProvision(context.Cmd, context.Client, &context.Arguments)

	installbase.DeleteResources(context.Client, statefulsetResource, context.Arguments.MeshNameSpace, installbase.DeleteStatefulsetResource)
	installbase.DeleteResources(context.Client, coreV1Resources, context.Arguments.MeshNameSpace, installbase.DeleteCoreV1Resource)
	return nil
}

// Describe leverage human-readable text to describe different phase
// in the process of the control plane installation
func Describe(context *installbase.StageContext, phase installbase.InstallPhase) string {
	switch phase {
	case installbase.BeginPhase:
		return fmt.Sprintf("Begin to install mesh control plane service in the namespace %s", context.Arguments.MeshNameSpace)
	case installbase.EndPhase:
		return fmt.Sprintf("\nControl panel statefulset %s\n%s", installbase.DefaultMeshControlPlaneName,
			installbase.FormatPodStatus(context.Client, context.Arguments.MeshNameSpace,
				installbase.AdaptListPodFunc(meshControlPanelLabel())))
	}
	return ""
}

func checkPVAccessModes(accessModel v1.PersistentVolumeAccessMode, volume *v1.PersistentVolume) bool {
	for _, mode := range volume.Spec.AccessModes {
		if mode == accessModel {
			return true
		}
	}
	return false
}

func checkEasegressControlPlaneStatus(cmd *cobra.Command, kubeClient *kubernetes.Clientset, args *installbase.InstallArgs) error {

	// Wait a fix time for the Easegress cluster to start
	time.Sleep(time.Second * 10)

	entrypoints, err := installbase.GetMeshControlPanelEntryPoints(kubeClient, args.MeshNameSpace,
		installbase.DefaultMeshControlPlanePlubicServiceName,
		installbase.DefaultMeshAdminPortName)
	if err != nil {
		return errors.Wrap(err, "get mesh control plane entrypoint failed")
	}

	timeOutPerTry := args.MeshControlPlaneCheckHealthzMaxTime / len(entrypoints)

	for i := 0; i < len(entrypoints); i++ {
		_, err := client.NewHTTPJSON(
			client.WrapRetryOptions(3, time.Second*time.Duration(timeOutPerTry)/3, func(body []byte, err error) bool {
				if err != nil && strings.Contains(err.Error(), "connection refused") {
					return true
				}

				members, err := unmarshalMember(body)
				if err != nil {
					common.OutputErrorInfo("parse member body error: %s", err)
					return true
				}

				return len(members) < (args.EaseMeshOperatorReplicas/2 + 1)
			})...).
			Get(entrypoints[i]+installbase.MemberList, nil, time.Second*time.Duration(timeOutPerTry), nil).
			HandleResponse(func(body []byte, statusCode int) (interface{}, error) {
				if statusCode != 200 {
					return nil, errors.Errorf("check control plane member list error, return status code is :%d", statusCode)
				}
				members, err := unmarshalMember(body)

				if err != nil {
					return nil, err
				}

				if len(members) < (args.EasegressControlPlaneReplicas/2 + 1) {
					return nil, errors.Errorf("easemesh control plane is not ready, expect %d of replicas, but %d", args.EasegressControlPlaneReplicas, len(members))
				}
				return nil, nil
			})
		if err != nil {
			common.OutputErrorInfo("check mesh control plane status failed, ignored check next node, current error is: %s", err)
		} else {
			return nil
		}
	}
	return errors.Errorf("mesh control plane is not ready")
}

func unmarshalMember(body []byte) ([]map[string]interface{}, error) {
	var options []map[string]interface{}
	err := yaml.Unmarshal(body, &options)
	if err != nil {
		return nil, err
	}
	return options, nil
}