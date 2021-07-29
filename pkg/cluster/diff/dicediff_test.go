// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package diff

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func Test_SpecDiff(t *testing.T) {
	current := spec.DiceCluster{
		Spec: spec.ClusterSpec{
			Dice: diceyml.Object{
				Envs: map[string]string{"aaa": "aaa"},
				Services: map[string]*diceyml.Service{
					"eventbox": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
				},
			},
		},
	}
	target := spec.DiceCluster{
		Spec: spec.ClusterSpec{
			Dice: diceyml.Object{
				Envs: map[string]string{"aaa": "aaa"},
				Services: map[string]*diceyml.Service{
					"eventbox": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
				},
			},
		},
	}
	diff := cmp.Diff(current, target)
	t.Logf("diff: %v", diff)
	result := cmp.Equal(&current, &target)
	assert.Equal(t, true, result)

}

func Test_SpecDiffReplica(t *testing.T) {
	current := spec.DiceCluster{
		Spec: spec.ClusterSpec{
			Dice: diceyml.Object{
				Envs: map[string]string{"aaa": "aaa"},
				Services: map[string]*diceyml.Service{
					"eventbox": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
				},
			},
		},
	}
	target := spec.DiceCluster{
		Spec: spec.ClusterSpec{
			Dice: diceyml.Object{
				Envs: map[string]string{"aaa": "bbb"},
				Services: map[string]*diceyml.Service{
					"eventbox": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 2,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
				},
			},
		},
	}
	d := NewSpecDiff(&current, &target)
	a := d.GetActions()
	fmt.Printf("%+v\n", target.Spec.Dice.Services["eventbox"]) // debug print

	assert.True(t, len(a.UpdatedServices) == 2, "%v", a)
	assert.True(t, a.UpdatedServices["eventbox"].Envs["aaa"] == "bbb")
	assert.True(t, a.UpdatedServices["eventbox2"].Envs["aaa"] == "bbb")

}

func Test_SpecDiff1(t *testing.T) {
	current := spec.DiceCluster{
		Spec: spec.ClusterSpec{
			Dice: diceyml.Object{
				Envs: map[string]string{"aaa": "aaa"},
				Services: map[string]*diceyml.Service{
					"eventbox": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
					"eventbox2": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
				},
			},
		},
	}
	target := spec.DiceCluster{
		Spec: spec.ClusterSpec{
			Dice: diceyml.Object{
				Envs: map[string]string{"aaa": "bbb"},
				Services: map[string]*diceyml.Service{
					"eventbox": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
					"eventbox2": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
				},
			},
		},
	}
	d := NewSpecDiff(&current, &target)
	a := d.GetActions()
	fmt.Printf("%+v\n", target.Spec.Dice.Services["eventbox"]) // debug print

	assert.True(t, len(a.UpdatedServices) == 2, "%v", a)
	assert.True(t, a.UpdatedServices["eventbox"].Envs["aaa"] == "bbb")
	assert.True(t, a.UpdatedServices["eventbox2"].Envs["aaa"] == "bbb")

}

func Test_SpecDiff2(t *testing.T) {
	current := spec.DiceCluster{
		Spec: spec.ClusterSpec{
			Dice: diceyml.Object{
				Services: map[string]*diceyml.Service{
					"eventbox": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
					"eventbox2": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV1"},
					},
				},
			},
		},
	}
	target := spec.DiceCluster{
		Spec: spec.ClusterSpec{
			Dice: diceyml.Object{
				Services: map[string]*diceyml.Service{
					"eventbox2": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV2"},
					},
				},
			},
		},
	}
	d := NewSpecDiff(&current, &target)
	a := d.GetActions()
	assert.True(t, len(a.UpdatedServices) == 1, "%v", a)
	assert.True(t, len(a.DeletedServices) == 1, "%v", a)
}
func Test_SpecDiff3(t *testing.T) {

	target := spec.DiceCluster{
		Spec: spec.ClusterSpec{
			Dice: diceyml.Object{
				Services: map[string]*diceyml.Service{
					"eventbox2": {
						Image: "xxx",
						Resources: diceyml.Resources{
							CPU: 0.05,
							Mem: 512,
						},
						Deployments: diceyml.Deployments{
							Replicas: 1,
						},
						Ports: []diceyml.ServicePort{{Port: 9528}},
						Envs:  map[string]string{"ENV1": "ENV2"},
					},
				},
			},
		},
	}
	d := NewSpecDiff(nil, &target)
	a := d.GetActions()
	assert.True(t, len(a.AddedServices) == 1, "%v", a)
}

func Test_isServiceEqual(t *testing.T) {
	obj1 := diceyml.Service{
		Image: "xxx",
		Resources: diceyml.Resources{
			CPU: 0.05,
			Mem: 512,
		},
		Deployments: diceyml.Deployments{
			Replicas: 1,
		},
		Ports: []diceyml.ServicePort{{Port: 9528}},
		Envs:  map[string]string{"ENV1": "xxxxxx"},
	}
	obj2 := diceyml.Service{
		Image: "xxx",
		Resources: diceyml.Resources{
			CPU: 0.05,
			Mem: 512,
		},
		Deployments: diceyml.Deployments{
			Replicas: 1,
		},
		Ports: []diceyml.ServicePort{{Port: 9528}},
		Envs:  map[string]string{"ENV1": "yyyyyy"},
	}
	assert.False(t, isServiceEqual(&obj1, &obj2))
}
func Test_isServiceEqual2(t *testing.T) {
	obj1 := diceyml.Service{
		Image: "xxx",
		Resources: diceyml.Resources{
			CPU: 0.05,
			Mem: 512,
		},
		Deployments: diceyml.Deployments{
			Replicas: 1,
		},
		Ports: []diceyml.ServicePort{{Port: 9528}},
		Envs:  map[string]string{"ENV1": "xxxxxx"},
	}
	obj2 := diceyml.Service{
		Image: "xxx",
		Resources: diceyml.Resources{
			CPU: 0.05,
			Mem: 512,
		},
		Deployments: diceyml.Deployments{
			Replicas: 1,
		},
		Ports: []diceyml.ServicePort{{Port: 9528}},
		Envs:  map[string]string{"ENV1": "xxxxxx"},
	}
	assert.True(t, isServiceEqual(&obj1, &obj2))
}
