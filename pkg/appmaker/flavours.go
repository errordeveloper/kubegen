package appmaker

const (
	DefaultFlavor         = "default"
	MinimalFlavor         = "minimal"
	ArgsAsConfigMapFlavor = "argsAsConfigMap"
)

var Flavors map[string]GeneralCustomizer

type Flavor struct{}

func init() {
	Flavors = map[string]GeneralCustomizer{
		DefaultFlavor: func(_ *AppComponentResources) {},
		MinimalFlavor: func(i *AppComponentResources) {
			pod := i.getPod()
			if len(pod.Containers) == 0 {
				return // XXX may be we should add error handling for customizers
			}

			if !i.manifest.Opts.WithoutStandardProbes {
				pod.Containers[0].LivenessProbe = nil
				pod.Containers[0].ReadinessProbe = nil
			}
			// TODO: standardSecurityContext, standardTmpVolume
		},
		ArgsAsConfigMapFlavor: func(_ *AppComponentResources) {},
	}
}
