package driver

import (
	"fmt"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ResourcePool struct {
	pool   *object.ResourcePool
	driver *Driver
}

func (d *Driver) NewResourcePool(ref *types.ManagedObjectReference) *ResourcePool {
	return &ResourcePool{
		pool:   object.NewResourcePool(d.client.Client, *ref),
		driver: d,
	}
}

func (d *Driver) FindResourcePool(name string) (*ResourcePool, error) {
	t := d.client.ServiceContent.About.ApiType
	switch t {
	case "HostAgent":
		ddc, err := d.finder.DefaultDatacenter(d.ctx)
		if err != nil {
			return nil, err
		}
		d.finder.SetDatacenter(ddc)
		p, err := d.finder.DefaultResourcePool(d.ctx)
		if err != nil {
			return nil, err
		}
		return &ResourcePool{
			pool:   p,
			driver: d,
		}, nil
	case "VirtualCenter":
		if name != "" {
			p, err := d.finder.ResourcePool(d.ctx, name)
			if err != nil {
				return nil, err
			}
			return &ResourcePool{
				pool:   p,
				driver: d,
			}, nil
		}
		p, err := d.finder.DefaultResourcePool(d.ctx)
		if err != nil {
			return nil, err
		}
		return &ResourcePool{
			pool:   p,
			driver: d,
		}, nil
	}
	return nil, fmt.Errorf("unsupported ApiType: %s", t)
}

func (p *ResourcePool) Info(params ...string) (*mo.ResourcePool, error) {
	var params2 []string
	if len(params) == 0 {
		params2 = []string{"*"}
	} else {
		params2 = params
	}
	var info mo.ResourcePool
	err := p.pool.Properties(p.driver.ctx, p.pool.Reference(), params2, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (p *ResourcePool) Path() (string, error) {
	poolInfo, err := p.Info("name", "parent")
	if err != nil {
		return "", err
	}
	if poolInfo.Parent.Type == "ComputeResource" {
		return "", nil
	} else {
		parent := p.driver.NewResourcePool(poolInfo.Parent)
		parentPath, err := parent.Path()
		if err != nil {
			return "", err
		}
		if parentPath == "" {
			return poolInfo.Name, nil
		} else {
			return fmt.Sprintf("%v/%v", parentPath, poolInfo.Name), nil
		}
	}
}
