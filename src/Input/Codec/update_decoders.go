package Codec

import "fmt"

type UpdateDecoder func(r *Reader, attr byte) (StdinMsg, bool)

var updateDecoders = map[string]UpdateDecoder{}

func RegisterUpdateDecoder(entity string, fn UpdateDecoder) {
	if fn == nil {
		panic("Codec.RegisterUpdateDecoder: nil decoder for " + entity)
	}
	if _, dup := updateDecoders[entity]; dup {
		panic("Codec.RegisterUpdateDecoder: " + entity + " registered twice; two packages claim the same entity")
	}
	updateDecoders[entity] = fn
}

func AssertUpdateDecodersComplete() {
	var missing []string
	for _, entity := range InUpdateKinds {
		if _, ok := updateDecoders[entity]; !ok {
			missing = append(missing, entity)
		}
	}
	if len(missing) > 0 {
		panic(fmt.Sprintf(
			"Codec.AssertUpdateDecodersComplete: no decoder registered for %v. The wire can carry "+
				"an update for each of these (InUpdateKinds, from INPUT_LAYOUT_FINGERPRINT), so every "+
				"edit naming one would decode to nothing and the affordance would look dead. The owning "+
				"package registers in init(); if that package is imported by nothing, its init() never runs.",
			missing))
	}
}
