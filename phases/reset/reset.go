package preflight

import (
	"context"
	"fmt"
	"github.com/yuyicai/kubei/config/constants"

	"github.com/go-kratos/kratos/pkg/sync/errgroup"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/tmpl"
	"k8s.io/klog"
)

func ResetKubeadm(nodes []*rundata.Node, apiDomainName string) error {
	return resetKubeadm(nodes, apiDomainName)
}

func resetKubeadm(nodes []*rundata.Node, apiDomainName string) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
			klog.V(2).Infof("[%s] [reset] Resetting node", node.HostInfo.Host)
			if err := resetkubeadmOnNode(node, apiDomainName); err != nil {
				return fmt.Errorf("[%s] [reset] Failed to reset node: %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [reset] Successfully reset node", node.HostInfo.Host)
			return nil
		})
	}

	return g.Wait()
}

func resetkubeadmOnNode(node *rundata.Node, apiDomainName string) error {
	if err := node.Run("yes | kubeadm reset"); err != nil {
		return err
	}

	return node.Run(tmpl.ResetHosts(apiDomainName))
}

func RemoveKubeComponente(nodes []*rundata.Node) error {
	return removeKubeComponente(nodes)
}

func removeKubeComponente(nodes []*rundata.Node) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
			klog.V(2).Infof("[%s] [remove] remove the kubernetes component from the node", node.HostInfo.Host)
			if err := removeKubeComponentOnNode(node); err != nil {
				return fmt.Errorf("[%s] [remove] Failed to remove the kubernetes component: %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [remove] Successfully remove the kubernetes component from the node", node.HostInfo.Host)
			return nil
		})
	}

	return g.Wait()
}

func removeKubeComponentOnNode(node *rundata.Node) error {
	cmdTmpl := tmpl.NewKubeText(node.PackageManagementType)
	return node.Run(cmdTmpl.RemoveKubeComponent())
}

func RemoveContainerEngine(nodes []*rundata.Node) error {
	return removeContainerEngine(nodes)
}

func removeContainerEngine(nodes []*rundata.Node) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
			klog.V(2).Infof("[%s] [remove] Remove container engine from the node", node.HostInfo.Host)
			if err := removeContainerEngineOnNode(node); err != nil {
				return fmt.Errorf("[%s] [remove] Failed to remove container engine: %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [remove] Successfully remove container engine", node.HostInfo.Host)
			return nil
		})
	}

	return g.Wait()
}

func removeContainerEngineOnNode(node *rundata.Node) error {
	cmdTmpl := tmpl.NewContainerEngineText(node.PackageManagementType)
	return node.Run(cmdTmpl.RemoveDocker())
}
