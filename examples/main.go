package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ZupIT/charlescd-kustomize/kustomize"
	"github.com/dgraph-io/ristretto/z"
	"github.com/hashicorp/go-getter"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/build"
	"strconv"
)

func main() {
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
	k := kustomize.New(kustomizer, &client, client.Dst, client.Src, path, kustomize.Options{})
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
