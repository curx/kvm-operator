package v12

import (
	"context"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/v12/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v12/key"
	"github.com/giantswarm/kvm-operator/service/controller/v12/resource/clusterrolebinding"
	"github.com/giantswarm/kvm-operator/service/controller/v12/resource/configmap"
	"github.com/giantswarm/kvm-operator/service/controller/v12/resource/deployment"
	"github.com/giantswarm/kvm-operator/service/controller/v12/resource/ingress"
	"github.com/giantswarm/kvm-operator/service/controller/v12/resource/namespace"
	"github.com/giantswarm/kvm-operator/service/controller/v12/resource/pvc"
	"github.com/giantswarm/kvm-operator/service/controller/v12/resource/service"
	"github.com/giantswarm/kvm-operator/service/controller/v12/resource/serviceaccount"
)

type ClusterResourceSetConfig struct {
	CertsSearcher      certs.Interface
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher randomkeys.Interface

	OIDC               cloudconfig.OIDCConfig
	GuestUpdateEnabled bool
	ProjectName        string
}

func NewClusterResourceSet(config ClusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var cloudConfig *cloudconfig.CloudConfig
	{
		c := cloudconfig.Config{
			Logger: config.Logger,

			OIDC: config.OIDC,
		}

		cloudConfig, err = cloudconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterRoleBindingResource controller.Resource
	{
		c := clusterrolebinding.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ops, err := clusterrolebinding.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterRoleBindingResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource controller.Resource
	{
		c := namespace.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := namespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		namespaceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceAccountResource controller.Resource
	{
		c := serviceaccount.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := serviceaccount.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		serviceAccountResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource controller.Resource
	{
		c := configmap.Config{
			CertSearcher: config.CertsSearcher,
			CloudConfig:  cloudConfig,
			K8sClient:    config.K8sClient,
			KeyWatcher:   config.RandomkeysSearcher,
			Logger:       config.Logger,
		}

		ops, err := configmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMapResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource controller.Resource
	{
		c := deployment.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := deployment.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		deploymentResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ingressResource controller.Resource
	{
		c := ingress.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := ingress.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		ingressResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pvcResource controller.Resource
	{
		c := pvc.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := pvc.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		pvcResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource controller.Resource
	{
		c := service.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := service.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		serviceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		clusterRoleBindingResource,
		namespaceResource,
		serviceAccountResource,
		configMapResource,
		deploymentResource,
		ingressResource,
		pvcResource,
		serviceResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{
			Name: config.ProjectName,
		}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	handlesFunc := func(obj interface{}) bool {
		kvmConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(kvmConfig) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		if config.GuestUpdateEnabled {
			updateallowedcontext.SetUpdateAllowed(ctx)
		}

		return ctx, nil
	}

	var clusterResourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		clusterResourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return clusterResourceSet, nil
}

func toCRUDResource(logger micrologger.Logger, ops controller.CRUDResourceOps) (*controller.CRUDResource, error) {
	c := controller.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := controller.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
