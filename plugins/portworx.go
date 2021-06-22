package plugins

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

var _ InTreePlugin = &portworxCSITranslator{}

// portworxCSITranslator handles translation of PV spec from In-tree
// Portworx File to CSI Portworx File and vice versa
type portworxCSITranslator struct{}

// NewPortworxCSITranslator returns a new instance of portworxCSITranslator
func NewPortworxCSITranslator() InTreePlugin {
	return &portworxCSITranslator{}
}

func (p portworxCSITranslator) TranslateInTreeStorageClassToCSI(sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	return sc, nil
}

func (p portworxCSITranslator) TranslateInTreeInlineVolumeToCSI(volume *v1.Volume, podNamespace string) (*v1.PersistentVolume, error) {
	if volume == nil || volume.PortworxVolume == nil {
		return nil, fmt.Errorf("volume is nil or Portworx Volume not defined on volume")
	}
	// portworxSource := volume.PortworxVolume

	panic("implement me")
}

func (p portworxCSITranslator) TranslateInTreePVToCSI(pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	panic("implement me")
}

func (p portworxCSITranslator) TranslateCSIPVToInTree(pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	panic("implement me")
}

func (p portworxCSITranslator) CanSupport(pv *v1.PersistentVolume) bool {
	panic("implement me")
}

func (p portworxCSITranslator) CanSupportInline(vol *v1.Volume) bool {
	panic("implement me")
}

func (p portworxCSITranslator) GetInTreePluginName() string {
	panic("implement me")
}

func (p portworxCSITranslator) GetCSIPluginName() string {
	panic("implement me")
}

func (p portworxCSITranslator) RepairVolumeHandle(volumeHandle, nodeID string) (string, error) {
	panic("implement me")
}
