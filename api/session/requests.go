package session

import (
	"encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
)

type RequestAddSession struct {
	AccAddress sdk.AccAddress
	Key        []byte
	Signature  []byte

	URI struct {
		AccAddress string `uri:"acc_address"`
		ID         uint64 `uri:"id" binding:"gt=0"`
	}
	Body struct {
		Key       string `json:"key"`
		Signature string `json:"signature"`
	}
}

func NewRequestAddSession(c *gin.Context) (req *RequestAddSession, err error) {
	req = &RequestAddSession{}
	if err = c.ShouldBindUri(&req.URI); err != nil {
		return nil, err
	}
	if err = c.ShouldBindJSON(&req.Body); err != nil {
		return nil, err
	}

	req.AccAddress, err = sdk.AccAddressFromBech32(req.URI.AccAddress)
	if err != nil {
		return nil, err
	}
	req.Key, err = base64.StdEncoding.DecodeString(req.Body.Key)
	if err != nil {
		return nil, err
	}
	req.Signature, err = base64.StdEncoding.DecodeString(req.Body.Signature)
	if err != nil {
		return nil, err
	}

	return req, nil
}
