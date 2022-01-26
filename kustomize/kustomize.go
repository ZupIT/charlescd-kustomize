/*
 * Copyright 2022 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kustomize

import (
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type KustomizerWrapper struct {
	FSys        filesys.FileSystem
	Renderer    Renderer
	Client      Getter
	Destination string
	Source      string
	Path        string
	Options     Options
}

type Renderer interface {
	Run(fSys filesys.FileSystem, path string) (resmap.ResMap, error)
}

type Getter interface {
	Get() error
}

type Cache interface {
	GetManifests(source string) ([]unstructured.Unstructured, error)
	Add(key, value interface{}) error
}
type Options struct {
	Cache Cache
}

// New Instantiate a new Wrapper of Kustomize that will do the `kustomize build` of the source
func New(kustomizer Renderer, client Getter, destination, source, path string, Options Options) KustomizerWrapper {
	fsys := filesys.MakeFsOnDisk()

	return KustomizerWrapper{Renderer: kustomizer, FSys: fsys, Client: client, Destination: destination, Source: source, Path: path, Options: Options}
}

// Render downloads the content of the source url and calls the kustomizer run to do the build of
// manifests stored on source
func (k KustomizerWrapper) Render() ([]unstructured.Unstructured, error) {
	var unstructuredManifests []unstructured.Unstructured
	var manifests, err = k.getCachedManifests()
	if err == nil {
		return manifests, nil
	}
	err = k.getSourceContent()
	if err != nil {
		return unstructuredManifests, err
	}

	resMap, err := k.Renderer.Run(k.FSys, filepath.Join(k.Destination, k.Path))
	if err != nil {
		return unstructuredManifests, err
	}
	resources, err := json.Marshal(resMap.Resources())
	if err != nil {
		return unstructuredManifests, fmt.Errorf("error marshalling kustomize resources: %w", err)
	}
	err = json.Unmarshal(resources, &unstructuredManifests)
	if err != nil {
		return unstructuredManifests, fmt.Errorf("error converting kustomize resources to unstructured manifests %w", err)
	}
	err = k.cacheManifests(unstructuredManifests)
	if err != nil {
		return nil, err
	}
	return unstructuredManifests, nil
}

func (k KustomizerWrapper) getSourceContent() error {

	if err := k.Client.Get(); err != nil {
		return err
	}
	return nil
}

func (k KustomizerWrapper) getCachedManifests() ([]unstructured.Unstructured, error) {
	if k.Options.Cache != nil {
		return k.Options.Cache.GetManifests(k.Source)
	}
	return nil, errors.New("cache option is not defined")
}

func (k KustomizerWrapper) cacheManifests(manifests []unstructured.Unstructured) error {
	if k.Options.Cache != nil {
		return k.Options.Cache.Add(k.Source, manifests)
	}
	return nil
}
