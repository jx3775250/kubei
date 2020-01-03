package phases

import (
	"errors"
	"github.com/yuyicai/kubei/config/options"
	kubeadmphases "github.com/yuyicai/kubei/phases/kubeadm"
	networkphases "github.com/yuyicai/kubei/phases/network"
	"github.com/yuyicai/kubei/preflight"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	"time"
)

// NewKubeadmPhase creates a kubei workflow phase that implements handling of kubeadm.
func NewKubeadmPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "kubeadm",
		Short:        "create a k8s cluster with kubeadm",
		Long:         "create a k8s cluster with kubeadm",
		InheritFlags: getKubeadmPhaseFlags(),
		Run:          runKubeadm,
	}
	return phase
}

func getKubeadmPhaseFlags() []string {
	flags := []string{
		options.JumpServer,
		options.ControlPlaneEndpoint,
		options.ImageRepository,
		options.PodNetworkCidr,
		options.ServiceCidr,
		options.Masters,
		options.Workers,
		options.Password,
		options.Port,
		options.User,
		options.Key,
	}
	return flags
}

func runKubeadm(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("kubeadm phase invoked with an invalid rundata struct")
	}

	cfg := data.Cfg()
	cluster := cfg.ClusterNodes
	masters := cluster.Masters
	workers := cluster.Worker
	nodes := append(masters, workers...)
	kubeadmCfg := data.KubeadmCfg()

	if len(masters) == 0 {
		return errors.New("You host to set the master nodes by the flag \"--masters\"")
	}

	if err := preflight.CheckSSH(nodes, &cfg.JumpServer); err != nil {
		return err
	}

	// init master0
	if err := kubeadmphases.InitMaster(masters[0], kubeadmCfg); err != nil {
		return err
	}

	// add network plugin
	if err := networkphases.Flannel(masters[0], kubeadmCfg.Networking.PodSubnet, "quay.azk8s.cn/coreos/flannel:v0.11.0-amd64", "vxlan"); err != nil {
		return err
	}

	g := errgroup.Group{}

	if len(masters) > 1 {
		cfg.IsHA = true
		g.Go(func() error {
			if err := kubeadmphases.JoinControlPlane(masters, kubeadmCfg); err != nil {
				return err
			}
			return nil
		})
	}

	g.Go(func() error {
		if err := kubeadmphases.JoinNode(cfg, kubeadmCfg); err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	interval := 2 * time.Second
	timeout := 6 * time.Minute
	output, done := kubeadmphases.CheckNodesReady(masters[0], interval, timeout)
	if done {
		klog.Info("Done\n", output)
	}

	return nil
}