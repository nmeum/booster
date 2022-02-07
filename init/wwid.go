package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/anatol/smart.go"
	"path"
	"strings"
)

// wwid computes systemd-style WWID for block devices
// the function tries to follow the code at https://github.com/systemd/systemd/blob/main/rules.d/60-persistent-storage.rules
func wwid(device string) ([]string, error) {
	ids := []string{}

	name := path.Base(device)

	if strings.HasPrefix(name, "nvme") {
		n, err := smart.OpenNVMe("/dev/" + name)
		if err != nil {
			return nil, err
		}

		c, nss, err := n.Identify()
		if err != nil {
			return nil, err
		}
		wwid := binary.BigEndian.Uint64(nss[0].Eui64[:])
		ids = append(ids, fmt.Sprintf("nvme-eui.%016x", wwid))

		// serial
		model := strings.ReplaceAll(string(bytes.TrimSpace(c.ModelNumber[:])), " ", "_")
		serial := strings.ReplaceAll(string(bytes.TrimSpace(c.SerialNumber[:])), " ", "_")
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
		dev, err := smart.Open("/dev/" + name)
		if err != nil {
			return nil, err
		}

		switch dev.Type() {
		case "sata":
			dev := dev.(*smart.SataDevice)
			i, err := dev.Identify()
			if err != nil {
				return nil, err
			}

			model := strings.ReplaceAll(i.ModelNumber(), " ", "_")
			serial := strings.ReplaceAll(i.SerialNumber(), " ", "_")
			ids = append(ids, "ata-"+model+"_"+serial)

			ids = append(ids, fmt.Sprintf("wwn-%#016x", i.WWN()))
		case "scsi":
			dev := dev.(*smart.ScsiDevice)
			_, err := dev.Inquiry()
			if err != nil {
				return nil, err
			}

			bus := "scsi"
			subsystems, err := sysfsSubsystems(device)
			if err != nil {
				return nil, err
			}
			if subsystems["usb"] {
				bus = "usb"
			}
			target := "0"
			lun := "0"
			instanceID := target + ":" + lun

			_ = bus
			_ = instanceID
		default:
			return nil, fmt.Errorf("unknow S.M.A.R.T. device type: %s", dev.Type())
		}
	}

	return ids, nil
}
