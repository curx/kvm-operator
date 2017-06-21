package kvm

import (
	"github.com/giantswarm/kvmtpr/kvm/dns"
	"github.com/giantswarm/kvmtpr/kvm/endpointupdater"
	"github.com/giantswarm/kvmtpr/kvm/flannel"
	"github.com/giantswarm/kvmtpr/kvm/k8skvm"
	"github.com/giantswarm/kvmtpr/kvm/kubectl"
	"github.com/giantswarm/kvmtpr/kvm/network"
)

type KVM struct {
	DNS             dns.DNS                         `json:"dns" yaml:"dns"`
	EndpointUpdater endpointupdater.EndpointUpdater `json:"endpointUpdater" yaml:"endpointUpdater"`
	Flannel         flannel.Flannel                 `json:"flannel" yaml:"flannel"`
	K8sKVM          k8skvm.K8sKVM                   `json:"k8sKVM" yaml:"k8sKVM"`
	Kubectl         kubectl.Kubectl                 `json:"kubectl" yaml:"kubectl"`
	Network         network.Network                 `json:"network" yaml:"network"`
}