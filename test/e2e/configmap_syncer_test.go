package e2e_test

import (
	"os"
	"time"

	api "github.com/appscode/kubed/apis/kubed/v1alpha1"
	"github.com/appscode/kubed/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/meta"
)

var _ = Describe("Config-Syncer", func() {
	var (
		f             *framework.Invocation
		cfgMap        *core.ConfigMap
		nsWithLabel   *core.Namespace
		stopCh        chan struct{}
		clusterConfig api.ClusterConfig
	)

	BeforeEach(func() {
		f = root.Invoke()
		cfgMap = f.NewConfigMap()
		nsWithLabel = f.NewNamespaceWithLabel()
	})

	JustBeforeEach(func() {
		if f.SelfHostedOperator {
			By("Restarting kubed operator")
			err := f.RestartKubedOperator(&clusterConfig)
			Expect(err).NotTo(HaveOccurred())
		} else {
			By("Starting Kubed")
			stopCh = make(chan struct{})
			err := f.RunKubed(stopCh, clusterConfig)
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for API server to be ready")
			root.EventuallyAPIServerReady().Should(Succeed())
			time.Sleep(time.Second * 5)
		}
	})

	AfterEach(func() {
		if !f.SelfHostedOperator {
			close(stopCh)
		}
		f.DeleteAllConfigmaps()

		err := f.KubeClient.CoreV1().Namespaces().Delete(nsWithLabel.Name, &metav1.DeleteOptions{})
		if kerr.IsNotFound(err) {
			err = nil
		}
		Expect(err).NotTo(HaveOccurred())
		f.EventuallyNamespaceDeleted(nsWithLabel.Name).Should(BeTrue())
	})

	var (
		shouldSyncConfigMapToAllNamespaces = func() {
			By("Creating configMap")
			sourceCM, err := f.KubeClient.CoreV1().ConfigMaps(cfgMap.Namespace).Create(cfgMap)
			Expect(err).NotTo(HaveOccurred())

			By("Checking configMap has not synced yet")
			f.EventuallyConfigMapNotSynced(sourceCM).Should(BeTrue())

			By("Adding sync annotation")
			sourceCM, _, err = core_util.PatchConfigMap(f.KubeClient, sourceCM, func(obj *core.ConfigMap) *core.ConfigMap {
				metav1.SetMetaDataAnnotation(&obj.ObjectMeta, api.ConfigSyncKey, "")
				return obj
			})
			Expect(err).ShouldNot(HaveOccurred())

			By("Checking configMap has synced to all namespaces")
			f.EventuallyConfigMapSynced(sourceCM).Should(BeTrue())
		}
	)

	Describe("Across Namespaces", func() {

		BeforeEach(func() {
			clusterConfig = framework.ConfigSyncClusterConfig()
		})

		Context("All Namespaces", func() {

			It("should sync configMap to all namespaces", shouldSyncConfigMapToAllNamespaces)
		})

		Context("New Namespace", func() {

			It("should synced configMap to new namespace", func() {
				shouldSyncConfigMapToAllNamespaces()

				By("Creating new namespace")
				err := f.CreateNamespace(nsWithLabel)
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking new namespace has the configMap")
				f.EventuallyConfigMapSyncedToNamespace(cfgMap, nsWithLabel.Name).Should(BeTrue())
			})
		})

		Context("Remove Sync Annotation", func() {

			It("should delete synced configMaps", func() {
				shouldSyncConfigMapToAllNamespaces()

				By("Removing sync annotation")
				source, err := f.KubeClient.CoreV1().ConfigMaps(cfgMap.Namespace).Get(cfgMap.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				_, _, err = core_util.PatchConfigMap(f.KubeClient, source, func(obj *core.ConfigMap) *core.ConfigMap {
					obj.Annotations = meta.RemoveKey(obj.Annotations, api.ConfigSyncKey)
					return obj
				})
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking synced configMaps has been deleted")
				f.EventuallySyncedConfigMapsDeleted(source)
			})
		})

		Context("Source Update", func() {

			It("should update synced configMaps", func() {
				shouldSyncConfigMapToAllNamespaces()

				By("Updating source configMap")
				source, err := f.KubeClient.CoreV1().ConfigMaps(cfgMap.Namespace).Get(cfgMap.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				source, _, err = core_util.PatchConfigMap(f.KubeClient, source, func(obj *core.ConfigMap) *core.ConfigMap {
					obj.Data["data"] = "test"
					return obj
				})
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking synced configMaps has been updated")
				f.EventuallySyncedConfigMapsUpdated(source).Should(BeTrue())
			})
		})

		Context("Backward Compatibility", func() {

			It("should sync configMap to all namespaces", func() {

				By("Creating configMap")
				source, err := f.CreateConfigMap(cfgMap)
				Expect(err).NotTo(HaveOccurred())

				By("Checking configMap has not synced yet")
				f.EventuallyConfigMapNotSynced(source).Should(BeTrue())

				By("Adding sync=true annotation")
				source, _, err = core_util.PatchConfigMap(f.KubeClient, source, func(obj *core.ConfigMap) *core.ConfigMap {
					metav1.SetMetaDataAnnotation(&obj.ObjectMeta, api.ConfigSyncKey, "true")
					return obj
				})
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking configMap has synced to all namespaces")
				f.EventuallyConfigMapSynced(source).Should(BeTrue())
			})
		})

		Context("Namespace Selector", func() {

			It("should add configMap to selected namespaces", func() {

				shouldSyncConfigMapToAllNamespaces()

				By("Adding selector annotation")
				source, err := f.KubeClient.CoreV1().ConfigMaps(cfgMap.Namespace).Get(cfgMap.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				source, _, err = core_util.PatchConfigMap(f.KubeClient, source, func(obj *core.ConfigMap) *core.ConfigMap {
					metav1.SetMetaDataAnnotation(&obj.ObjectMeta, api.ConfigSyncKey, "app="+f.App())
					return obj
				})
				Expect(err).NotTo(HaveOccurred())

				By("Checking configMap has not synced to other namespaces")
				f.EventuallyConfigMapNotSynced(source).Should(BeTrue())

				By("Creating new namespace with label")
				err = f.CreateNamespace(nsWithLabel)
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking configmap synced to new namespace")
				f.EventuallyConfigMapSyncedToNamespace(source, nsWithLabel.Name)

				By("Changing selector annotation")
				_, _, err = core_util.PatchConfigMap(f.KubeClient, source, func(obj *core.ConfigMap) *core.ConfigMap {
					metav1.SetMetaDataAnnotation(&obj.ObjectMeta, api.ConfigSyncKey, "app=do-not-match")
					return obj
				})
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking synced configMap has been deleted")
				f.EventuallySyncedConfigMapsDeleted(source)

				By("Removing selector annotation")
				source, err = f.KubeClient.CoreV1().ConfigMaps(source.Namespace).Get(source.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				source, _, err = core_util.PatchConfigMap(f.KubeClient, source, func(obj *core.ConfigMap) *core.ConfigMap {
					metav1.SetMetaDataAnnotation(&obj.ObjectMeta, api.ConfigSyncKey, "")
					return obj
				})
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking configMap synced to all namespaces")
				f.EventuallyConfigMapSynced(source).Should(BeTrue())
			})
		})

		Context("Source Deleted", func() {

			It("should delete synced configMaps", func() {
				shouldSyncConfigMapToAllNamespaces()

				By("Deleting source configMap")
				source, err := f.KubeClient.CoreV1().ConfigMaps(cfgMap.Namespace).Get(cfgMap.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				err = f.DeleteConfigMap(source.ObjectMeta)
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking synced configMaps has been deleted")
				f.EventuallySyncedConfigMapsDeleted(source).Should(BeTrue())
			})
		})

		Context("Source Namespace Deleted", func() {
			var sourceNamespace *core.Namespace

			BeforeEach(func() {
				sourceNamespace = f.NewNamespace("source")
				cfgMap.Namespace = sourceNamespace.Name
			})

			It("should delete synced configMaps", func() {

				By("Creating source namespace")
				err := f.CreateNamespace(sourceNamespace)
				Expect(err).NotTo(HaveOccurred())

				shouldSyncConfigMapToAllNamespaces()

				By("Deleting source namespace")
				source, err := f.KubeClient.CoreV1().ConfigMaps(cfgMap.Namespace).Get(cfgMap.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				err = f.DeleteNamespace(sourceNamespace.Name)
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking synced configMaps has been deleted")
				f.EventuallySyncedConfigMapsDeleted(source).Should(BeTrue())
			})
		})
	})

	Describe("Across Cluster", func() {
		Context("ConfigMap Context Syncer Test", func() {
			var (
				kubeConfigPath = "/home/dipta/all/kubed-test/kubeconfig"
				context        = "gke_tigerworks-kube_us-central1-f_kite"
			)

			BeforeEach(func() {
				clusterConfig = framework.ConfigSyncClusterConfig()
				clusterConfig.ClusterName = "minikube"
				clusterConfig.KubeConfigFile = kubeConfigPath

				if _, err := os.Stat(kubeConfigPath); err != nil {
					Skip(`"config" file not found on` + kubeConfigPath)
				}

				By("Creating namespace for context")
				f.EnsureNamespaceForContext(kubeConfigPath, context)
			})

			AfterEach(func() {
				By("Deleting namespaces for contexts")
				f.DeleteNamespaceForContext(kubeConfigPath, context)
			})

			It("Should add configmap to contexts", func() {
				By("Creating source ns in remote cluster")
				f.EnsureNamespaceForContext(kubeConfigPath, context)

				By("Creating configmap")
				cfgMap, err := f.KubeClient.CoreV1().ConfigMaps(cfgMap.Namespace).Create(cfgMap)
				Expect(err).NotTo(HaveOccurred())

				By("Adding sync annotation")
				cfgMap, _, err = core_util.PatchConfigMap(f.KubeClient, cfgMap, func(obj *core.ConfigMap) *core.ConfigMap {
					metav1.SetMetaDataAnnotation(&obj.ObjectMeta, api.ConfigSyncContexts, context)
					return obj
				})
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking configmap added to contexts")
				f.EventuallyNumOfConfigmapsForContext(kubeConfigPath, context).Should(BeNumerically("==", 1))

				By("Removing sync annotation")
				cfgMap, _, err = core_util.PatchConfigMap(f.KubeClient, cfgMap, func(obj *core.ConfigMap) *core.ConfigMap {
					obj.Annotations = meta.RemoveKey(obj.Annotations, api.ConfigSyncContexts)
					return obj
				})
				Expect(err).ShouldNot(HaveOccurred())

				By("Checking configmap removed from contexts")
				f.EventuallyNumOfConfigmapsForContext(kubeConfigPath, context).Should(BeNumerically("==", 0))
			})
		})
	})
})
