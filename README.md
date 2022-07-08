# The Charles project has been archived by Zup Innovation. It might start again eventually; however, we won't deliver support for now.

## Description 
A kustomize lib that downloads the contents from the source, do the `kustomize build` of kustomize 
and return an array of manifests\
## How to use
Instantiate the kustomizer and client of your preference(we recommend the go getter client) 
and pass as an argument of the **New** method with the source url, the destination and path of manifests.
See the example below:

    
        kustomizer := krusty.MakeKustomizer(
		build.HonorKustomizeFlags(krusty.MakeDefaultOptions()),)
        client := getter.Client{
            Pwd:  pwd,
            Ctx:  context.TODO(),
            Mode: getter.ClientModeAny,
            Src:  "github.com/thallesfreitaszup/kustomize-demo",
            Dst:  filepath.Join(os.TempDir(), "kustomize"+strconv.Itoa(int(z.FastRand()))),
	    }
        path := "overlays/dev"
	    k := kustomize.New(kustomizer, &client, client.Dst, client.Src, path)


    
After that call the **Render** method that will return the desired manifests:

    ```
        manifests, err := k.Render()

    ``

