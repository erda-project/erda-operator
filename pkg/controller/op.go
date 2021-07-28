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

package controller

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/dice-operator/pkg/cluster"
	"github.com/erda-project/dice-operator/pkg/spec"
)

func (c *Controller) onAdd(obj interface{}) {
	spec := obj.(*spec.DiceCluster)
	clus, err := cluster.New(spec, c.client, c.k8sclient)
	if err != nil {
		logrus.Errorf("Failed to create cluster: %v", err)
		return
	}
	c.diceClusters[spec.Name] = clus
}

func (c *Controller) onUpdate(obj interface{}) {
	spec := obj.(*spec.DiceCluster)
	dc, ok := c.diceClusters[spec.Name]
	if !ok {
		logrus.Errorf("Failed to update, cluster not exists: %v", spec.Name)
		return
	}
	dc.Update(spec)
}

func (c *Controller) onDelete(obj interface{}) {
	spec := obj.(*spec.DiceCluster)
	dc, ok := c.diceClusters[spec.Name]
	if !ok {
		logrus.Errorf("Failed to delete, cluster not exists: %v", spec.Name)
		return
	}
	dc.Delete()
}
