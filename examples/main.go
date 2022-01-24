package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ZupIT/charlescd-kustomize/cache"
	"github.com/ZupIT/charlescd-kustomize/kustomize"
	"github.com/dgraph-io/ristretto"
	"github.com/dgraph-io/ristretto/z"
	"github.com/hashicorp/go-getter"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/build"
	"strconv"
)

func main() {
	cacheClient, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
	wrapper := cache.New(cacheClient, &http.Client{})
	kustomizer := krusty.MakeKustomizer(
		build.HonorKustomizeFlags(krusty.MakeDefaultOptions()))
	pwd, err := os.Getwd()
	client := getter.Client{
		Pwd:  pwd,
		Ctx:  context.TODO(),
		Mode: getter.ClientModeAny,
		Src:  "github.com/thallesfreitaszup/kustomize-demo",
		Dst:  filepath.Join(os.TempDir(), "kustomize"+strconv.Itoa(int(z.FastRand()))),
	}
	path := "overlays/dev"
	k := kustomize.New(kustomizer, &client, client.Dst, client.Src, path, wrapper)
	manifests, err := k.Render()
	if err != nil {
		panic(err)
	}
	bytes, err := json.Marshal(manifests)
	fmt.Println(string(bytes))

	manifests, err = k.Render()
	if err != nil {
		panic(err)
	}
	bytes, err = json.Marshal(manifests)
	fmt.Println(string(bytes))
}
