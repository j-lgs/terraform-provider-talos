package datatypes

import (
	"github.com/dustin/go-humanize"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

func (partition PartitionData) Data() (any, error) {
	size, err := humanize.ParseBytes(partition.Size.Value)
	if err != nil {
		return nil, err
	}

	part := &v1alpha1.DiskPartition{
		DiskSize:       v1alpha1.DiskSize(size),
		DiskMountPoint: partition.MountPoint.Value,
	}

	return part, nil
}

func (partition *PartitionData) Read(part any) error {
	diskPartition := part.(*v1alpha1.DiskPartition)

	size, err := diskPartition.DiskSize.MarshalYAML()
	if err != nil {
		return err
	}
	partition.Size = types.String{Value: size.(string)}

	return nil
}

type Partitions = []*v1alpha1.DiskPartition
type TalosPartitions struct {
	*Partitions
}

func (diskData MachineDiskData) Data() (any, error) {
	disk := &v1alpha1.MachineDisk{
		DeviceName: diskData.DeviceName.Value,
	}

	for _, partition := range diskData.Partitions {
		part, err := partition.Data()
		if err != nil {
			return nil, err
		}
		disk.DiskPartitions = append(disk.DiskPartitions, part.(*v1alpha1.DiskPartition))
	}

	return disk, nil
}

func (data *MachineDiskData) Read(diskData any) error {
	machineDisk := diskData.(*v1alpha1.MachineDisk)

	data.DeviceName.Value = machineDisk.DeviceName
	for _, partition := range machineDisk.DiskPartitions {
		part := PartitionData{}
		err := part.Read(partition)
		if err != nil {
			return err
		}
		data.Partitions = append(data.Partitions, part)
	}

	return nil
}

func (diskData MachineDiskDataList) GenOpts() (out []generate.GenOption, err error) {
	userDisks := []*v1alpha1.MachineDisk{}
	for _, disk := range diskData {
		userDisk, err := disk.Data()
		if err != nil {
			return nil, err
		}
		userDisks = append(userDisks, userDisk.(*v1alpha1.MachineDisk))
	}

	out = append(out, generate.WithUserDisks(userDisks))

	return
}

type MachineDisks = []*v1alpha1.MachineDisk
type TalosMachineDisk struct {
	MachineDisks
}

func (talosMachineDisk TalosMachineDisk) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.Disks == nil {
				planConfig.Disks = make([]MachineDiskData, 0)
			}

			for _, talosDisk := range talosMachineDisk.MachineDisks {
				disk := MachineDiskData{
					DeviceName: readString(talosDisk.DeviceName),
				}

				for _, talosPart := range talosDisk.DiskPartitions {
					disk.Partitions = append(disk.Partitions, PartitionData{
						Size:       readString(humanize.Bytes(talosPart.Size())),
						MountPoint: readString(talosPart.MountPoint()),
					})
				}

				planConfig.Disks = append(planConfig.Disks, disk)
			}

			return nil
		},
	}

	return funs
}
