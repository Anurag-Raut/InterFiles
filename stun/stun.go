package stun

import (
	"fmt"

	"github.com/pion/stun"
)

func GetAddressFromStun() (string, error) {
	u, err := stun.ParseURI("stun:stun.l.google.com:19302")
	if err != nil {
		return "", fmt.Errorf("failed to parse URI: %v", err)
	}

	c, err := stun.DialURI(u, &stun.DialConfig{})
	if err != nil {
		return "", fmt.Errorf("failed to dial URI: %v", err)
	}

	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	var address string
	err = c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			err = res.Error
			return
		}

		var xorAddr stun.XORMappedAddress
		if newerr := xorAddr.GetFrom(res.Message); newerr != nil {
			err = fmt.Errorf("failed to get XOR mapped address: %v", err)
			return
		}

		address = xorAddr.IP.To16().String()
	})

	if err != nil {
		return "", err
	}

	return address, nil
}
