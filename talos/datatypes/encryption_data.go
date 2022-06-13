package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

func (encryptionData EncryptionData) Data() (any, error) {
	encryption := &v1alpha1.SystemDiskEncryptionConfig{}

	if encryptionData.State != nil {
		state, err := encryptionData.State.Data()
		if err != nil {
			return nil, err
		}
		encryption.StatePartition = state.(*v1alpha1.EncryptionConfig)
	}

	if encryptionData.Ephemeral != nil {
		ephemeral, err := encryptionData.Ephemeral.Data()
		if err != nil {
			return nil, err
		}
		encryption.EphemeralPartition = ephemeral.(*v1alpha1.EncryptionConfig)
	}

	return encryption, nil
}

func (encryptionData EncryptionData) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			enc, err := encryptionData.Data()
			if err != nil {
				return err
			}
			cfg.MachineConfig.MachineSystemDiskEncryption = enc.(*v1alpha1.SystemDiskEncryptionConfig)
			return nil
		},
	}
}

func (encryptionData EncryptionData) GenOpts() (out []generate.GenOption, err error) {
	systemEncryption, err := encryptionData.Data()
	if err != nil {
		return nil, err
	}
	out = append(out, generate.WithSystemDiskEncryption(systemEncryption.(*v1alpha1.SystemDiskEncryptionConfig)))

	return
}

func (data *EncryptionData) Read(diskData any) error {
	encryptionConfig := diskData.(*v1alpha1.SystemDiskEncryptionConfig)

	if encryptionConfig.StatePartition != nil {
		data.State = &EncryptionConfigData{}
		err := data.State.Read(encryptionConfig.StatePartition)
		if err != nil {
			return err
		}
	}

	if encryptionConfig.EphemeralPartition != nil {
		ephemeralData := &EncryptionConfigData{}
		err := ephemeralData.Read(encryptionConfig.EphemeralPartition)
		if err != nil {
			return err
		}
		data.Ephemeral = ephemeralData
	}

	return nil
}

type TalosSystemDiskEncryptionConfig struct {
	*v1alpha1.SystemDiskEncryptionConfig
}

func (encryptionData EncryptionConfigData) Data() (any, error) {
	encryptionConfig := &v1alpha1.EncryptionConfig{
		EncryptionProvider: encryptionData.Provider.Value,
	}

	for _, key := range encryptionData.Keys {
		k, err := key.Data()
		if err != nil {
			return nil, err
		}
		encryptionConfig.EncryptionKeys = append(encryptionConfig.EncryptionKeys, k.(*v1alpha1.EncryptionKey))
	}

	if !encryptionData.Cipher.Null {
		encryptionConfig.EncryptionCipher = encryptionData.Cipher.Value
	}

	if !encryptionData.KeySize.Null {
		encryptionConfig.EncryptionKeySize = uint(encryptionData.KeySize.Value)
	}

	if !encryptionData.BlockSize.Null {
		encryptionConfig.EncryptionBlockSize = uint64(encryptionData.BlockSize.Value)
	}

	for _, opt := range encryptionData.PerfOptions {
		encryptionConfig.EncryptionPerfOptions = append(encryptionConfig.EncryptionPerfOptions, opt.Value)
	}

	return encryptionConfig, nil
}

func (data *EncryptionConfigData) Read(encryptionData any) error {
	partEncryptionConfig := encryptionData.(*v1alpha1.EncryptionConfig)

	data.Provider.Value = partEncryptionConfig.EncryptionProvider

	for _, key := range partEncryptionConfig.EncryptionKeys {
		keyconfig := KeyConfig{}
		err := keyconfig.Read(key)
		if err != nil {
			return err
		}
		data.Keys = append(data.Keys, keyconfig)
	}

	if partEncryptionConfig.EncryptionCipher != *new(string) {
		data.Cipher.Value = partEncryptionConfig.EncryptionCipher
	}

	if partEncryptionConfig.EncryptionKeySize != *new(uint) {
		data.KeySize.Value = int64(partEncryptionConfig.EncryptionKeySize)
	}

	if partEncryptionConfig.EncryptionBlockSize != *new(uint64) {
		data.BlockSize.Value = int64(partEncryptionConfig.EncryptionBlockSize)
	}

	for _, opt := range partEncryptionConfig.EncryptionPerfOptions {
		data.PerfOptions = append(data.PerfOptions, types.String{Value: opt})
	}

	return nil
}

type TalosEncryptionConfig struct {
	*v1alpha1.EncryptionConfig
}

func (keyData KeyConfig) Data() (any, error) {
	encryptionKey := &v1alpha1.EncryptionKey{
		KeySlot: int(keyData.Slot.Value),
	}

	if !keyData.KeyStatic.Null {
		encryptionKey.KeyStatic = &v1alpha1.EncryptionKeyStatic{
			KeyData: keyData.KeyStatic.Value,
		}
	}

	if !keyData.NodeID.Null {
		encryptionKey.KeyNodeID = &v1alpha1.EncryptionKeyNodeID{}
	}

	return encryptionKey, nil
}

func (data *KeyConfig) Read(keyData any) error {
	key := keyData.(*v1alpha1.EncryptionKey)

	data.Slot.Value = int64(key.KeySlot)

	if key.KeySlot != *new(int) {
		data.Slot.Value = int64(key.KeySlot)
	}

	if key.KeyNodeID != nil {
		data.NodeID.Value = true
		data.KeyStatic.Null = true
	}

	if key.KeyStatic != nil {
		data.KeyStatic.Value = key.KeyStatic.KeyData
		data.NodeID.Null = true
	}

	return nil
}

type TalosEncryptionKey struct {
	*v1alpha1.EncryptionKey
}
