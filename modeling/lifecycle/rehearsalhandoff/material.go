package rehearsalhandoff

import "github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"

// InspectMaterial verifies the existing carrier and returns owned payload bytes.
// Reports remain opaque under this operation, exactly as in Verify. Callers must
// inspect their meaning separately; neither operation authenticates execution.
func InspectMaterial(raw []byte, refs References) (Material, Receipt, error) {
	receipt, err := Verify(raw, refs)
	if err != nil {
		return Material{}, Receipt{}, err
	}
	var d bundleDTO
	if strictjson.Decode(raw, &d, MaxBundleBytes, MaxDepth) != nil {
		return Material{}, Receipt{}, ErrInvalid
	}
	material, err := decodePayload(d.Payload)
	if err != nil {
		return Material{}, Receipt{}, err
	}
	return material, receipt, nil
}
