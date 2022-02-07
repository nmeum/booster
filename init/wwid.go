package main

import (
	"path"
	"strings"
)

// wwid computes systemd-style WWID for block devices
// the function tries to follow the code at https://github.com/systemd/systemd/blob/main/rules.d/60-persistent-storage.rules
func wwid(device string) ([]string, error) {
	ids := []string{}

	name := path.Base(device)

	if strings.HasPrefix(name, "nvme") {
		// wwid attribute
		id, err := sysfsAttributeValue(device, "wwid")
		if err != nil {
			return nil, err
		}

		ids = append(ids, "nvme-"+id)

		// serial
		model, err := sysfsAttributeValue(device, "model")
		if err != nil {
			return nil, err
		}
		model = strings.ReplaceAll(strings.TrimSpace(model), " ", "_")
		serial, err := sysfsAttributeValue(device, "serial")
		if err != nil {
			return nil, err
		}
		serial = strings.ReplaceAll(strings.TrimSpace(serial), " ", "_")
		ids = append(ids, "nvme-"+model+"_"+serial)
	} else if strings.HasPrefix(name, "dm-") {
		name, err := sysfsAttributeValue(device+"/dm", "name")
		if err != nil {
			return nil, err
		}
		ids = append(ids, "dm-name-"+name)

		uuid, err := sysfsAttributeValue(device+"/dm", "uuid")
		if err != nil {
			return nil, err
		}
		ids = append(ids, "dm-uuid-"+uuid)
	} else if strings.HasPrefix(name, "sd") || strings.HasPrefix(name, "sr") {
		subsystems, err := sysfsSubsystems(device)
		if err != nil {
			return nil, err
		}

		// serial
		vendor, err := sysfsAttributeValue(device, "vendor")
		if err != nil {
			return nil, err
		}
		vendor = strings.ReplaceAll(vendor, " ", "_")
		model, err := sysfsAttributeValue(device, "model")
		if err != nil {
			return nil, err
		}
		model = strings.ReplaceAll(model, " ", "_")
		serial, err := sysfsAttributeValue(device, "serial")
		if err != nil {
			return nil, err
		}

		bus := "scsi"
		if subsystems["usb"] {
			bus = "usb"
		}
		target := "0"
		lun := "0"
		instanceID := target + ":" + lun

		ids = append(ids, bus+"-"+vendor+"_"+model+"_"+serial+"-"+instanceID)

		// wwid
		wwid, err := sysfsAttributeValue(device, "wwid")
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(wwid, "naa.") {
			wwid = "0x" + wwid[4:]
		}

		ids = append(ids, "wwn-"+wwid)
	}

	return ids, nil
}
