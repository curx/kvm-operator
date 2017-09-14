package legacy

import (
	"github.com/giantswarm/kvm-operator/service/key"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const (
	// Name is the identifier of the resource.
	Name = "legacy"
)

// Config represents the configuration used to create a new reconciler.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Settings.
	// Resources is the list of resources to be processed during creation and
	// deletion reconciliations. The current reconciliation is synchronous and
	// processes resources in a series. One resource after another will be
	// processed.
	Resources []Resource
}

// DefaultConfig provides a default configuration to create a new reconciler by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,

		// Settings.
		Resources: nil,
	}
}

// New creates a new configured reconciler.
func New(config Config) (*Reconciler, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	if len(config.Resources) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "config.Resources must not be empty")
	}

	newReconciler := &Reconciler{
		// Dependencies.
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		// Settings
		resources: config.Resources,
	}

	return newReconciler, nil
}

// Reconciler implements the reconciler.
type Reconciler struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	// Settings.
	resources []Resource
}

func (r *Reconciler) GetCurrentState(obj interface{}) (interface{}, error) {
	return nil, nil
}

func (r *Reconciler) GetDesiredState(obj interface{}) (interface{}, error) {
	return nil, nil
}

func (r *Reconciler) GetCreateState(obj, currentState, desiredState interface{}) (interface{}, error) {
	return nil, nil
}

func (r *Reconciler) GetDeleteState(obj, currentState, desiredState interface{}) (interface{}, error) {
	return nil, nil
}

func (r *Reconciler) Name() string {
	return Name
}

func (r *Reconciler) ProcessCreateState(obj, createState interface{}) error {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}
	namespace := key.ClusterNamespace(*customObject)

	r.logger.Log("debug", "executing the reconciler's add function", "event", "create")

	var runtimeObjects []runtime.Object

	for _, res := range r.resources {
		ros, err := res.GetForCreate(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		runtimeObjects = append(runtimeObjects, ros...)
	}

	for _, ro := range runtimeObjects {
		var err error

		switch t := ro.(type) {
		case *v1.ConfigMap:
			_, err = r.k8sClient.Core().ConfigMaps(namespace).Create(t)
		case *v1beta1.Deployment:
			_, err = r.k8sClient.Extensions().Deployments(namespace).Create(t)
		default:
			return microerror.Maskf(executionFailedError, "unknown type '%T'", t)
		}

		if apierrors.IsAlreadyExists(err) {
			// Resource already being created, all good.
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Reconciler) ProcessDeleteState(obj, deleteState interface{}) error {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}
	namespace := key.ClusterNamespace(*customObject)

	r.logger.Log("debug", "executing the reconciler's delete function", "event", "delete")

	var runtimeObjects []runtime.Object

	for _, res := range r.resources {
		ros, err := res.GetForDelete(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		runtimeObjects = append(runtimeObjects, ros...)
	}

	for _, ro := range runtimeObjects {
		var err error

		switch t := ro.(type) {
		case *v1.ConfigMap:
			err = r.k8sClient.Core().ConfigMaps(namespace).Delete(t.Name, nil)
		case *v1beta1.Deployment:
			err = r.k8sClient.Extensions().Deployments(namespace).Delete(t.Name, nil)
		default:
			return microerror.Maskf(executionFailedError, "unknown type '%T'", t)
		}

		if apierrors.IsNotFound(err) {
			// Resource already being deleted, all good.
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Reconciler) Underlying() framework.Resource {
	return r
}