package utils

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	hubtypes "github.com/sentinel-official/hub/types"
)

func WriteKeys(w io.Writer, keys ...keyring.Info) error {
	tw := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)
	if _, err := fmt.Fprintf(
		tw, "%s\t%s\t%s\n",
		"Name", "Address", "Operator",
	); err != nil {
		return err
	}

	for i := 0; i < len(keys); i++ {
		var (
			name     = keys[i].GetName()
			address  = hubtypes.NodeAddress(keys[i].GetAddress().Bytes()).String()
			operator = keys[i].GetAddress().String()
		)

		if _, err := fmt.Fprintf(
			tw, "%s\t%s\t%s\n",
			name, address, operator,
		); err != nil {
			return err
		}
	}

	return tw.Flush()
}
