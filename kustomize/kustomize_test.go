package kustomize_test

import (
	"encoding/json"
	"errors"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	cache "github.com/thallesfreitaszup/lib-kustomize/cache"
	mocksCache "github.com/thallesfreitaszup/lib-kustomize/cache/mocks"
	"github.com/thallesfreitaszup/lib-kustomize/kustomize"
	"github.com/thallesfreitaszup/lib-kustomize/kustomize/mocks"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/http"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var _ = Describe("Kustomize", func() {
	var source string
	var destination string
	var path string
	var getter *mocks.Getter
	var renderer *mocks.Renderer
	var cacheWrapper cache.Wrapper
	var mockCache *mocksCache.Cache
	var mockHttp *mocksCache.HttpClient
	BeforeEach(func() {

		source = "example.com/test"
		destination = "/destination"
		path = "path"
		getter = new(mocks.Getter)
		renderer = new(mocks.Renderer)
		mockCache = new(mocksCache.Cache)
		mockHttp = new(mocksCache.HttpClient)
		cacheWrapper = cache.New(mockCache, mockHttp)
	})
	Context("when fails to download repository content", func() {
		It("should return error", func() {
			mockCache.On("Get", source).Return(nil, false)
			mockHttp.On("Do", mock.Anything).Return(GetHTTPResponseWithStatusNotModified(), nil)
			mockCache.On("Set", source, mock.Anything, int64(1)).Times(1).Return(true)
			error := errors.New("failed to download resource")

			getter.On("Get").Return(error)

			k := kustomize.New(renderer, getter, destination, source, path, cacheWrapper)
			manifests, renderError := k.Render()
			assert.Equal(GinkgoT(), renderError, error)
			assert.Equal(GinkgoT(), len(manifests), 0)
		})
	})

	Context("when fails to render manifests", func() {
		It("should return error", func() {

			error := errors.New("failed to render resource")

			getter.On("Get").Return(nil)
			renderer.On("Run", filesys.MakeFsOnDisk(), filepath.Join(destination, path)).Return(resmap.New(), error)
			mockCache.On("Get", source).Return(nil, false)
			mockHttp.On("Do", mock.Anything).Return(GetHTTPResponseWithStatusNotModified(), nil)
			mockCache.On("Set", source, mock.Anything, int64(1)).Times(1).Return(true)
			k := kustomize.New(renderer, getter, destination, source, path, cacheWrapper)
			manifests, renderError := k.Render()
			assert.Equal(GinkgoT(), renderError, error)
			assert.Equal(GinkgoT(), len(manifests), 0)
		})
	})

	Context("when successfully render manifests", func() {
		It("should return the correct unstructured manifests", func() {

			getter.On("Get").Return(nil)
			renderer.On("Run", filesys.MakeFsOnDisk(), filepath.Join(destination, path)).Return(getManifestsResponseMap(), nil)
			mockCache.On("Get", source).Return("123", true)
			mockHttp.On("Do", mock.Anything).Return(GetHTTPResponseWithStatusModified(), nil)
			mockCache.On("Set", "123", getManifestsUnstructured(), int64(1)).Times(1).Return(true)
			k := kustomize.New(renderer, getter, destination, source, path, cacheWrapper)
			manifests, renderError := k.Render()
			assert.Equal(GinkgoT(), renderError, nil)
			assert.Equal(GinkgoT(), len(manifests), 2)
			assert.Equal(GinkgoT(), manifests[0].GetName(), "deploy1")
			assert.Equal(GinkgoT(), manifests[0].GetKind(), "Deployment")
			assert.Equal(GinkgoT(), manifests[0].GetAPIVersion(), "apps/v1")
			assert.Equal(GinkgoT(), manifests[1].GetAPIVersion(), "apps/v1")
			assert.Equal(GinkgoT(), manifests[1].GetKind(), "Deployment")
			assert.Equal(GinkgoT(), manifests[1].GetName(), "deploy2")
		})
	})

	Context("when fails to add manifests in cache", func() {
		It("should return error", func() {

			getter.On("Get").Return(nil)
			renderer.On("Run", filesys.MakeFsOnDisk(), filepath.Join(destination, path)).Return(getManifestsResponseMap(), nil)
			mockCache.On("Get", source).Return("123", true)
			mockHttp.On("Do", mock.Anything).Return(GetHTTPResponseWithStatusModified(), nil)
			mockCache.On("Set", "123", getManifestsUnstructured(), int64(1)).Times(1).Return(false)
			k := kustomize.New(renderer, getter, destination, source, path, cacheWrapper)
			manifests, renderError := k.Render()
			assert.Equal(GinkgoT(), renderError, errors.New("failed to set manifests to cache"))
			assert.Equal(GinkgoT(), len(manifests), 0)

		})
	})

	Context("when successfully  get manifests in cache", func() {
		It("should return manifests", func() {
			etag := "dummy-etag"
			getter.On("Get").Return(nil)
			renderer.On("Run", filesys.MakeFsOnDisk(), filepath.Join(destination, path)).Return(getManifestsResponseMap(), nil)
			mockCache.On("Get", source).Return(etag, true)
			mockHttp.On("Do", mock.Anything).Return(GetHTTPResponseWithStatusNotModified(), nil)
			mockCache.On("Get", etag).Times(1).Return(getManifestsUnstructured(), true)
			k := kustomize.New(renderer, getter, destination, source, path, cacheWrapper)
			manifests, renderError := k.Render()
			assert.Equal(GinkgoT(), renderError, nil)
			assert.Equal(GinkgoT(), len(manifests), 2)
			assert.Equal(GinkgoT(), renderError, nil)
			assert.Equal(GinkgoT(), len(manifests), 2)
			assert.Equal(GinkgoT(), manifests[0].GetName(), "deploy1")
			assert.Equal(GinkgoT(), manifests[0].GetKind(), "Deployment")
			assert.Equal(GinkgoT(), manifests[0].GetAPIVersion(), "apps/v1")
			assert.Equal(GinkgoT(), manifests[1].GetAPIVersion(), "apps/v1")
			assert.Equal(GinkgoT(), manifests[1].GetKind(), "Deployment")
			assert.Equal(GinkgoT(), manifests[1].GetName(), "deploy2")

		})
	})

})

