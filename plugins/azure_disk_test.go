/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugins

import (
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsManagedDisk(t *testing.T) {
	tests := []struct {
		options  string
		expected bool
	}{
		{
			options:  "testurl/subscriptions/12/resourceGroups/23/providers/Microsoft.Compute/disks/name",
			expected: true,
		},
		{
			options:  "test.com",
			expected: true,
		},
		{
			options:  "HTTP://test.com",
			expected: false,
		},
		{
			options:  "http://test.com/vhds/name",
			expected: false,
		},
	}

	for _, test := range tests {
		result := isManagedDisk(test.options)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("input: %q, isManagedDisk result: %t, expected: %t", test.options, result, test.expected)
		}
	}
}

func TestGetDiskName(t *testing.T) {
	mDiskPathRE := managedDiskPathRE
	uDiskPathRE := unmanagedDiskPathRE
	tests := []struct {
		options   string
		expected1 string
		expected2 error
	}{
		{
			options:   "testurl/subscriptions/12/resourceGroups/23/providers/Microsoft.Compute/disks/name",
			expected1: "name",
			expected2: nil,
		},
		{
			options:   "testurl/subscriptions/23/providers/Microsoft.Compute/disks/name",
			expected1: "",
			expected2: fmt.Errorf("could not get disk name from testurl/subscriptions/23/providers/Microsoft.Compute/disks/name, correct format: %s", mDiskPathRE),
		},
		{
			options:   "http://test.com/vhds/name",
			expected1: "name",
			expected2: nil,
		},
		{
			options:   "http://test.io/name",
			expected1: "",
			expected2: fmt.Errorf("could not get disk name from http://test.io/name, correct format: %s", uDiskPathRE),
		},
	}

	for _, test := range tests {
		result1, result2 := getDiskName(test.options)
		if !reflect.DeepEqual(result1, test.expected1) || !reflect.DeepEqual(result2, test.expected2) {
			t.Errorf("input: %q, getDiskName result1: %q, expected1: %q, result2: %q, expected2: %q", test.options, result1, test.expected1,
				result2, test.expected2)
		}
	}
}

func TestTranslateAzureDiskInTreeStorageClassToCSI(t *testing.T) {
	translator := NewAzureDiskCSITranslator()

	cases := []struct {
		name   string
		volume *corev1.Volume
		expVol *corev1.PersistentVolume
		expErr bool
	}{
		{
			name:   "empty volume",
			expErr: true,
		},
		{
			name:   "no azure disk volume",
			volume: &corev1.Volume{},
			expErr: true,
		},
		{
			name: "azure disk volume",
			volume: &corev1.Volume{
				VolumeSource: corev1.VolumeSource{
					AzureDisk: &corev1.AzureDiskVolumeSource{
						DiskName:    "diskname",
						DataDiskURI: "datadiskuri",
					},
				},
			},
			expVol: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "disk.csi.azure.com-diskname",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver:           "disk.csi.azure.com",
							VolumeHandle:     "datadiskuri",
							VolumeAttributes: map[string]string{azureDiskKind: "Managed"},
						},
					},
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Logf("Testing %v", tc.name)
		got, err := translator.TranslateInTreeInlineVolumeToCSI(tc.volume)
		if err != nil && !tc.expErr {
			t.Errorf("Did not expect error but got: %v", err)
		}

		if err == nil && tc.expErr {
			t.Errorf("Expected error, but did not get one.")
		}

		if !reflect.DeepEqual(got, tc.expVol) {
			t.Errorf("Got parameters: %v, expected :%v", got, tc.expVol)
		}
	}
}

func TestTranslateAzureDiskInTreePVToCSI(t *testing.T) {
	translator := NewAzureDiskCSITranslator()

	cachingMode := corev1.AzureDataDiskCachingMode("cachingmode")
	fsType := "fstype"
	readOnly := true

	cases := []struct {
		name   string
		volume *corev1.PersistentVolume
		expVol *corev1.PersistentVolume
		expErr bool
	}{
		{
			name:   "empty volume",
			expErr: true,
		},
		{
			name:   "no azure file volume",
			volume: &corev1.PersistentVolume{},
			expErr: true,
		},
		{
			name: "azure file volume",
			volume: &corev1.PersistentVolume{
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						AzureDisk: &corev1.AzureDiskVolumeSource{
							CachingMode: &cachingMode,
							DataDiskURI: "datadiskuri",
							FSType:      &fsType,
							ReadOnly:    &readOnly,
						},
					},
				},
			},
			expVol: &corev1.PersistentVolume{
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver:   "disk.csi.azure.com",
							FSType:   "fstype",
							ReadOnly: true,
							VolumeAttributes: map[string]string{
								azureDiskCachingMode: "cachingmode",
								azureDiskFSType:      fsType,
								azureDiskKind:        "Managed",
							},
							VolumeHandle: "datadiskuri",
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Logf("Testing %v", tc.name)
		got, err := translator.TranslateInTreePVToCSI(tc.volume)
		if err != nil && !tc.expErr {
			t.Errorf("Did not expect error but got: %v", err)
		}

		if err == nil && tc.expErr {
			t.Errorf("Expected error, but did not get one.")
		}

		if !reflect.DeepEqual(got, tc.expVol) {
			t.Errorf("Got parameters: %v, expected :%v", got, tc.expVol)
		}
	}
}
