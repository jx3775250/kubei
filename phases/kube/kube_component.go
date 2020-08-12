package kube

import (
	"fmt"

	"github.com/fatih/color"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"github.com/yuyicai/kubei/tmpl"
)

func InstallKubeComponent(c *rundata.Cluster) error {
	color.HiBlue("Installing Kubernetes component ☸️")
	return c.RunOnAllNodes(func(node *rundata.Node) error {
		klog.V(2).Infof("[%s] [kube] Installing Kubernetes component", node.HostInfo.Host)
		if err := installKubeComponent(c.Kubernetes.Version, node); err != nil {
			return fmt.Errorf("[%s] [kube] Failed to install Kubernetes component: %v", node.HostInfo.Host, err)
		}

		if err := system.Restart("kubelet", node); err != nil {
			return err
		}
		fmt.Printf("[%s] [kube] install Kubernetes component: %s\n", node.HostInfo.Host, color.HiGreenString("done✅️"))
		return nil
	})
}

func installKubeComponent(version string, node *rundata.Node) error {

	cmdTmpl := tmpl.NewKubeText(node.PackageManagementType)
	cmd, err := cmdTmpl.KubeComponent(version, node.InstallType)
	if err != nil {
		return err
	}

	return node.Run(cmd)

}
