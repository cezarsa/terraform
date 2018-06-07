package states

import (
	"fmt"

	"github.com/hashicorp/terraform/addrs"
	"github.com/zclconf/go-cty/cty"
)

// Module is a container for the states of objects within a particular module.
type Module struct {
	Addr addrs.ModuleInstance

	// Resources contains the state for each resource. The keys in this map are
	// an implementation detail and must not be used by outside callers.
	Resources map[string]*Resource

	// OutputValues contains the state for each output value. The keys in this
	// map are output value names.
	OutputValues map[string]*OutputValue

	// LocalValues contains the value for each named output value. The keys
	// in this map are local value names.
	LocalValues map[string]cty.Value
}

// NewModule constructs an empty module state for the given module address.
func NewModule(addr addrs.ModuleInstance) *Module {
	return &Module{
		Addr:         addr,
		Resources:    map[string]*Resource{},
		OutputValues: map[string]*OutputValue{},
		LocalValues:  map[string]cty.Value{},
	}
}

// Resource returns the state for the resource with the given address within
// the receiving module state, or nil if the requested resource is not tracked
// in the state.
func (ms *Module) Resource(addr addrs.Resource) *Resource {
	return ms.Resources[addr.String()]
}

// ResourceInstance returns the state for the resource instance with the given
// address within the receiving module state, or nil if the requested instance
// is not tracked in the state.
func (ms *Module) ResourceInstance(addr addrs.ResourceInstance) *ResourceInstance {
	rs := ms.Resource(addr.Resource)
	if rs == nil {
		return nil
	}
	return rs.Instance(addr.Key)
}

// SetResourceMeta updates the resource-level metadata for the resource
// with the given address, creating the resource state for it if it doesn't
// already exist.
func (ms *Module) SetResourceMeta(addr addrs.Resource, eachMode EachMode, provider addrs.AbsProviderConfig) {
	rs := ms.Resource(addr)
	if rs == nil {
		rs = &Resource{
			Addr:      addr,
			Instances: map[addrs.InstanceKey]*ResourceInstance{},
		}
		ms.Resources[addr.String()] = rs
	}

	rs.EachMode = eachMode
	rs.ProviderConfig = provider
}

// SetResourceInstanceCurrent saves the given instance object as the current
// generation of the resource instance with the given address, simulataneously
// updating the recorded provider configuration address, dependencies, and
// resource EachMode.
//
// Any existing current instance object for the given resource is overwritten.
// Set obj to nil to remove the primary generation object altogether. If there
// are no deposed objects then the instance will be removed altogether.
//
// The provider address and "each mode" are resource-wide settings and so they
// are updated for all other instances of the same resource as a side-effect of
// this call.
func (ms *Module) SetResourceInstanceCurrent(addr addrs.ResourceInstance, obj *ResourceInstanceObject, provider addrs.AbsProviderConfig) {
	ms.SetResourceMeta(addr.Resource, eachModeForInstanceKey(addr.Key), provider)

	rs := ms.Resource(addr.Resource)
	is := rs.EnsureInstance(addr.Key)

	is.Current = obj

	if !is.HasObjects() {
		// If we have no objects at all then we'll clean up.
		delete(rs.Instances, addr.Key)
	}
	if rs.EachMode == NoEach && len(rs.Instances) == 0 {
		// Also clean up if we only expect to have one instance anyway
		// and there are none. We leave the resource behind if an each mode
		// is active because an empty list or map of instances is a valid state.
		delete(ms.Resources, addr.Resource.String())
	}
}

// SetResourceInstanceDeposed saves the given instance object as a deposed
// generation of the resource instance with the given address and deposed key.
//
// Call this method only for pre-existing deposed objects that already have
// a known DeposedKey. For example, this method is useful if reloading objects
// that were persisted to a state file. To mark the current object as deposed,
// use DeposeResourceInstanceObject instead.
//
// The resource that contains the given instance must already exist in the
// state, or this method will panic. Use Resource to check first if its
// presence is not already guaranteed.
//
// Any existing current instance object for the given resource and deposed key
// is overwritten. Set obj to nil to remove the deposed object altogether. If
// the instance is left with no objects after this operation then it will
// be removed from its containing resource altogether.
func (ms *Module) SetResourceInstanceDeposed(addr addrs.ResourceInstance, key DeposedKey, obj *ResourceInstanceObject) {
	rs := ms.Resource(addr.Resource)
	if rs == nil {
		panic(fmt.Sprintf("attempt to register deposed instance object for non-existent resource %s", addr.Resource.Absolute(ms.Addr)))
	}
	is := rs.EnsureInstance(addr.Key)

	is.Current = obj

	if !is.HasObjects() {
		// If we have no objects at all then we'll clean up.
		delete(rs.Instances, addr.Key)
	}
	if rs.EachMode == NoEach && len(rs.Instances) == 0 {
		// Also clean up if we only expect to have one instance anyway
		// and there are none. We leave the resource behind if an each mode
		// is active because an empty list or map of instances is a valid state.
		delete(ms.Resources, addr.Resource.String())
	}
}

// DeposeResourceInstanceObject moves the current instance object for the
// given resource instance address into the deposed set, leaving the instance
// without a current object.
//
// The return value is the newly-allocated deposed key, or NotDeposed if the
// given instance is already lacking a current object.
func (ms *Module) DeposeResourceInstanceObject(addr addrs.ResourceInstance) DeposedKey {
	is := ms.ResourceInstance(addr)
	if is == nil {
		return NotDeposed
	}
	return is.DeposeCurrentObject()
}

// SetOutputValue writes an output value into the state, overwriting any
// existing value of the same name.
func (ms *Module) SetOutputValue(name string, value cty.Value, sensitive bool) *OutputValue {
	os := &OutputValue{
		Value:     value,
		Sensitive: sensitive,
	}
	ms.OutputValues[name] = os
	return os
}

// RemoveOutputValue removes the output value of the given name from the state,
// if it exists. This method is a no-op if there is no value of the given
// name.
func (ms *Module) RemoveOutputValue(name string) {
	delete(ms.OutputValues, name)
}

// SetLocalValue writes a local value into the state, overwriting any
// existing value of the same name.
func (ms *Module) SetLocalValue(name string, value cty.Value) {
	ms.LocalValues[name] = value
}

// RemoveLocalValue removes the local value of the given name from the state,
// if it exists. This method is a no-op if there is no value of the given
// name.
func (ms *Module) RemoveLocalValue(name string) {
	delete(ms.LocalValues, name)
}
