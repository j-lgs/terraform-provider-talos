package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planBond BondData) Data() (interface{}, error) {
	bond := &v1alpha1.Bond{}
	for _, netInterface := range planBond.Interfaces {
		bond.BondInterfaces = append(bond.BondInterfaces, netInterface.Value)
	}
	for _, arpIPTarget := range planBond.ARPIPTarget {
		bond.BondARPIPTarget = append(bond.BondARPIPTarget, arpIPTarget.Value)
	}

	b := planBond
	bond.BondMode = b.Mode.Value

	if !b.XmitHashPolicy.Null {
		bond.BondHashPolicy = b.XmitHashPolicy.Value
	}
	if !b.LacpRate.Null {
		bond.BondLACPRate = b.LacpRate.Value
	}
	if !b.AdActorSystem.Null {
		bond.BondADActorSystem = b.AdActorSystem.Value
	}
	if !b.ArpValidate.Null {
		bond.BondARPValidate = b.ArpValidate.Value
	}
	if !b.ArpAllTargets.Null {
		bond.BondARPAllTargets = b.ArpAllTargets.Value
	}
	if !b.Primary.Null {
		bond.BondPrimary = b.Primary.Value
	}
	if !b.PrimaryReselect.Null {
		bond.BondPrimaryReselect = b.PrimaryReselect.Value
	}
	if !b.FailoverMac.Null {
		bond.BondFailOverMac = b.FailoverMac.Value
	}
	if !b.AdSelect.Null {
		bond.BondADSelect = b.AdSelect.Value
	}
	if !b.MiiMon.Null {
		bond.BondMIIMon = uint32(b.MiiMon.Value)
	}
	if !b.UpDelay.Null {
		bond.BondUpDelay = uint32(b.UpDelay.Value)
	}
	if !b.DownDelay.Null {
		bond.BondDownDelay = uint32(b.DownDelay.Value)
	}
	if !b.ArpInterval.Null {
		bond.BondARPInterval = uint32(b.ArpInterval.Value)
	}
	if !b.ResendIgmp.Null {
		bond.BondResendIGMP = uint32(b.ResendIgmp.Value)
	}
	if !b.MinLinks.Null {
		bond.BondMinLinks = uint32(b.MinLinks.Value)
	}
	if !b.LpInterval.Null {
		bond.BondLPInterval = uint32(b.LpInterval.Value)
	}
	if !b.PacketsPerSlave.Null {
		bond.BondPacketsPerSlave = uint32(b.PacketsPerSlave.Value)
	}
	if !b.NumPeerNotif.Null {
		bond.BondNumPeerNotif = uint8(b.NumPeerNotif.Value)
	}
	if !b.TlbDynamicLb.Null {
		bond.BondTLBDynamicLB = uint8(b.TlbDynamicLb.Value)
	}
	if !b.AllSlavesActive.Null {
		bond.BondAllSlavesActive = uint8(b.AllSlavesActive.Value)
	}
	if !b.UseCarrier.Null {
		// TODO: Fix, doesn't set properly.
		carrier := b.UseCarrier.Value
		bond.BondUseCarrier = &carrier
	}
	if !b.AdActorSysPrio.Null {
		bond.BondADActorSysPrio = uint16(b.AdActorSysPrio.Value)
	}
	if !b.AdUserPortKey.Null {
		bond.BondADUserPortKey = uint16(b.AdUserPortKey.Value)
	}
	if !b.PeerNotifyDelay.Null {
		bond.BondPeerNotifyDelay = uint32(b.PeerNotifyDelay.Value)
	}

	return bond, nil
}

func readBond(bond config.Bond) (out *BondData) {
	out = &BondData{}

	out.Interfaces = readStringList(bond.Interfaces())
	out.ARPIPTarget = readStringList(bond.ARPIPTarget())
	out.Mode = readString(bond.Mode())

	out.ArpValidate = readString(bond.ARPValidate())
	out.Primary = readString(bond.Primary())
	out.PrimaryReselect = readString(bond.PrimaryReselect())
	out.AdSelect = readString(bond.ADSelect())
	out.AdUserPortKey = readInt(int(bond.ADUserPortKey()))
	out.AllSlavesActive = readInt(int(bond.AllSlavesActive()))
	out.ArpAllTargets = readString(bond.ARPAllTargets())
	out.DownDelay = readInt(int(bond.DownDelay()))
	out.FailoverMac = readString(bond.FailOverMac())
	out.LacpRate = readString(bond.LACPRate())
	out.LpInterval = readInt(int(bond.LPInterval()))
	out.MiiMon = readInt(int(bond.MIIMon()))
	out.MinLinks = readInt(int(bond.MinLinks()))
	out.NumPeerNotif = readInt(int(bond.NumPeerNotif()))
	out.PeerNotifyDelay = readInt(int(bond.PeerNotifyDelay()))
	out.ResendIgmp = readInt(int(bond.ResendIGMP()))
	out.UpDelay = readInt(int(bond.UpDelay()))
	out.XmitHashPolicy = readString(bond.HashPolicy())

	out.AdActorSysPrio = readInt(int(bond.ADActorSysPrio()))

	return
}