func getManifestsResponseMap() resmap.ResMap {
	var depProvider = provider.NewDefaultDepProvider()
	var rf = depProvider.GetResourceFactory()
	resourceStr := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
---
`
	fSys := filesys.MakeFsInMemory()
	err := fSys.WriteFile("deployment.yaml", []byte(resourceStr))
	assert.NoError(GinkgoT(), err)
	ldr, err := loader.NewLoader(
		loader.RestrictionRootOnly, filesys.Separator, fSys)
	assert.NoError(GinkgoT(), err)
	var rmF = resmap.NewFactory(rf)
	resmap, err := rmF.FromFile(ldr, "deployment.yaml")
	assert.NoError(GinkgoT(), err)
	return resmap
}

func GetHTTPResponseWithStatusNotModified() *http.Response {
	response := new(http.Response)
	response.Header = make(map[string][]string)
	response.Header.Set("etag", "123")
	response.StatusCode = http.StatusNotModified
	return response
}

func GetHTTPResponseWithStatusModified() *http.Response {
	response := new(http.Response)
	response.Header = make(map[string][]string)
	response.Header.Set("etag", "123")
	response.StatusCode = http.StatusOK
	return response
}
func getManifestsUnstructured() []unstructured.Unstructured {
	var unstructuredManifest []unstructured.Unstructured
	resMap := getManifestsResponseMap()
	resMapBytes, err := json.Marshal(resMap.Resources())
	assert.NoError(GinkgoT(), err)
	err = json.Unmarshal(resMapBytes, &unstructuredManifest)
	assert.NoError(GinkgoT(), err)
	return unstructuredManifest
}
